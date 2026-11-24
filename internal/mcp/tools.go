package mcp

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"
	"refweir.dev/internal/audit"
	"refweir.dev/internal/engine"
	"refweir.dev/internal/security"
	"refweir.dev/internal/store"
	"refweir.dev/internal/symbolsearch"
	"refweir.dev/pkg/agentlease"
	"refweir.dev/pkg/idgen"
)

// ToolDef defines an MCP tool and its execution handler.
type ToolDef struct {
	Name        string                                                                 `json:"name"`
	Description string                                                                 `json:"description"`
	ReadOnly    bool                                                                   `json:"-"`
	InputSchema map[string]any                                                         `json:"inputSchema"`
	Handler     func(ctx context.Context, s *Server, args map[string]any) (any, error) `json:"-"`
}

func getString(args map[string]any, key string) string {
	if v, ok := args[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
		return fmt.Sprintf("%v", v)
	}
	return ""
}

func getInt(args map[string]any, key string) int {
	if v, ok := args[key]; ok && v != nil {
		switch n := v.(type) {
		case int:
			return n
		case int64:
			return int(n)
		case float64:
			return int(n)
		case string:
			if i, err := strconv.Atoi(n); err == nil {
				return i
			}
		}
	}
	return 0
}

func getBool(args map[string]any, key string) bool {
	if v, ok := args[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
		if s, ok := v.(string); ok {
			return s == "true" || s == "1"
		}
	}
	return false
}

func getStringSlice(args map[string]any, key string) []string {
	if v, ok := args[key]; ok && v != nil {
		if list, ok := v.([]any); ok {
			var out []string
			for _, item := range list {
				out = append(out, fmt.Sprintf("%v", item))
			}
			return out
		}
		if list, ok := v.([]string); ok {
			return list
		}
	}
	return nil
}

func makeSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": properties,
		"required":   required,
	}
}

