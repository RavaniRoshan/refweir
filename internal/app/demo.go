package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"refweir.dev/internal/audit"
	"refweir.dev/internal/store"
	"refweir.dev/pkg/agentlease"
	"refweir.dev/pkg/idgen"
)

const (
	demoRunMinInterval = 20 * time.Second
	demoRunsPerHour    = 5
	demoLeaseTTL       = 10 * time.Minute
)

// demoRateLimiter is a small per-client limiter for the public demo run. It is
// intentionally in-memory: the demo is a single process and resets on restart.
type demoRateLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func newDemoRateLimiter() *demoRateLimiter {
	return &demoRateLimiter{hits: make(map[string][]time.Time)}
}

// allow reports whether the key may run now and, if not, how long to wait.
func (l *demoRateLimiter) allow(key string, now time.Time) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := now.Add(-time.Hour)
	kept := l.hits[key][:0]
	for _, hit := range l.hits[key] {
		if hit.After(cutoff) {
			kept = append(kept, hit)
		}
	}

	if n := len(kept); n > 0 {
		if since := now.Sub(kept[n-1]); since < demoRunMinInterval {
			l.hits[key] = kept
			return false, demoRunMinInterval - since
		}
	}
	if len(kept) >= demoRunsPerHour {
		l.hits[key] = kept
		return false, kept[0].Add(time.Hour).Sub(now)
	}

	l.hits[key] = append(kept, now)
	return true, 0
}

func (a *App) registerDemoRoutes(mux *http.ServeMux) {
	if !a.Config.DemoMode {
		return
	}
	if a.demoLimiter == nil {
		a.demoLimiter = newDemoRateLimiter()
	}
	if a.demoSlots == nil {
		a.demoSlots = make(chan struct{}, 1)
	}

	mux.HandleFunc("POST /api/v1/demo/agent-run", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")

		ok, retry := a.demoLimiter.allow(clientKey(r), time.Now())
		if !ok {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(retry.Seconds())+1))
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":      false,
				"error":   "The demo allows one agent run every 20 seconds (five per hour).",
				"retryIn": retry.Round(time.Second).String(),
			})
			return
		}

		select {
		case a.demoSlots <- struct{}{}:
			defer func() { <-a.demoSlots }()
		default:
			w.Header().Set("Retry-After", "5")
			w.WriteHeader(http.StatusTooManyRequests)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"ok":    false,
				"error": "Another demo agent run is in progress. Try again in a few seconds.",
			})
			return
		}

		result, err := a.runDemoAgent(r.Context())
		if err != nil {
			_ = audit.Log(audit.Event{
				Actor:     "demo-agent",
				ActorType: "agent",
				Action:    "demo.agent_run_failed",
				Resource:  a.Config.DemoRepo,
				ClientIP:  r.RemoteAddr,
				Severity:  audit.SeverityWarning,
				Details:   err.Error(),
			})
			status := http.StatusInternalServerError
			if errors.Is(err, errDemoRepoUnavailable) {
				status = http.StatusServiceUnavailable
			}
			w.WriteHeader(status)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(result)
	})
}

var errDemoRepoUnavailable = errors.New("demo repository is not seeded yet; run deploy/demo/seed.sh")

type demoAgentRunResult struct {
	OK        bool     `json:"ok"`
	Agent     string   `json:"agent"`
	LeaseID   string   `json:"leaseId"`
	ExpiresAt string   `json:"expiresAt"`
	Branch    string   `json:"branch"`
	Title     string   `json:"title"`
	PRNumber  int      `json:"prNumber"`
	PRURL     string   `json:"prUrl"`
	Files     []string `json:"files"`
	AuditTail []any    `json:"auditTail"`
}

