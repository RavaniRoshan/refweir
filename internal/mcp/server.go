// Package mcp implements the Model Context Protocol (MCP) JSON-RPC 2.0 server
// and tool definitions, providing AI coding agents with structured access to refweir.dev.
package mcp

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"refweir.dev/internal/agentactivity"
	"refweir.dev/internal/audit"
	"refweir.dev/internal/auth"
	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/metrics"
	"refweir.dev/internal/store"
)

// Server implements the Model Context Protocol (MCP) HTTP JSON-RPC 2.0 server.
type Server struct {
	store       *store.DB
	git         *gitstore.Store
	allTools    []ToolDef
	toolMap     map[string]ToolDef
	leaseSecret []byte
	leasePubKey ed25519.PublicKey
	mergeHook   func(context.Context, store.PullRequest)
}

// SetMergeHook registers a callback invoked after a successful pull-request merge.
func (s *Server) SetMergeHook(fn func(context.Context, store.PullRequest)) {
	s.mergeHook = fn
}

// New creates a new MCP Server.
func New(db *store.DB, git *gitstore.Store) *Server {
	tools := BuiltinTools()
	tmap := make(map[string]ToolDef, len(tools))
	for _, t := range tools {
		tmap[t.Name] = t
	}
	return &Server{
		store:    db,
		git:      git,
		allTools: tools,
		toolMap:  tmap,
	}
}

// SetLeaseSecret sets the HMAC signing secret for capability leases.
func (s *Server) SetLeaseSecret(secret []byte) {
	s.leaseSecret = secret
}

// LeaseSecret returns the configured HMAC signing secret for capability leases.
func (s *Server) LeaseSecret() []byte {
	return s.leaseSecret
}

// SetLeasePublicKey sets the Ed25519 public key used to verify asymmetric capability leases.
func (s *Server) SetLeasePublicKey(pub ed25519.PublicKey) {
	s.leasePubKey = pub
}

// LeasePublicKey returns the configured Ed25519 public key for capability leases.
func (s *Server) LeasePublicKey() ed25519.PublicKey {
	return s.leasePubKey
}

// JSON-RPC request and response envelopes
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id,omitempty"`
	Result  any       `json:"result,omitempty"`
	Error   *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type toolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type toolCallContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type toolCallResult struct {
	Content []toolCallContent `json:"content"`
	IsError bool              `json:"isError"`
}