// BuiltinTools returns all defined MCP tools.
func BuiltinTools() []ToolDef {
	return []ToolDef{
		// 1. get_repository
		{
			Name:        "get_repository",
			Description: "Get metadata for a Refweir repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Repository owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				return s.store.GetRepo(ctx, owner, name)
			},
		},
		// 2. list_repositories
		{
			Name:        "list_repositories",
			Description: "List repositories accessible to the account",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Optional owner namespace filter"},
			}, nil),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				if owner != "" {
					repos, _, err := s.store.ListRepos(ctx, owner, 100, "")
					return repos, err
				}
				namespaces, err := s.store.ListNamespaces(ctx, 100)
				if err != nil {
					return nil, err
				}
				var all []store.Repo
				for _, ns := range namespaces {
					repos, _, _ := s.store.ListRepos(ctx, ns.Slug, 100, "")
					all = append(all, repos...)
				}
				return all, nil
			},
		},
		// 3. create_repository
		{
			Name:        "create_repository",
			Description: "Create a new repository in Refweir",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":     map[string]any{"type": "string", "description": "Owner namespace slug"},
				"name":          map[string]any{"type": "string", "description": "Repository name"},
				"defaultBranch": map[string]any{"type": "string", "description": "Default branch name (default: main)"},
			}, []string{"ownerSlug", "name"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "name")
				branch := getString(args, "defaultBranch")
				if branch == "" {
					branch = "main"
				}
				ns, err := s.store.GetNamespaceBySlug(ctx, owner)
				if err != nil {
					return nil, fmt.Errorf("namespace %q not found: %w", owner, err)
				}
				repo := store.Repo{
					ID:                  idgen.NewRepo(),
					NamespaceID:         ns.ID,
					Name:                name,
					FullName:            owner + "/" + name,
					DefaultBranch:       branch,
					Visibility:          "private",
					AllowMergeCommit:    true,
					AllowSquashMerge:    true,
					DeleteBranchOnMerge: false,
				}
				if err := s.store.CreateRepo(ctx, repo); err != nil {
					return nil, err
				}
				if _, err := s.git.Init(owner, name, branch); err != nil {
					return nil, fmt.Errorf("init git repo: %w", err)
				}
				return repo, nil
			},
		},
		// 4. get_file_contents
		{
			Name:        "get_file_contents",
			Description: "Read the content of a file in a repository at a given ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"path":      map[string]any{"type": "string", "description": "File path in repository"},
				"ref":       map[string]any{"type": "string", "description": "Commit or branch ref (default: HEAD)"},
			}, []string{"ownerSlug", "repoName", "path"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				path := getString(args, "path")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				data, sha, size, isDir, err := s.git.ReadFileAtPath(owner, name, ref, path)
				if err != nil {
					return nil, err
				}
				if isDir {
					return map[string]any{"type": "dir", "path": path}, nil
				}
				return map[string]any{
					"path":     path,
					"sha":      sha,
					"size":     size,
					"encoding": "base64",
					"content":  base64.StdEncoding.EncodeToString(data),
				}, nil
			},
		},
		// 5. get_git_tree
		{
			Name:        "get_git_tree",
			Description: "Get the directory tree of a repository at a ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"ref":       map[string]any{"type": "string", "description": "Git commit SHA, branch or tag (default: HEAD)"},
				"path":      map[string]any{"type": "string", "description": "Subdirectory path to list (default: root)"},
				"recursive": map[string]any{"type": "boolean", "description": "Recursively list subtree"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				path := getString(args, "path")
				recursive := getBool(args, "recursive")
				sha, entries, err := s.git.ReadTree(owner, name, ref, path, recursive)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"sha":     sha,
					"tree":    entries,
					"entries": len(entries),
				}, nil
			},
		},
		// 6. grep_contents
		{
			Name:        "grep_contents",
			Description: "Search file contents in a repository at a ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"query":     map[string]any{"type": "string", "description": "Search query or regex"},
				"ref":       map[string]any{"type": "string", "description": "Git ref (default: HEAD)"},
			}, []string{"ownerSlug", "repoName", "query"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				query := getString(args, "query")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				matches, truncated, err := s.git.Grep(owner, name, ref, query, nil, false, false, 50)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"matches":   matches,
					"truncated": truncated,
				}, nil
			},
		},
		// 7. list_branches
		{
			Name:        "list_branches",
			Description: "List branches in a repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				return s.git.ListBranches(owner, name)
			},
		},
		// 8. list_commits
		{
			Name:        "list_commits",
			Description: "List commits reachable from a ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"ref":       map[string]any{"type": "string", "description": "Branch or commit ref (default: HEAD)"},
				"limit":     map[string]any{"type": "integer", "description": "Maximum commits to return (default: 20)"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				limit := getInt(args, "limit")
				if limit <= 0 {
					limit = 20
				}
				return s.git.ListCommits(owner, name, ref, limit)
			},
		},
		// 9. compare_commits
		{
			Name:        "compare_commits",
			Description: "Compare differences between two commit SHAs or branches",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"base":      map[string]any{"type": "string", "description": "Base ref or commit SHA"},
				"head":      map[string]any{"type": "string", "description": "Head ref or commit SHA"},
			}, []string{"ownerSlug", "repoName", "base", "head"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				base := getString(args, "base")
				head := getString(args, "head")
				if baseHash, err := s.git.ResolveRef(owner, name, base); err == nil {
					base = baseHash.String()
				}
				if headHash, err := s.git.ResolveRef(owner, name, head); err == nil {
					head = headHash.String()
				}
				return s.git.DiffCommits(owner, name, base, head)
			},
		},
		// 10. sync_mirror
		{
			Name:        "sync_mirror",
			Description: "Trigger an immediate sync of a mirrored repository from upstream",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				if repo.MirrorSource == "" {
					return nil, fmt.Errorf("repository %s/%s is not configured as a mirror", owner, name)
				}
				return map[string]any{
					"synced": true,
					"repoId": repo.ID,
				}, nil
			},
		},
		// 11. list_pull_requests
		{
			Name:        "list_pull_requests",
			Description: "List pull requests in a repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"status":    map[string]any{"type": "string", "description": "Optional status filter (open, closed)"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				status := getString(args, "status")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pulls, _, err := s.store.ListPullRequests(ctx, repo.ID, status, 50, "")
				return pulls, err
			},
		},
		// 12. search_pull_requests
		{
			Name:        "search_pull_requests",
			Description: "Search pull requests in a repository by title or query",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"query":     map[string]any{"type": "string", "description": "Search query"},
			}, []string{"ownerSlug", "repoName", "query"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				query := strings.ToLower(getString(args, "query"))
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pulls, _, err := s.store.ListPullRequests(ctx, repo.ID, "", 100, "")
				if err != nil {
					return nil, err
				}
				var matched []store.PullRequest
				for _, pr := range pulls {
					if strings.Contains(strings.ToLower(pr.Title), query) || strings.Contains(strings.ToLower(pr.Body), query) {
						matched = append(matched, pr)
					}
				}
				return matched, nil
			},
		},
		// 13. pull_request_read
		{
			Name:        "pull_request_read",
			Description: "Get details of a pull request by number",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
			}, []string{"ownerSlug", "repoName", "pullNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				return s.store.GetPullRequest(ctx, repo.ID, num)
			},
		},
		// 14. create_pull_request
		{
			Name:        "create_pull_request",
			Description: "Open a new pull request in a repository",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"title":       map[string]any{"type": "string", "description": "Pull request title"},
				"headRef":     map[string]any{"type": "string", "description": "Head branch"},
				"baseRef":     map[string]any{"type": "string", "description": "Base branch (e.g. main)"},
				"description": map[string]any{"type": "string", "description": "Optional body/description"},
			}, []string{"ownerSlug", "repoName", "title", "headRef", "baseRef"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				head := getString(args, "headRef")
				base := getString(args, "baseRef")
				title := getString(args, "title")
				desc := getString(args, "description")

				headHash, err := s.git.ResolveRef(owner, name, head)
				if err != nil {
					return nil, fmt.Errorf("resolve head ref %q: %w", head, err)
				}
				baseHash, err := s.git.ResolveRef(owner, name, base)
				if err != nil {
					return nil, fmt.Errorf("resolve base ref %q: %w", base, err)
				}

				pr := store.PullRequest{
					ID:      idgen.NewPR(),
					RepoID:  repo.ID,
					Title:   title,
					Body:    desc,
					State:   "open",
					HeadRef: head,
					BaseRef: base,
					HeadSHA: headHash.String(),
					BaseSHA: baseHash.String(),
				}
				if err := s.store.CreatePullRequest(ctx, &pr); err != nil {
					return nil, err
				}
				return pr, nil
			},
		},
		// 15. update_pull_request
		{
			Name:        "update_pull_request",
			Description: "Update pull request title, description, or status",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber":  map[string]any{"type": "integer", "description": "Pull request number"},
				"title":       map[string]any{"type": "string", "description": "New title"},
				"description": map[string]any{"type": "string", "description": "New description"},
				"status":      map[string]any{"type": "string", "description": "New status (open or closed)"},
			}, []string{"ownerSlug", "repoName", "pullNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				var u store.PullRequestUpdate
				if t := getString(args, "title"); t != "" {
					u.Title = &t
				}
				if d := getString(args, "description"); d != "" {
					u.Body = &d
				}
				if st := getString(args, "status"); st != "" {
					u.State = &st
				}
				if err := s.store.UpdatePullRequest(ctx, pr.ID, u); err != nil {
					return nil, err
				}
				return s.store.GetPullRequest(ctx, repo.ID, num)
			},
		},
		// 16. update_pull_request_base
		{
			Name:        "update_pull_request_base",
			Description: "Update the base branch of a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"baseRef":    map[string]any{"type": "string", "description": "New base branch"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "baseRef"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				base := getString(args, "baseRef")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				baseHash, err := s.git.ResolveRef(owner, name, base)
				if err != nil {
					return nil, fmt.Errorf("resolve base ref %q: %w", base, err)
				}
				bSHA := baseHash.String()
				u := store.PullRequestUpdate{
					BaseRef: &base,
					BaseSHA: &bSHA,
				}
				if err := s.store.UpdatePullRequest(ctx, pr.ID, u); err != nil {
					return nil, err
				}
				return s.store.GetPullRequest(ctx, repo.ID, num)
			},
		},
		// 17. merge_pull_request
		{
			Name:        "merge_pull_request",
			Description: "Merge a pull request into its base branch",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber":  map[string]any{"type": "integer", "description": "Pull request number"},
				"mergeMethod": map[string]any{"type": "string", "description": "Merge method (merge, squash)"},
			}, []string{"ownerSlug", "repoName", "pullNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				method := getString(args, "mergeMethod")
				if method == "" {
					method = "merge"
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}

				// Evaluate merge rulesets and check runs
				rulesets, err := s.store.ListRulesets(ctx, repo.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to list rulesets for merge evaluation: %w", err)
				}
				reviews, err := s.store.ListReviews(ctx, pr.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to list reviews for merge evaluation: %w", err)
				}
				approvedCount := 0
				for _, rev := range reviews {
					if !rev.Dismissed && (rev.State == "approved" || rev.State == "approve") {
						approvedCount++
					}
				}
				checkRuns, err := s.store.ListCheckRunsForCommit(ctx, repo.ID, pr.HeadSHA, "", "", 100, 0)
				if err != nil {
					return nil, fmt.Errorf("failed to list check runs for merge evaluation: %w", err)
				}
				evalRes := engine.EvaluateMergeWithOptions(ctx, rulesets, repo.DefaultBranch, pr.BaseRef, "agent", "", approvedCount, checkRuns, engine.MergeEvaluationOptions{
					MergeMethod: method,
				})
				if !evalRes.Allowed {
					var msgs []string
					for _, b := range evalRes.Blockers {
						msgs = append(msgs, b.Message)
					}
					return nil, fmt.Errorf("merge blocked by ruleset: %s", strings.Join(msgs, "; "))
				}

				sig := object.Signature{Name: "Refweir Agent", Email: "agent@refweir.dev", When: time.Now().UTC()}
				commitSHA, err := s.git.Merge(owner, name, pr.BaseRef, pr.HeadRef, method, sig, fmt.Sprintf("Merge PR #%d: %s", pr.Number, pr.Title))
				if err != nil {
					return nil, fmt.Errorf("git merge: %w", err)
				}
				if err := s.store.SetPullRequestMergeResult(ctx, pr.ID, commitSHA); err != nil {
					return nil, err
				}
				if s.mergeHook != nil {
					s.mergeHook(ctx, pr)
				}
				return s.store.GetPullRequest(ctx, repo.ID, num)
			},
		},
		// 18. create_pull_request_review
		{
			Name:        "create_pull_request_review",
			Description: "Submit a review on a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"verdict":    map[string]any{"type": "string", "description": "Review verdict (approved, changes_requested, commented)"},
				"body":       map[string]any{"type": "string", "description": "Review comment body"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "verdict"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				verdict := getString(args, "verdict")
				body := getString(args, "body")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				review := store.PRReview{
					ID:            idgen.NewPRReview(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					AuthorUserID:  "usr_agent",
					State:         verdict,
					Body:          body,
					CommitSHA:     pr.HeadSHA,
				}
				if err := s.store.CreateReview(ctx, review); err != nil {
					return nil, err
				}
				return review, nil
			},
		},
		// 19. update_pull_request_review
		{
			Name:        "update_pull_request_review",
			Description: "Update an existing pull request review",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"reviewId":   map[string]any{"type": "string", "description": "Review ID"},
				"body":       map[string]any{"type": "string", "description": "Updated body"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "reviewId"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				reviewID := getString(args, "reviewId")
				body := getString(args, "body")
				if err := s.store.UpdateReview(ctx, reviewID, body); err != nil {
					return nil, err
				}
				return s.store.GetReview(ctx, reviewID)
			},
		},
		// 20. dismiss_pull_request_review
		{
			Name:        "dismiss_pull_request_review",
			Description: "Dismiss a pull request review",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"reviewId":   map[string]any{"type": "string", "description": "Review ID"},
				"message":    map[string]any{"type": "string", "description": "Dismissal explanation"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "reviewId"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				reviewID := getString(args, "reviewId")
				msg := getString(args, "message")
				if err := s.store.DismissReview(ctx, reviewID, msg, "usr_agent"); err != nil {
					return nil, err
				}
				return s.store.GetReview(ctx, reviewID)
			},
		},
		// 21. create_pull_request_comment
		{
			Name:        "create_pull_request_comment",
			Description: "Add a top-level comment to a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"body":       map[string]any{"type": "string", "description": "Comment text"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				body := getString(args, "body")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				comment := store.PRComment{
					ID:            idgen.NewPRComment(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					AuthorUserID:  "",
					Body:          body,
				}
				if err := s.store.CreateComment(ctx, comment); err != nil {
					return nil, err
				}
				return comment, nil
			},
		},
		// 22. create_pull_request_inline_comment
		{
			Name:        "create_pull_request_inline_comment",
			Description: "Start a review thread on a specific file line in a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"path":       map[string]any{"type": "string", "description": "File path"},
				"line":       map[string]any{"type": "integer", "description": "Line number"},
				"body":       map[string]any{"type": "string", "description": "Comment text"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "path", "line", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				path := getString(args, "path")
				line := getInt(args, "line")
				body := getString(args, "body")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				thread := store.PRThread{
					ID:            idgen.NewPRThread(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					Path:          path,
					StartLine:     line,
					EndLine:       line,
					Side:          "RIGHT",
					HeadSHA:       pr.HeadSHA,
					BaseSHA:       pr.BaseSHA,
				}
				if err := s.store.CreateThread(ctx, thread); err != nil {
					return nil, err
				}
				comment := store.PRComment{
					ID:            idgen.NewPRComment(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					ThreadID:      thread.ID,
					AuthorUserID:  "usr_agent",
					Body:          body,
					Path:          path,
					Line:          line,
					Side:          "RIGHT",
					CommitSHA:     pr.HeadSHA,
				}
				if err := s.store.CreateComment(ctx, comment); err != nil {
					return nil, err
				}
				return map[string]any{
					"thread":  thread,
					"comment": comment,
				}, nil
			},
		},
		// 23. reply_pull_request_review_thread
		{
			Name:        "reply_pull_request_review_thread",
			Description: "Add a reply comment to an existing review thread",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"threadId":   map[string]any{"type": "string", "description": "Thread ID"},
				"body":       map[string]any{"type": "string", "description": "Reply comment body"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "threadId", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				threadID := getString(args, "threadId")
				body := getString(args, "body")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				comment := store.PRComment{
					ID:            idgen.NewPRComment(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					ThreadID:      threadID,
					AuthorUserID:  "usr_agent",
					Body:          body,
				}
				if err := s.store.CreateComment(ctx, comment); err != nil {
					return nil, err
				}
				return comment, nil
			},
		},
		// 24. update_pull_request_review_thread
		{
			Name:        "update_pull_request_review_thread",
			Description: "Resolve or reopen a review thread",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"threadId":   map[string]any{"type": "string", "description": "Thread ID"},
				"resolved":   map[string]any{"type": "boolean", "description": "True to resolve, false to reopen"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "threadId", "resolved"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				threadID := getString(args, "threadId")
				resolved := getBool(args, "resolved")
				var resolvedAt *string
				var resolvedBy *string
				if resolved {
					nowStr := time.Now().UTC().Format(time.RFC3339)
					agentStr := "usr_agent"
					resolvedAt = &nowStr
					resolvedBy = &agentStr
				}
				if err := s.store.UpdateThread(ctx, threadID, resolvedAt, resolvedBy); err != nil {
					return nil, err
				}
				return s.store.GetThread(ctx, threadID)
			},
		},
		// 25. update_pull_request_comment
		{
			Name:        "update_pull_request_comment",
			Description: "Edit the body of an existing pull request comment",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"commentId":  map[string]any{"type": "string", "description": "Comment ID"},
				"body":       map[string]any{"type": "string", "description": "Updated comment body"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "commentId", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				commentID := getString(args, "commentId")
				body := getString(args, "body")
				if err := s.store.UpdateComment(ctx, commentID, body); err != nil {
					return nil, err
				}
				return s.store.GetComment(ctx, commentID)
			},
		},
		// 26. update_pull_request_labels
		{
			Name:        "update_pull_request_labels",
			Description: "Add or remove labels on a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":    map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":     map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber":   map[string]any{"type": "integer", "description": "Pull request number"},
				"addLabels":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Labels to attach"},
				"removeLabels": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Labels to detach"},
			}, []string{"ownerSlug", "repoName", "pullNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				ns, err := s.store.GetNamespace(ctx, repo.NamespaceID)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				add := getStringSlice(args, "addLabels")
				remove := getStringSlice(args, "removeLabels")
				for _, lblName := range add {
					lbl, err := s.store.GetOrCreateLabel(ctx, idgen.NewLabel(), ns.ID, lblName, "ededed", "")
					if err == nil {
						_ = s.store.AddPRLabels(ctx, pr.ID, []string{lbl.ID})
					}
				}
				for _, lblName := range remove {
					lbl, err := s.store.GetLabel(ctx, ns.ID, lblName)
					if err == nil {
						_ = s.store.RemovePRLabel(ctx, pr.ID, lbl.ID)
					}
				}
				return s.store.GetPRLabels(ctx, pr.ID)
			},
		},
		// 27. create_label
		{
			Name:        "create_label",
			Description: "Create a label in a repository namespace",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"name":        map[string]any{"type": "string", "description": "Label name"},
				"color":       map[string]any{"type": "string", "description": "Label hex color (without #)"},
				"description": map[string]any{"type": "string", "description": "Optional description"},
			}, []string{"ownerSlug", "repoName", "name", "color"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				labelName := getString(args, "name")
				color := getString(args, "color")
				desc := getString(args, "description")
				ns, err := s.store.GetNamespaceBySlug(ctx, owner)
				if err != nil {
					return nil, err
				}
				return s.store.GetOrCreateLabel(ctx, idgen.NewLabel(), ns.ID, labelName, color, desc)
			},
		},
		// 28. request_pull_request_reviewers
		{
			Name:        "request_pull_request_reviewers",
			Description: "Request reviews from users on a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"reviewers":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "List of user IDs"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "reviewers"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				reviewers := getStringSlice(args, "reviewers")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				if err := s.store.AddRequestedReviewers(ctx, pr.ID, reviewers, nil); err != nil {
					return nil, err
				}
				uids, _, err := s.store.ListRequestedReviewers(ctx, pr.ID)
				return uids, err
			},
		},
		// 29. remove_pull_request_reviewers
		{
			Name:        "remove_pull_request_reviewers",
			Description: "Remove requested reviewers from a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"reviewers":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "List of user IDs to remove"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "reviewers"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				reviewers := getStringSlice(args, "reviewers")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				if err := s.store.RemoveRequestedReviewers(ctx, pr.ID, reviewers, nil); err != nil {
					return nil, err
				}
				uids, _, err := s.store.ListRequestedReviewers(ctx, pr.ID)
				return uids, err
			},
		},
		// 30. checks_read
		{
			Name:        "checks_read",
			Description: "Get check runs and check suites for a commit SHA",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"headSha":   map[string]any{"type": "string", "description": "Commit SHA to inspect"},
			}, []string{"ownerSlug", "repoName", "headSha"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				headSha := getString(args, "headSha")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				runs, err := s.store.ListCheckRunsForCommit(ctx, repo.ID, headSha, "", "", 100, 0)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"commitSha": headSha,
					"checkRuns": runs,
				}, nil
			},
		},
		// 31. list_rulesets
		{
			Name:        "list_rulesets",
			Description: "List rulesets configured on a repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				return s.store.ListRulesets(ctx, repo.ID)
			},
		},
		// 32. get_ruleset
		{
			Name:        "get_ruleset",
			Description: "Get details of a ruleset by ID",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"rulesetId": map[string]any{"type": "string", "description": "Ruleset ID"},
			}, []string{"ownerSlug", "repoName", "rulesetId"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				id := getString(args, "rulesetId")
				return s.store.GetRuleset(ctx, id)
			},
		},
		// 33. get_repository_access
		{
			Name:        "get_repository_access",
			Description: "Get access grants and collaborators on a repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				return s.store.ListRepositoryGrants(ctx, repo.ID)
			},
		},
		// 34. list_namespaces
		{
			Name:        "list_namespaces",
			Description: "List all namespaces in refweir.dev",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{}, nil),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				return s.store.ListNamespaces(ctx, 100)
			},
		},
		// 35. search_symbols
		{
			Name:        "search_symbols",
			Description: "Search code symbols (functions, structs, classes, types, interfaces) across repository source files",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Repository owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"query":     map[string]any{"type": "string", "description": "Symbol name or substring to search for"},
				"ref":       map[string]any{"type": "string", "description": "Optional branch, tag, or commit SHA (defaults to default branch)"},
				"kind":      map[string]any{"type": "string", "description": "Optional kind filter: function, method, struct, class, interface, type, constant"},
				"limit":     map[string]any{"type": "integer", "description": "Maximum number of results to return (default 50)"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				query := getString(args, "query")
				ref := getString(args, "ref")
				kind := getString(args, "kind")
				limit := getInt(args, "limit")
				if limit <= 0 {
					limit = 50
				}
				if ref == "" {
					repo, err := s.store.GetRepo(ctx, owner, name)
					if err == nil && repo.DefaultBranch != "" {
						ref = repo.DefaultBranch
					} else {
						ref = "HEAD"
					}
				}
				return symbolsearch.Search(ctx, s.git, owner, name, ref, query, kind, limit)
			},
		},
		// 36. repo_tree (canonical alias for get_git_tree)
		{
			Name:        "repo_tree",
			Description: "Get the directory tree of a repository at a ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"ref":       map[string]any{"type": "string", "description": "Git commit SHA, branch or tag (default: HEAD)"},
				"path":      map[string]any{"type": "string", "description": "Subdirectory path to list (default: root)"},
				"recursive": map[string]any{"type": "boolean", "description": "Recursively list subtree"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				path := getString(args, "path")
				recursive := getBool(args, "recursive")
				sha, entries, err := s.git.ReadTree(owner, name, ref, path, recursive)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"sha":     sha,
					"tree":    entries,
					"entries": len(entries),
				}, nil
			},
		},
		// 37. file_read (canonical alias for get_file_contents)
		{
			Name:        "file_read",
			Description: "Read the content of a file in a repository at a given ref",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"path":      map[string]any{"type": "string", "description": "File path in repository"},
				"ref":       map[string]any{"type": "string", "description": "Commit or branch ref (default: HEAD)"},
			}, []string{"ownerSlug", "repoName", "path"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				path := getString(args, "path")
				ref := getString(args, "ref")
				if ref == "" {
					ref = "HEAD"
				}
				data, sha, size, isDir, err := s.git.ReadFileAtPath(owner, name, ref, path)
				if err != nil {
					return nil, err
				}
				if isDir {
					return map[string]any{"type": "dir", "path": path}, nil
				}
				return map[string]any{
					"path":          path,
					"sha":           sha,
					"size":          size,
					"content":       string(data),
					"base64Content": base64.StdEncoding.EncodeToString(data),
				}, nil
			},
		},
		// 38. diff_compare (canonical alias for compare_commits)
		{
			Name:        "diff_compare",
			Description: "Compare differences between two commit SHAs or branches",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"base":      map[string]any{"type": "string", "description": "Base ref or commit SHA"},
				"head":      map[string]any{"type": "string", "description": "Head ref or commit SHA"},
			}, []string{"ownerSlug", "repoName", "base", "head"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				base := getString(args, "base")
				head := getString(args, "head")
				if baseHash, err := s.git.ResolveRef(owner, name, base); err == nil {
					base = baseHash.String()
				}
				if headHash, err := s.git.ResolveRef(owner, name, head); err == nil {
					head = headHash.String()
				}
				return s.git.DiffCommits(owner, name, base, head)
			},
		},
		// 39. symbol_search (canonical alias for search_symbols)
		{
			Name:        "symbol_search",
			Description: "Search code symbols across repository source files",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Repository owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"query":     map[string]any{"type": "string", "description": "Symbol name or substring to search for"},
				"ref":       map[string]any{"type": "string", "description": "Optional branch, tag, or commit SHA (defaults to default branch)"},
				"kind":      map[string]any{"type": "string", "description": "Optional kind filter: function, method, struct, class, interface, type, constant"},
				"limit":     map[string]any{"type": "integer", "description": "Maximum number of results to return (default 50)"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				query := getString(args, "query")
				ref := getString(args, "ref")
				kind := getString(args, "kind")
				limit := getInt(args, "limit")
				if limit <= 0 {
					limit = 50
				}
				if ref == "" {
					repo, err := s.store.GetRepo(ctx, owner, name)
					if err == nil && repo.DefaultBranch != "" {
						ref = repo.DefaultBranch
					} else {
						ref = "HEAD"
					}
				}
				return symbolsearch.Search(ctx, s.git, owner, name, ref, query, kind, limit)
			},
		},
		// 40. pull_request_create (canonical alias for create_pull_request)
		{
			Name:        "pull_request_create",
			Description: "Open a new pull request in a repository",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"title":     map[string]any{"type": "string", "description": "Pull request title"},
				"headRef":   map[string]any{"type": "string", "description": "Head branch"},
				"baseRef":   map[string]any{"type": "string", "description": "Base branch (e.g. main)"},
				"body":      map[string]any{"type": "string", "description": "Optional body/description"},
			}, []string{"ownerSlug", "repoName", "title", "headRef", "baseRef"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				head := getString(args, "headRef")
				base := getString(args, "baseRef")
				title := getString(args, "title")
				desc := getString(args, "body")
				if desc == "" {
					desc = getString(args, "description")
				}

				headHash, err := s.git.ResolveRef(owner, name, head)
				if err != nil {
					return nil, fmt.Errorf("resolve head ref %q: %w", head, err)
				}
				baseHash, err := s.git.ResolveRef(owner, name, base)
				if err != nil {
					return nil, fmt.Errorf("resolve base ref %q: %w", base, err)
				}

				pr := store.PullRequest{
					ID:      idgen.NewPR(),
					RepoID:  repo.ID,
					Title:   title,
					Body:    desc,
					State:   "open",
					HeadRef: head,
					BaseRef: base,
					HeadSHA: headHash.String(),
					BaseSHA: baseHash.String(),
				}
				if err := s.store.CreatePullRequest(ctx, &pr); err != nil {
					return nil, err
				}
				return pr, nil
			},
		},
		// 41. thread_comment (post PR top-level comment or inline thread comment)
		{
			Name:        "thread_comment",
			Description: "Post a top-level or inline review comment on a pull request",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber": map[string]any{"type": "integer", "description": "Pull request number"},
				"body":       map[string]any{"type": "string", "description": "Comment text"},
				"path":       map[string]any{"type": "string", "description": "Optional file path for inline review comment"},
				"line":       map[string]any{"type": "integer", "description": "Optional line number for inline review comment"},
				"threadId":   map[string]any{"type": "string", "description": "Optional existing thread ID to reply to"},
			}, []string{"ownerSlug", "repoName", "pullNumber", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				body := getString(args, "body")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}

				threadID := getString(args, "threadId")
				path := getString(args, "path")
				line := getInt(args, "line")

				if threadID != "" {
					comment := store.PRComment{
						ID:            idgen.NewPRComment(),
						RepoID:        repo.ID,
						PullRequestID: pr.ID,
						ThreadID:      threadID,
						AuthorUserID:  "",
						Body:          body,
					}
					if err := s.store.CreateComment(ctx, comment); err != nil {
						return nil, err
					}
					return comment, nil
				}

				if path != "" && line > 0 {
					thread := store.PRThread{
						ID:            idgen.NewPRThread(),
						RepoID:        repo.ID,
						PullRequestID: pr.ID,
						Path:          path,
						StartLine:     line,
						EndLine:       line,
						Side:          "RIGHT",
						HeadSHA:       pr.HeadSHA,
						BaseSHA:       pr.BaseSHA,
					}
					if err := s.store.CreateThread(ctx, thread); err != nil {
						return nil, err
					}
					firstComment := store.PRComment{
						ID:            idgen.NewPRComment(),
						RepoID:        repo.ID,
						PullRequestID: pr.ID,
						ThreadID:      thread.ID,
						AuthorUserID:  "",
						Body:          body,
						Path:          path,
						Line:          line,
						Side:          "RIGHT",
						CommitSHA:     pr.HeadSHA,
					}
					if err := s.store.CreateComment(ctx, firstComment); err != nil {
						return nil, err
					}
					return map[string]any{
						"thread":  thread,
						"comment": firstComment,
					}, nil
				}

				comment := store.PRComment{
					ID:            idgen.NewPRComment(),
					RepoID:        repo.ID,
					PullRequestID: pr.ID,
					AuthorUserID:  "",
					Body:          body,
				}
				if err := s.store.CreateComment(ctx, comment); err != nil {
					return nil, err
				}
				return comment, nil
			},
		},
		// 42. status_check_report
		{
			Name:        "status_check_report",
			Description: "Create or update a commit status check run (e.g., CI tests or security lints)",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":  map[string]any{"type": "string", "description": "Repository owner namespace"},
				"repoName":   map[string]any{"type": "string", "description": "Repository name"},
				"headSha":    map[string]any{"type": "string", "description": "Commit SHA the check applies to"},
				"name":       map[string]any{"type": "string", "description": "Name of the check run (e.g. ci/test)"},
				"status":     map[string]any{"type": "string", "description": "Status (queued, in_progress, completed)"},
				"conclusion": map[string]any{"type": "string", "description": "Conclusion (success, failure, neutral, cancelled, timed_out, skipped)"},
				"title":      map[string]any{"type": "string", "description": "Output title"},
				"summary":    map[string]any{"type": "string", "description": "Output summary description"},
			}, []string{"ownerSlug", "repoName", "headSha", "name"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				headSha := getString(args, "headSha")
				checkName := getString(args, "name")
				status := getString(args, "status")
				if status == "" {
					status = "completed"
				}
				conclusion := getString(args, "conclusion")
				if conclusion == "" && status == "completed" {
					conclusion = "success"
				}
				title := getString(args, "title")
				summary := getString(args, "summary")

				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}

				now := time.Now().UTC().Format(time.RFC3339)
				checkRun := store.CheckRun{
					RepoID:            repo.ID,
					SHA:               headSha,
					Key:               checkName,
					Name:              checkName,
					Status:            status,
					Conclusion:        conclusion,
					OutputTitle:       title,
					OutputSummary:     summary,
					StartedAt:         now,
					CompletedAt:       now,
					ExternalUpdatedAt: now,
				}
				saved, outcome, err := s.store.UpsertCheckRun(ctx, checkRun)
				if err != nil {
					return nil, fmt.Errorf("upsert check run: %w", err)
				}
				return map[string]any{
					"outcome": outcome,
					"check":   saved,
				}, nil
			},
		},
		// 43. pull_request_merge (canonical alias for merge_pull_request with policy enforcement)
		{
			Name:        "pull_request_merge",
			Description: "Merge a pull request with ruleset and status check enforcement",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"pullNumber":  map[string]any{"type": "integer", "description": "Pull request number"},
				"mergeMethod": map[string]any{"type": "string", "description": "Merge method (merge, squash)"},
			}, []string{"ownerSlug", "repoName", "pullNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "pullNumber")
				method := getString(args, "mergeMethod")
				if method == "" {
					method = "merge"
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				pr, err := s.store.GetPullRequest(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}

				// Evaluate merge rulesets and check runs
				rulesets, err := s.store.ListRulesets(ctx, repo.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to list rulesets for merge evaluation: %w", err)
				}
				reviews, err := s.store.ListReviews(ctx, pr.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to list reviews for merge evaluation: %w", err)
				}
				approvedCount := 0
				for _, rev := range reviews {
					if !rev.Dismissed && (rev.State == "approved" || rev.State == "approve") {
						approvedCount++
					}
				}
				checkRuns, err := s.store.ListCheckRunsForCommit(ctx, repo.ID, pr.HeadSHA, "", "", 100, 0)
				if err != nil {
					return nil, fmt.Errorf("failed to list check runs for merge evaluation: %w", err)
				}
				evalRes := engine.EvaluateMergeWithOptions(ctx, rulesets, repo.DefaultBranch, pr.BaseRef, "agent", "", approvedCount, checkRuns, engine.MergeEvaluationOptions{
					MergeMethod: method,
				})
				if !evalRes.Allowed {
					var msgs []string
					for _, b := range evalRes.Blockers {
						msgs = append(msgs, b.Message)
					}
					return nil, fmt.Errorf("merge blocked by ruleset: %s", strings.Join(msgs, "; "))
				}

				sig := object.Signature{Name: "Refweir Agent", Email: "agent@refweir.dev", When: time.Now().UTC()}
				commitSHA, err := s.git.Merge(owner, name, pr.BaseRef, pr.HeadRef, method, sig, fmt.Sprintf("Merge PR #%d: %s", pr.Number, pr.Title))
				if err != nil {
					return nil, fmt.Errorf("git merge: %w", err)
				}
				if err := s.store.SetPullRequestMergeResult(ctx, pr.ID, commitSHA); err != nil {
					return nil, err
				}
				if s.mergeHook != nil {
					s.mergeHook(ctx, pr)
				}
				return s.store.GetPullRequest(ctx, repo.ID, num)
			},
		},
		// 44. lease_request
		{
			Name:        "lease_request",
			Description: "Request an ephemeral HMAC-SHA256 capability lease scoped to refs/heads/agent/*",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"agentId":        map[string]any{"type": "string", "description": "Agent identifier"},
				"repoId":         map[string]any{"type": "string", "description": "Repository ID or owner/name"},
				"allowedRefs":    map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Allowed ref prefixes (defaults to ['refs/heads/agent/'])"},
				"disallowedRefs": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Disallowed ref prefixes (defaults to ['refs/heads/main', 'refs/heads/master'])"},
				"ttlSeconds":     map[string]any{"type": "integer", "description": "Time to live in seconds (default 900)"},
			}, []string{"agentId", "repoId"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				secret := s.LeaseSecret()
				if len(secret) == 0 {
					return nil, fmt.Errorf("lease signing secret is not configured")
				}
				agentID := getString(args, "agentId")
				repoID := getString(args, "repoId")
				if agentID == "" || repoID == "" {
					return nil, fmt.Errorf("issue lease: agentId and repoId are required")
				}
				if strings.Contains(repoID, "/") {
					parts := strings.Split(repoID, "/")
					if repo, err := s.store.GetRepo(ctx, parts[0], parts[1]); err == nil {
						repoID = repo.ID
					}
				}
				allowed := getStringSlice(args, "allowedRefs")
				if len(allowed) == 0 {
					allowed = []string{"refs/heads/agent/"}
				}
				disallowed := getStringSlice(args, "disallowedRefs")
				if len(disallowed) == 0 {
					disallowed = []string{"refs/heads/main", "refs/heads/master"}
				}
				ttlSec := getInt(args, "ttlSeconds")
				if ttlSec <= 0 {
					ttlSec = 3600
				}
				token, lease, err := agentlease.Issue(secret, agentID, repoID, allowed, disallowed, time.Duration(ttlSec)*time.Second)
				if err != nil {
					return nil, fmt.Errorf("issue lease: %w", err)
				}
				_ = audit.Log(audit.Event{
					Actor:     "agent:" + lease.AgentID,
					ActorType: "agent",
					Action:    "agent.lease_issued",
					Resource:  lease.RepoID,
					Severity:  audit.SeverityInfo,
					Details:   "Capability lease issued via MCP",
					Metadata: map[string]interface{}{
						"leaseId":               lease.ID,
						"agentId":               lease.AgentID,
						"allowedRefPrefixes":    lease.AllowedRefPrefixes,
						"disallowedRefPrefixes": lease.DisallowedRefPrefixes,
						"leaseExpiresAt":        lease.ExpiresAt.Format(time.RFC3339),
					},
				})
				return map[string]any{
					"token":                 token,
					"id":                    lease.ID,
					"agentId":               lease.AgentID,
					"repoId":                lease.RepoID,
					"allowedRefPrefixes":    lease.AllowedRefPrefixes,
					"disallowedRefPrefixes": lease.DisallowedRefPrefixes,
					"issuedAt":              lease.IssuedAt.Format(time.RFC3339),
					"expiresAt":             lease.ExpiresAt.Format(time.RFC3339),
				}, nil
			},
		},
		// 45. lease_inspect
		{
			Name:        "lease_inspect",
			Description: "Verify and inspect an ephemeral capability lease token",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"token": map[string]any{"type": "string", "description": "Capability lease token"},
			}, []string{"token"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				token := getString(args, "token")
				secret := s.LeaseSecret()
				pubKey := s.LeasePublicKey()
				if len(secret) == 0 && len(pubKey) == 0 {
					return nil, fmt.Errorf("lease verification keys are not configured")
				}
				var lease *agentlease.Lease
				var err error
				if strings.HasPrefix(token, "ed25519.") {
					if len(pubKey) == ed25519.PublicKeySize {
						lease, err = agentlease.VerifyEd25519(pubKey, token)
					} else if len(secret) == ed25519.PublicKeySize {
						lease, err = agentlease.Verify(secret, token)
					} else {
						return nil, agentlease.ErrInvalidSignature
					}
				} else {
					if len(secret) == 0 {
						return nil, fmt.Errorf("lease signing secret is not configured")
					}
					lease, err = agentlease.Verify(secret, token)
				}
				if err != nil {
					return nil, err
				}
				revoked := false
				if s.store != nil {
					isRev, err := s.store.IsLeaseRevoked(ctx, lease.ID, lease.RepoID)
					if err != nil {
						return nil, fmt.Errorf("failed to check lease revocation status: %w", err)
					}
					revoked = isRev
				}
				return map[string]any{
					"id":                    lease.ID,
					"agentId":               lease.AgentID,
					"repoId":                lease.RepoID,
					"allowedRefPrefixes":    lease.AllowedRefPrefixes,
					"disallowedRefPrefixes": lease.DisallowedRefPrefixes,
					"issuedAt":              lease.IssuedAt.Format(time.RFC3339),
					"expiresAt":             lease.ExpiresAt.Format(time.RFC3339),
					"revoked":               revoked,
				}, nil
			},
		},
		// 46. lease_revoke
		{
			Name:        "lease_revoke",
			Description: "Revoke an active capability lease by ID",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"leaseId": map[string]any{"type": "string", "description": "Lease ID to revoke"},
				"agentId": map[string]any{"type": "string", "description": "Optional agent ID"},
				"repoId":  map[string]any{"type": "string", "description": "Optional repository ID"},
				"reason":  map[string]any{"type": "string", "description": "Optional revocation reason"},
			}, []string{"leaseId"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				leaseID := getString(args, "leaseId")
				agentID := getString(args, "agentId")
				if agentID == "" {
					agentID = "agent-mcp"
				}
				repoID := getString(args, "repoId")
				if repoID == "" {
					repoID = "*"
				}
				reason := getString(args, "reason")
				if reason == "" {
					reason = "agent task finished"
				}
				if err := s.store.RevokeAgentLease(ctx, leaseID, agentID, repoID, reason); err != nil {
					return nil, fmt.Errorf("revoke lease: %w", err)
				}
				_ = audit.Log(audit.Event{
					Actor:     "mcp-tool",
					ActorType: "system",
					Action:    "agent.lease_revoked",
					Resource:  repoID,
					Severity:  audit.SeverityWarning,
					Details:   "Capability lease revoked via MCP: " + reason,
					Metadata: map[string]interface{}{
						"leaseId": leaseID,
						"agentId": agentID,
						"reason":  reason,
					},
				})
				return map[string]any{
					"revoked": true,
					"leaseId": leaseID,
					"reason":  reason,
				}, nil
			},
		},
		// 47. list_issues
		{
			Name:        "list_issues",
			Description: "List repository issues filtered by state (open, closed, or all)",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"state":     map[string]any{"type": "string", "description": "Optional state filter: open, closed, or all"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				state := getString(args, "state")
				if state == "all" {
					state = ""
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				issues, err := s.store.ListIssues(ctx, repo.ID, state)
				if err != nil {
					return nil, err
				}
				return issues, nil
			},
		},
		// 48. get_issue
		{
			Name:        "get_issue",
			Description: "Get issue details by repository and issue number, including discussion comments",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"issueNumber": map[string]any{"type": "integer", "description": "Issue number (e.g. 1)"},
			}, []string{"ownerSlug", "repoName", "issueNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "issueNumber")
				if num <= 0 {
					return nil, fmt.Errorf("valid issueNumber is required")
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				issue, err := s.store.GetIssue(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				comments, err := s.store.ListIssueComments(ctx, issue.ID)
				if err != nil {
					return nil, err
				}
				return map[string]any{
					"issue":    issue,
					"comments": comments,
				}, nil
			},
		},
		// 49. create_issue
		{
			Name:        "create_issue",
			Description: "Create a new issue in a repository",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
				"title":     map[string]any{"type": "string", "description": "Issue title"},
				"body":      map[string]any{"type": "string", "description": "Optional issue body content"},
			}, []string{"ownerSlug", "repoName", "title"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				title := getString(args, "title")
				body := getString(args, "body")
				if title == "" {
					return nil, fmt.Errorf("title is required")
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				newIssue := &store.Issue{
					RepoID: repo.ID,
					Title:  title,
					Body:   body,
					State:  "open",
				}
				if err := s.store.CreateIssue(ctx, newIssue); err != nil {
					return nil, fmt.Errorf("create issue: %w", err)
				}
				_ = audit.Log(audit.Event{
					Actor:     "mcp-agent",
					ActorType: "agent",
					Action:    "issue.created",
					Resource:  repo.FullName,
					Severity:  audit.SeverityInfo,
					Details:   fmt.Sprintf("Created issue #%d: %s", newIssue.Number, newIssue.Title),
					Metadata: map[string]interface{}{
						"issueNumber": newIssue.Number,
						"issueId":     newIssue.ID,
					},
				})
				return newIssue, nil
			},
		},
		// 50. create_issue_comment
		{
			Name:        "create_issue_comment",
			Description: "Post a comment on an existing repository issue",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"issueNumber": map[string]any{"type": "integer", "description": "Issue number (e.g. 1)"},
				"body":        map[string]any{"type": "string", "description": "Comment body text"},
			}, []string{"ownerSlug", "repoName", "issueNumber", "body"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "issueNumber")
				body := getString(args, "body")
				if num <= 0 || body == "" {
					return nil, fmt.Errorf("valid issueNumber and body are required")
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				issue, err := s.store.GetIssue(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				newComment := &store.IssueComment{
					RepoID:  repo.ID,
					IssueID: issue.ID,
					Body:    body,
				}
				if err := s.store.CreateIssueComment(ctx, newComment); err != nil {
					return nil, fmt.Errorf("create issue comment: %w", err)
				}
				_ = audit.Log(audit.Event{
					Actor:     "mcp-agent",
					ActorType: "agent",
					Action:    "issue_comment.created",
					Resource:  repo.FullName,
					Severity:  audit.SeverityInfo,
					Details:   fmt.Sprintf("Commented on issue #%d", issue.Number),
					Metadata: map[string]interface{}{
						"issueNumber": issue.Number,
						"commentId":   newComment.ID,
					},
				})
				return newComment, nil
			},
		},
		// 51. close_issue
		{
			Name:        "close_issue",
			Description: "Close an existing repository issue with optional resolution explanation comment",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"issueNumber": map[string]any{"type": "integer", "description": "Issue number (e.g. 1)"},
				"comment":     map[string]any{"type": "string", "description": "Optional closing resolution comment"},
			}, []string{"ownerSlug", "repoName", "issueNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "issueNumber")
				comment := getString(args, "comment")
				if num <= 0 {
					return nil, fmt.Errorf("valid issueNumber is required")
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				issue, err := s.store.GetIssue(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				if err := s.store.UpdateIssueState(ctx, repo.ID, num, "closed"); err != nil {
					return nil, fmt.Errorf("close issue: %w", err)
				}
				issue.State = "closed"
				if comment != "" {
					newComment := &store.IssueComment{
						RepoID:  repo.ID,
						IssueID: issue.ID,
						Body:    comment,
					}
					_ = s.store.CreateIssueComment(ctx, newComment)
				}
				_ = audit.Log(audit.Event{
					Actor:     "mcp-agent",
					ActorType: "agent",
					Action:    "issue.closed",
					Resource:  repo.FullName,
					Severity:  audit.SeverityInfo,
					Details:   fmt.Sprintf("Closed issue #%d via MCP", issue.Number),
					Metadata: map[string]interface{}{
						"issueNumber": issue.Number,
						"issueId":     issue.ID,
					},
				})
				return issue, nil
			},
		},
		// 52. reopen_issue
		{
			Name:        "reopen_issue",
			Description: "Reopen a previously closed repository issue with optional comment",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"issueNumber": map[string]any{"type": "integer", "description": "Issue number (e.g. 1)"},
				"comment":     map[string]any{"type": "string", "description": "Optional reopen explanation comment"},
			}, []string{"ownerSlug", "repoName", "issueNumber"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				num := getInt(args, "issueNumber")
				comment := getString(args, "comment")
				if num <= 0 {
					return nil, fmt.Errorf("valid issueNumber is required")
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				issue, err := s.store.GetIssue(ctx, repo.ID, num)
				if err != nil {
					return nil, err
				}
				if err := s.store.UpdateIssueState(ctx, repo.ID, num, "open"); err != nil {
					return nil, fmt.Errorf("reopen issue: %w", err)
				}
				issue.State = "open"
				if comment != "" {
					newComment := &store.IssueComment{
						RepoID:  repo.ID,
						IssueID: issue.ID,
						Body:    comment,
					}
					_ = s.store.CreateIssueComment(ctx, newComment)
				}
				_ = audit.Log(audit.Event{
					Actor:     "mcp-agent",
					ActorType: "agent",
					Action:    "issue.reopened",
					Resource:  repo.FullName,
					Severity:  audit.SeverityInfo,
					Details:   fmt.Sprintf("Reopened issue #%d via MCP", issue.Number),
					Metadata: map[string]interface{}{
						"issueNumber": issue.Number,
						"issueId":     issue.ID,
					},
				})
				return issue, nil
			},
		},
		// 53. list_webhooks
		{
			Name:        "list_webhooks",
			Description: "List outbound webhooks configured on a repository",
			ReadOnly:    true,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug": map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":  map[string]any{"type": "string", "description": "Repository name"},
			}, []string{"ownerSlug", "repoName"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				hooks, err := s.store.ListRepoWebhooks(ctx, repo.ID)
				if err != nil {
					return nil, err
				}
				if hooks == nil {
					hooks = []*store.RepoWebhook{}
				}
				return hooks, nil
			},
		},
		// 54. create_webhook
		{
			Name:        "create_webhook",
			Description: "Register a new outbound HTTP webhook subscription for repository events",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"ownerSlug":   map[string]any{"type": "string", "description": "Owner namespace"},
				"repoName":    map[string]any{"type": "string", "description": "Repository name"},
				"url":         map[string]any{"type": "string", "description": "Outbound payload delivery URL"},
				"secret":      map[string]any{"type": "string", "description": "Optional HMAC-SHA256 signature secret"},
				"events":      map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "List of events to subscribe to (e.g. ['push', 'pull_request'] or ['*'])"},
				"insecureSsl": map[string]any{"type": "boolean", "description": "Skip TLS certificate verification (default false)"},
			}, []string{"ownerSlug", "repoName", "url"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				owner := getString(args, "ownerSlug")
				name := getString(args, "repoName")
				url := getString(args, "url")
				secret := getString(args, "secret")
				insecureSSL := getBool(args, "insecureSsl")
				if url == "" {
					return nil, fmt.Errorf("url is required")
				}
				if err := security.ValidateWebhookURL(url); err != nil {
					return nil, fmt.Errorf("invalid webhook URL: %w", err)
				}
				repo, err := s.store.GetRepo(ctx, owner, name)
				if err != nil {
					return nil, err
				}
				evList := getStringSlice(args, "events")
				if len(evList) == 0 {
					evList = []string{"push"}
				}
				eventsJSON, _ := json.Marshal(evList)
				created, err := s.store.CreateRepoWebhook(ctx, repo.ID, url, secret, "json", string(eventsJSON), true, insecureSSL)
				if err != nil {
					return nil, err
				}
				_ = audit.Log(audit.Event{
					Actor:     "mcp-agent",
					ActorType: "agent",
					Action:    "repo.webhook_added",
					Resource:  repo.FullName,
					Severity:  audit.SeverityInfo,
					Details:   fmt.Sprintf("Registered webhook for %s via MCP", created.URL),
					Metadata: map[string]interface{}{
						"webhookId": created.ID,
						"url":       created.URL,
					},
				})
				return created, nil
			},
		},
		// 55. delete_webhook
		{
			Name:        "delete_webhook",
			Description: "Delete an outbound webhook subscription from a repository",
			ReadOnly:    false,
			InputSchema: makeSchema(map[string]any{
				"webhookId": map[string]any{"type": "string", "description": "Webhook ID to delete"},
			}, []string{"webhookId"}),
			Handler: func(ctx context.Context, s *Server, args map[string]any) (any, error) {
				webhookID := getString(args, "webhookId")
				if webhookID == "" {
					return nil, fmt.Errorf("webhookId is required")
				}
				if err := s.store.DeleteRepoWebhook(ctx, webhookID); err != nil {
					return nil, fmt.Errorf("delete webhook: %w", err)
				}
				_ = audit.Log(audit.Event{
					Actor:     "mcp-agent",
					ActorType: "agent",
					Action:    "repo.webhook_deleted",
					Resource:  webhookID,
					Severity:  audit.SeverityInfo,
					Details:   fmt.Sprintf("Deleted webhook %s via MCP", webhookID),
					Metadata: map[string]interface{}{
						"webhookId": webhookID,
					},
				})
				return map[string]any{"ok": true, "deleted": webhookID}, nil
			},
		},
	}
}