// runDemoAgent performs a real, bounded agent loop against the demo repository:
// issue a scoped lease, clone over Smart-HTTP with that lease, commit on an
// agent/* branch, push, and open a pull request. It never writes to main.
func (a *App) runDemoAgent(ctx context.Context) (*demoAgentRunResult, error) {
	owner, name, ok := strings.Cut(a.Config.DemoRepo, "/")
	if !ok || owner == "" || name == "" {
		return nil, fmt.Errorf("REFWEIR_DEMO_REPO must be owner/name, got %q", a.Config.DemoRepo)
	}
	repo, err := a.Store.GetRepo(ctx, owner, name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, errDemoRepoUnavailable
		}
		return nil, fmt.Errorf("load demo repository: %w", err)
	}

	agentID := "demo-agent"
	allowed := []string{"refs/heads/agent/"}
	disallowed := []string{"refs/heads/main", "refs/heads/master"}
	token, lease, err := agentlease.Issue(a.LeaseSecret, agentID, repo.ID, allowed, disallowed, demoLeaseTTL)
	if err != nil {
		return nil, fmt.Errorf("issue demo lease: %w", err)
	}
	_ = audit.Log(audit.Event{
		Actor:     "agent:" + agentID,
		ActorType: "agent",
		Action:    "agent.lease_issued",
		Resource:  repo.ID,
		Severity:  audit.SeverityInfo,
		Details:   "Capability lease issued for the hosted demo run",
		Metadata: map[string]interface{}{
			"leaseId":               lease.ID,
			"agentId":               lease.AgentID,
			"allowedRefPrefixes":    lease.AllowedRefPrefixes,
			"disallowedRefPrefixes": lease.DisallowedRefPrefixes,
			"leaseExpiresAt":        lease.ExpiresAt.Format(time.RFC3339),
			"demo":                  true,
		},
	})

	stamp := time.Now().UTC().Format("20060102-150405")
	branch := "agent/demo-" + stamp
	file := filepath.ToSlash(filepath.Join("notes", "agent-runs", stamp+".md"))
	content := fmt.Sprintf(`# Agent run %s

This file was written by the hosted Refweir demo agent.

- Agent: %s
- Lease: %s
- Scope: refs/heads/agent/*
- Expires: %s
- Base branch: %s

The push was authenticated with the lease, not with a long-lived token. The
lease cannot write to %s. Every step is in the operator's audit log.

This is a real commit on a real branch in this demo instance. It is not mock
data, and it does not touch %s.
`, stamp, agentID, lease.ID, lease.ExpiresAt.Format(time.RFC3339), repo.DefaultBranch, repo.DefaultBranch, repo.DefaultBranch)

	cloneURL, err := a.demoCloneURL(repo.FullName, token)
	if err != nil {
		return nil, err
	}
	workDir, err := os.MkdirTemp("", "refweir-demo-run-*")
	if err != nil {
		return nil, fmt.Errorf("create demo workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	if out, err := runGit(workDir, "clone", "--quiet", cloneURL, "repo"); err != nil {
		return nil, fmt.Errorf("clone demo repository: %s", sanitizeGitOutput(out, token))
	}
	repoDir := filepath.Join(workDir, "repo")

	if out, err := runGit(repoDir, "checkout", "--quiet", "-b", branch); err != nil {
		return nil, fmt.Errorf("create agent branch: %s", sanitizeGitOutput(out, token))
	}
	fullPath := filepath.Join(repoDir, filepath.FromSlash(file))
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return nil, fmt.Errorf("create agent file directory: %w", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("write agent file: %w", err)
	}
	if out, err := runGit(repoDir, "add", file); err != nil {
		return nil, fmt.Errorf("stage agent file: %s", sanitizeGitOutput(out, token))
	}
	if out, err := runGit(repoDir, "commit", "--quiet", "-m", "docs: add agent run note"); err != nil {
		return nil, fmt.Errorf("commit agent change: %s", sanitizeGitOutput(out, token))
	}
	if out, err := runGit(repoDir, "push", "--quiet", "origin", "HEAD:refs/heads/"+branch); err != nil {
		return nil, fmt.Errorf("push agent branch: %s", sanitizeGitOutput(out, token))
	}

	headSHA, err := a.Git.ResolveRef(owner, name, branch)
	if err != nil {
		return nil, fmt.Errorf("resolve agent branch: %w", err)
	}
	baseSHA, err := a.Git.ResolveRef(owner, name, repo.DefaultBranch)
	if err != nil {
		return nil, fmt.Errorf("resolve base branch: %w", err)
	}

	pr := &store.PullRequest{
		ID:        idgen.NewPR(),
		RepoID:    repo.ID,
		Title:     "docs: add an agent-authored run note",
		Body:      "This pull request was opened by the hosted demo agent after a scoped lease push.\n\nThe lease allows `refs/heads/agent/*` and denies `main`/`master`. The audit log records the lease, the accepted push, and the PR.",
		State:     "open",
		HeadRef:   branch,
		BaseRef:   repo.DefaultBranch,
		HeadSHA:   headSHA.String(),
		BaseSHA:   baseSHA.String(),
		Mergeable: "mergeable",
	}
	if err := a.Store.CreatePullRequest(ctx, pr); err != nil {
		return nil, fmt.Errorf("open demo pull request: %w", err)
	}

	_ = audit.Log(audit.Event{
		Actor:     "agent:" + agentID,
		ActorType: "agent",
		Action:    "demo.agent_run",
		Resource:  repo.FullName + "#" + fmt.Sprintf("%d", pr.Number),
		Severity:  audit.SeverityInfo,
		Details:   "Hosted demo agent pushed an agent/* branch and opened a pull request",
		Metadata: map[string]interface{}{
			"leaseId":  lease.ID,
			"branch":   branch,
			"prNumber": pr.Number,
			"demo":     true,
		},
	})

	return &demoAgentRunResult{
		OK:        true,
		Agent:     agentID,
		LeaseID:   lease.ID,
		ExpiresAt: lease.ExpiresAt.Format(time.RFC3339),
		Branch:    branch,
		Title:     pr.Title,
		PRNumber:  pr.Number,
		PRURL:     "/?tab=prs",
		Files:     []string{file},
		AuditTail: demoAuditTail(),
	}, nil
}