// Handler returns an HTTP handler for the MCP endpoint.
// When readonly is true, only read-only tools are listed and allowed.
func (s *Server) Handler(readonly bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed: MCP expects POST", http.StatusMethodNotAllowed)
			return
		}

		var req rpcRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeRPCError(w, nil, -32700, "Parse error: "+err.Error())
			return
		}

		ctx := r.Context()

		var callerScopes string
		// Enforce authentication when users are registered on the forge
		count, _ := s.store.CountUsers(ctx)
		if count > 0 {
			if _, err := auth.ResolveUser(ctx, s.store, r); err != nil {
				writeRPCError(w, req.ID, -32001, "Unauthorized: Authentication required via Bearer token or session")
				return
			}
		}

		// Resolve caller token scopes if Bearer authorization is provided
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			rawToken := strings.TrimPrefix(authHeader, "Bearer ")
			tokHash := auth.HashCredential(rawToken)
			if key, err := s.store.GetAPIKeyByHash(ctx, tokHash); err == nil && key.Scopes != "" {
				callerScopes = key.Scopes
			} else if tokRecord, err := s.store.GetAccessTokenByHash(ctx, tokHash); err == nil && tokRecord.Scopes != "" {
				callerScopes = tokRecord.Scopes
			}
		}

		effectiveReadOnly := readonly || isScopeReadOnly(callerScopes)

		switch req.Method {
		case "initialize":
			writeRPCResult(w, req.ID, map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "refweir.dev",
					"version": "1.0.0",
				},
			})

		case "notifications/initialized":
			// MCP initialized notification; no response needed for notifications
			w.WriteHeader(http.StatusNoContent)

		case "ping":
			writeRPCResult(w, req.ID, map[string]any{})

		case "tools/list":
			var exposed []ToolDef
			for _, t := range s.allTools {
				if effectiveReadOnly && !t.ReadOnly {
					continue
				}
				exposed = append(exposed, t)
			}
			writeRPCResult(w, req.ID, map[string]any{
				"tools": exposed,
			})

		case "tools/call":
			var params toolCallParams
			if len(req.Params) > 0 {
				if err := json.Unmarshal(req.Params, &params); err != nil {
					writeRPCError(w, req.ID, -32602, "Invalid params: "+err.Error())
					return
				}
			}

			tool, ok := s.toolMap[params.Name]
			if !ok {
				writeRPCResult(w, req.ID, toolCallResult{
					Content: []toolCallContent{{Type: "text", Text: fmt.Sprintf("Unknown tool %q", params.Name)}},
					IsError: true,
				})
				return
			}

			if effectiveReadOnly && !tool.ReadOnly {
				msg := "Tool %q is a write operation not permitted on /mcp/readonly"
				if isScopeReadOnly(callerScopes) {
					msg = fmt.Sprintf("Tool %%q is a write operation not permitted for token with read-only scopes (%s)", callerScopes)
				}
				writeRPCResult(w, req.ID, toolCallResult{
					Content: []toolCallContent{{Type: "text", Text: fmt.Sprintf(msg, params.Name)}},
					IsError: true,
				})
				return
			}

			if params.Arguments == nil {
				params.Arguments = make(map[string]any)
			}

			metrics.DefaultRegistry.IncMCPCalls()
			start := time.Now()

			res, err := tool.Handler(ctx, s, params.Arguments)
			durationMs := int(time.Since(start).Milliseconds())

			status := "success"
			if err != nil {
				status = "failed"
			}

			repoName := ""
			if val, ok := params.Arguments["repo_name"]; ok {
				repoName = fmt.Sprintf("%v", val)
			} else if val, ok := params.Arguments["repoName"]; ok {
				repoName = fmt.Sprintf("%v", val)
			}

			agentactivity.DefaultHub.Record(agentactivity.Activity{
				AgentName:  "MCP Agent",
				AgentType:  "refweir-mcp",
				Action:     "tool_call",
				ToolName:   params.Name,
				Resource:   repoName,
				Status:     status,
				DurationMs: durationMs,
				Details:    fmt.Sprintf("Invoked tool %q", params.Name),
			})

			if err != nil {
				writeRPCResult(w, req.ID, toolCallResult{
					Content: []toolCallContent{{Type: "text", Text: err.Error()}},
					IsError: true,
				})
				return
			}

			if !tool.ReadOnly {
				_ = audit.Log(audit.Event{
					Actor:     "agent-mcp",
					ActorType: "agent",
					Action:    "mcp.tool_execute:" + params.Name,
					Resource:  repoName,
					Severity:  audit.SeverityInfo,
					Details:   fmt.Sprintf("Autonomous agent executed mutating tool %q", params.Name),
				})
			}

			resBytes, _ := json.MarshalIndent(res, "", "  ")
			writeRPCResult(w, req.ID, toolCallResult{
				Content: []toolCallContent{{Type: "text", Text: string(resBytes)}},
				IsError: false,
			})

		default:
			writeRPCError(w, req.ID, -32601, fmt.Sprintf("Method %q not found", req.Method))
		}
	})
}

func writeRPCResult(w http.ResponseWriter, id any, result any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func writeRPCError(w http.ResponseWriter, id any, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &rpcError{
			Code:    code,
			Message: message,
		},
	})
}

// isScopeReadOnly returns true if the token's granted scopes only permit read operations,
// or specifically lack mutating write permissions.
func isScopeReadOnly(scopes string) bool {
	if scopes == "" {
		return false
	}
	tokens := strings.Fields(strings.ReplaceAll(scopes, ",", " "))
	hasWrite := false
	hasRead := false
	for _, tok := range tokens {
		tok = strings.TrimSpace(tok)
		if tok == "*" || tok == "repo" || tok == "write" || tok == "repo:write" || tok == "admin" || tok == "admin:org" || tok == "admin:repo" {
			hasWrite = true
		}
		if tok == "read" || tok == "read:repo" || tok == "repo:read" || tok == "read:all" || tok == "mcp:read" {
			hasRead = true
		}
	}
	return hasRead && !hasWrite
}