func (a *App) demoCloneURL(fullName, token string) (string, error) {
	base := a.Config.DemoSelfURL
	if base == "" {
		host, port, err := net.SplitHostPort(a.Config.ListenAddr)
		if err != nil {
			return "", fmt.Errorf("derive demo self URL from %q: %w", a.Config.ListenAddr, err)
		}
		if host == "" || host == "0.0.0.0" || host == "::" {
			host = "127.0.0.1"
		}
		base = "http://" + net.JoinHostPort(host, port)
	}
	u, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return "", fmt.Errorf("parse demo self URL: %w", err)
	}
	u.User = url.UserPassword("agent", token)
	u.Path = "/" + fullName + ".git"
	return u.String(), nil
}

func runGit(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...) // #nosec G204 -- fixed git subcommands with validated repo paths
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_AUTHOR_NAME=Refweir Demo Agent",
		"GIT_AUTHOR_EMAIL=demo-agent@refweir.dev",
		"GIT_COMMITTER_NAME=Refweir Demo Agent",
		"GIT_COMMITTER_EMAIL=demo-agent@refweir.dev",
	)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func sanitizeGitOutput(out, token string) string {
	clean := strings.TrimSpace(out)
	if token != "" {
		clean = strings.ReplaceAll(clean, token, "<lease>")
	}
	if clean == "" {
		return "git command failed"
	}
	return clean
}

func clientKey(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if first := strings.TrimSpace(strings.Split(forwarded, ",")[0]); first != "" {
			return first
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func demoAuditTail() []any {
	events := audit.Recent(6)
	out := make([]any, 0, len(events))
	for _, ev := range events {
		out = append(out, map[string]any{
			"timestamp": ev.Timestamp.Format(time.RFC3339),
			"actor":     ev.Actor,
			"actorType": ev.ActorType,
			"action":    ev.Action,
			"resource":  ev.Resource,
			"severity":  ev.Severity,
		})
	}
	return out
}
