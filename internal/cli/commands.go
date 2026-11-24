package cli

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"refweir.dev/internal/security"
	"refweir.dev/pkg/agentlease"
)

// CLI is the runner for the refweir CLI client.
type CLI struct {
	Endpoint    string
	Token       string
	DefaultRepo string
	HTTPClient  *http.Client
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
}

// New creates a new CLI instance configured from the environment.
func New() *CLI {
	endpoint := os.Getenv("REFWEIR_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8080"
	}
	endpoint = strings.TrimRight(endpoint, "/")

	token := os.Getenv("REFWEIR_TOKEN")
	defaultRepo := os.Getenv("REFWEIR_REPO")

	return &CLI{
		Endpoint:    endpoint,
		Token:       token,
		DefaultRepo: defaultRepo,
		HTTPClient:  http.DefaultClient,
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
	}
}

var Version = "dev"

// Run executes the command specified by args.
func (c *CLI) Run(args []string) error {
	if len(args) == 0 {
		return c.printHelp()
	}

	// Global flags check
	if args[0] == "--help" || args[0] == "-h" {
		return c.printHelp()
	}
	if args[0] == "--version" || args[0] == "-v" {
		fmt.Fprintf(c.Stdout, "refweir version %s\n", Version)
		return nil
	}

	switch args[0] {
	case "auth":
		return c.runAuth(args[1:])
	case "repo":
		return c.runRepo(args[1:])
	case "pr":
		return c.runPR(args[1:])
	case "issue":
		return c.runIssue(args[1:])
	case "webhook":
		return c.runWebhook(args[1:])
	case "lease":
		return c.runLease(args[1:])
	case "app":
		return c.runApp(args[1:])
	case "doctor":
		return c.runDoctor(args[1:])
	case "completion":
		return c.runCompletion(args[1:])
	case "audit":
		return c.runAudit(args[1:])
	case "cluster":
		return c.runCluster(args[1:])
	case "hook":
		return c.runHook(args[1:])
	case "api":
		return c.runAPI(args[1:])
	case "conformance":
		return c.runConformance(args[1:])
	default:
		return fmt.Errorf("unknown command %q. Run 'refweir --help' for usage", args[0])
	}
}

func (c *CLI) printHelp() error {
	help := `refweir <command>

Commands:
  refweir auth               Manage authentication
  refweir repo               Manage repositories
  refweir pr                 Manage pull requests
  refweir issue              Manage issues (list, view, create, close, comment)
  refweir webhook            Manage webhooks (list, create, delete, ping, deliveries, redeliver)
  refweir lease              Manage agent capability leases (issue, revoke, verify)
  refweir app                Manage apps
  refweir doctor             Diagnose daemon connectivity, auth, and local environment
  refweir completion <shell> Generate shell autocompletion script (bash, zsh, fish)
  refweir audit              Inspect and export administrative audit trail (list, export)
  refweir cluster            Manage cluster replication and disaster recovery (status, checkpoint, snapshot, promote)
  refweir hook <action>      Run Git hook handler (pre-receive)
  refweir api <endpoint>     Make an authenticated request to REST API
  refweir conformance        Run API conformance test suite against refweird

Options:
  --version                 Show version number
  --help                    Show help
`
	fmt.Fprint(c.Stdout, help)
	return nil
}

// Request helper
func (c *CLI) doRequest(method, path string, body any) ([]byte, int, error) {
	base := strings.TrimRight(c.Endpoint, "/")
	if strings.HasPrefix(path, "/api/") {
		base = strings.TrimSuffix(base, "/v1/origin")
	}
	url := base + path
	var bodyReader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, err
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	return respBytes, resp.StatusCode, err
}

// Auth commands
func (c *CLI) runAuth(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: refweir auth <status|token|login|logout>")
	}
	switch args[0] {
	case "status":
		if c.Token != "" {
			fmt.Fprintf(c.Stdout, "Logged in to %s (token configured)\n", c.Endpoint)
		} else {
			fmt.Fprintf(c.Stdout, "Not logged in (target: %s)\n", c.Endpoint)
		}
		return nil
	case "token":
		if c.Token != "" {
			fmt.Fprintln(c.Stdout, c.Token)
			return nil
		}
		return errors.New("no auth token configured")
	case "login":
		fmt.Fprintln(c.Stdout, "Set REFWEIR_TOKEN in your environment.")
		return nil
	case "logout":
		c.Token = ""
		fmt.Fprintln(c.Stdout, "Logged out.")
		return nil
	default:
		return fmt.Errorf("unknown auth subcommand %q", args[0])
	}
}

// Repo commands
func (c *CLI) runRepo(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: refweir repo <list|view|get|create|delete|mirror>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("repo list", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		ns := fs.String("namespace", "", "Namespace slug")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		path := "/repos"
		if *ns != "" {
			path = "/repos/" + *ns
		}
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "view", "get":
		if len(args) < 2 {
			return errors.New("usage: refweir repo view <owner/name>")
		}
		owner, name, err := splitRepo(args[1])
		if err != nil {
			return err
		}
		data, code, err := c.doRequest(http.MethodGet, fmt.Sprintf("/repos/%s/%s", owner, name), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "create":
		fs := flag.NewFlagSet("repo create", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		branch := fs.String("default-branch", "main", "Default branch")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: refweir repo create <owner/name>")
		}
		owner, name, err := splitRepo(fs.Arg(0))
		if err != nil {
			return err
		}
		reqBody := map[string]any{
			"repo": map[string]any{
				"name":          name,
				"defaultBranch": *branch,
			},
		}
		data, code, err := c.doRequest(http.MethodPost, "/repos/"+owner, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "delete":
		if len(args) < 2 {
			return errors.New("usage: refweir repo delete <owner/name>")
		}
		owner, name, err := splitRepo(args[1])
		if err != nil {
			return err
		}
		data, code, err := c.doRequest(http.MethodDelete, fmt.Sprintf("/repos/%s/%s", owner, name), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintf(c.Stdout, "Deleted repository %s/%s\n", owner, name)
		return nil

	case "mirror":
		return c.runRepoMirror(args[1:])

	default:
		return fmt.Errorf("unknown repo subcommand %q", args[0])
	}
}

func (c *CLI) runRepoMirror(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: refweir repo mirror <status|sync|detach> <owner/name>")
	}
	sub := args[0]
	if len(args) < 2 {
		return fmt.Errorf("usage: refweir repo mirror %s <owner/name>", sub)
	}
	owner, name, err := splitRepo(args[1])
	if err != nil {
		return err
	}

	switch sub {
	case "status":
		data, code, err := c.doRequest(http.MethodGet, fmt.Sprintf("/repos/%s/%s/mirror", owner, name), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "sync":
		ref := "refs/heads/main"
		if len(args) >= 4 && args[2] == "--ref" {
			ref = args[3]
		}
		reqBody := map[string]any{
			"ref":  ref,
			"wait": true,
		}
		data, code, err := c.doRequest(http.MethodPost, fmt.Sprintf("/repos/%s/%s:syncMirror", owner, name), reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "detach":
		data, code, err := c.doRequest(http.MethodDelete, fmt.Sprintf("/repos/%s/%s/mirror", owner, name), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintf(c.Stdout, "Detached mirror for %s/%s\n", owner, name)
		return nil

	default:
		return fmt.Errorf("unknown mirror subcommand %q", sub)
	}
}

// Pull request commands
func (c *CLI) runPR(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: refweir pr <list|view|create|merge|close>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("pr list", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		status := fs.String("status", "", "Filter by status")
		state := fs.String("state", "", "Filter by state (open, closed, all)")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}

		filterState := "open"
		if *state != "" {
			filterState = *state
		} else if *status != "" {
			filterState = *status
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls?state=%s", owner, name, filterState)
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "view":
		fs := flag.NewFlagSet("pr view", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: refweir pr view <number> [-R owner/name]")
		}
		num := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls/%s", owner, name, num)
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "create":
		fs := flag.NewFlagSet("pr create", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		title := fs.String("title", "", "Pull request title")
		head := fs.String("head", "", "Head branch")
		base := fs.String("base", "main", "Base branch")
		body := fs.String("body", "", "Pull request description")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		if *title == "" || *head == "" {
			return errors.New("flags --title and --head are required")
		}

		reqBody := map[string]any{
			"title":       *title,
			"head":        *head,
			"base":        *base,
			"description": *body,
		}
		path := fmt.Sprintf("/repos/%s/%s/pulls", owner, name)
		data, code, err := c.doRequest(http.MethodPost, path, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "merge":
		fs := flag.NewFlagSet("pr merge", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		method := fs.String("method", "merge", "Merge method (merge, squash)")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: refweir pr merge <number> [-R owner/name]")
		}
		num := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls/%s/merge", owner, name, num)
		reqBody := map[string]any{
			"mergeMethod": *method,
		}
		data, code, err := c.doRequest(http.MethodPost, path, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "close":
		fs := flag.NewFlagSet("pr close", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: refweir pr close <number> [-R owner/name]")
		}
		num := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}

		path := fmt.Sprintf("/repos/%s/%s/pulls/%s", owner, name, num)
		reqBody := map[string]any{
			"state": "closed",
		}
		data, code, err := c.doRequest(http.MethodPatch, path, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintf(c.Stdout, "Closed pull request #%s\n", num)
		return nil

	default:
		return fmt.Errorf("unknown pr subcommand %q", args[0])
	}
}

// Issue commands
func (c *CLI) runIssue(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: refweir issue <list|view|create|close|comment>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("issue list", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		state := fs.String("state", "open", "Filter by state: open, closed, all")
		jsonOut := fs.Bool("json", false, "Output as JSON")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/issues?state=%s", owner, name, *state)
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		var resp struct {
			Issues []struct {
				Number    int    `json:"number"`
				Title     string `json:"title"`
				State     string `json:"state"`
				Author    string `json:"author"`
				Comments  int    `json:"comments"`
				CreatedAt string `json:"createdAt"`
			} `json:"issues"`
		}
		if err := json.Unmarshal(data, &resp); err == nil && len(resp.Issues) > 0 {
			for _, iss := range resp.Issues {
				fmt.Fprintf(c.Stdout, "#%-4d  %-8s  %-50s  (%s)\n", iss.Number, iss.State, iss.Title, iss.Author)
			}
			return nil
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "view":
		fs := flag.NewFlagSet("issue view", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		jsonOut := fs.Bool("json", false, "Output as JSON")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: refweir issue view <number> [-R owner/name]")
		}
		num := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/issues/%s", owner, name, num)
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		var iss struct {
			Number int    `json:"number"`
			Title  string `json:"title"`
			Body   string `json:"body"`
			State  string `json:"state"`
			Author string `json:"author"`
		}
		if err := json.Unmarshal(data, &iss); err == nil && iss.Title != "" {
			fmt.Fprintf(c.Stdout, "#%d %s [%s]\nAuthor: %s\n\n%s\n", iss.Number, iss.Title, iss.State, iss.Author, iss.Body)
			return nil
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "create":
		fs := flag.NewFlagSet("issue create", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		title := fs.String("title", "", "Issue title")
		body := fs.String("body", "", "Issue body")
		jsonOut := fs.Bool("json", false, "Output as JSON")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *title == "" {
			return errors.New("flag --title is required")
		}
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		reqBody := map[string]any{
			"title": *title,
			"body":  *body,
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/issues", owner, name)
		data, code, err := c.doRequest(http.MethodPost, path, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		var resp struct {
			Issue struct {
				Number int    `json:"number"`
				Title  string `json:"title"`
			} `json:"issue"`
		}
		if err := json.Unmarshal(data, &resp); err == nil && resp.Issue.Number > 0 {
			fmt.Fprintf(c.Stdout, "Created issue #%d: %s\n", resp.Issue.Number, resp.Issue.Title)
			return nil
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "close":
		fs := flag.NewFlagSet("issue close", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: refweir issue close <number> [-R owner/name]")
		}
		num := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/issues/%s", owner, name, num)
		reqBody := map[string]any{"state": "closed"}
		data, code, err := c.doRequest(http.MethodPatch, path, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintf(c.Stdout, "Closed issue #%s\n", num)
		return nil

	case "comment":
		fs := flag.NewFlagSet("issue comment", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		body := fs.String("body", "", "Comment body")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 || *body == "" {
			return errors.New("usage: refweir issue comment <number> -R owner/name --body <body>")
		}
		num := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/issues/%s/comments", owner, name, num)
		reqBody := map[string]any{"body": *body}
		data, code, err := c.doRequest(http.MethodPost, path, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintf(c.Stdout, "Added comment to issue #%s\n", num)
		return nil

	default:
		return fmt.Errorf("unknown issue subcommand %q", args[0])
	}
}

// App commands
func (c *CLI) runApp(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: refweir app <list|create>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("app list", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		ns := fs.String("namespace", "", "Namespace slug")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *ns == "" {
			return errors.New("flag --namespace is required")
		}
		data, code, err := c.doRequest(http.MethodGet, fmt.Sprintf("/namespaces/%s/apps", *ns), nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "create":
		fs := flag.NewFlagSet("app create", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		ns := fs.String("namespace", "", "Namespace slug")
		name := fs.String("name", "", "App display name")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *ns == "" || *name == "" {
			return errors.New("flags --namespace and --name are required")
		}
		reqBody := map[string]any{
			"app": map[string]any{
				"displayName": *name,
			},
		}
		data, code, err := c.doRequest(http.MethodPost, fmt.Sprintf("/namespaces/%s/apps", *ns), reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	default:
		return fmt.Errorf("unknown app subcommand %q", args[0])
	}
}

// Arbitrary REST API call (origin api <path>)
func (c *CLI) runAPI(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: origin api <path> [-X METHOD] [-d DATA]")
	}
	fs := flag.NewFlagSet("api", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	method := fs.String("X", "GET", "HTTP method")
	dataFlag := fs.String("d", "", "JSON request body")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() < 1 {
		return errors.New("endpoint path required")
	}
	path := fs.Arg(0)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	var body any
	if *dataFlag != "" {
		var parsed any
		if err := json.Unmarshal([]byte(*dataFlag), &parsed); err == nil {
			body = parsed
		} else {
			body = *dataFlag
		}
	}

	res, code, err := c.doRequest(strings.ToUpper(*method), path, body)
	if err != nil {
		return err
	}
	if code >= 400 {
		return fmt.Errorf("server error (%d): %s", code, string(res))
	}
	fmt.Fprintln(c.Stdout, string(res))
	return nil
}

// Lease commands
func (c *CLI) runLease(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: refweir lease <issue|revoke|verify>")
	}
	switch args[0] {
	case "issue":
		fs := flag.NewFlagSet("lease issue", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		agent := fs.String("agent", "agent:cli", "Agent identifier (e.g. agent:cursor, agent:claude)")
		repoFlag := fs.String("R", c.DefaultRepo, "Target repository owner/name")
		repoLong := fs.String("repo", "", "Target repository owner/name")
		prefix := fs.String("prefix", "refs/heads/agent/*", "Comma-separated allowed ref prefixes")
		disallow := fs.String("disallow", "", "Comma-separated disallowed ref prefixes (e.g. refs/heads/main,refs/heads/master)")
		ttlStr := fs.String("ttl", "30m", "TTL duration (e.g. 30m, 1h, 900s)")
		algoFlag := fs.String("algorithm", "hmac", "Signing algorithm: hmac or ed25519")
		algoShort := fs.String("a", "", "Signing algorithm shorthand (hmac or ed25519)")
		keyFile := fs.String("key-file", "", "Path to signing key or secret file")
		privKey := fs.String("privkey", "", "Inline signing key or secret string")
		ed25519Flag := fs.Bool("ed25519", false, "Use Ed25519 asymmetric signature algorithm")
		jsonOut := fs.Bool("json", false, "Output response in JSON format")

		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}

		repoTarget := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoTarget = *repoLong
		}
		if repoTarget == "" {
			repoTarget = "acme/rocket"
		}

		ttlSeconds := 1800
		if d, err := time.ParseDuration(*ttlStr); err == nil {
			ttlSeconds = int(d.Seconds())
		} else if s, err := strconv.Atoi(*ttlStr); err == nil {
			ttlSeconds = s
		}

		var prefixes []string
		for _, p := range strings.Split(*prefix, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				prefixes = append(prefixes, p)
			}
		}
		if len(prefixes) == 0 {
			prefixes = []string{"refs/heads/agent/*"}
		}

		var disallowed []string
		if *disallow != "" {
			for _, d := range strings.Split(*disallow, ",") {
				d = strings.TrimSpace(d)
				if d != "" {
					disallowed = append(disallowed, d)
				}
			}
		}

		selectedAlgo := strings.ToLower(strings.TrimSpace(*algoFlag))
		if *algoShort != "" {
			selectedAlgo = strings.ToLower(strings.TrimSpace(*algoShort))
		}
		if *ed25519Flag {
			selectedAlgo = "ed25519"
		}

		var rawKey []byte
		if *privKey != "" {
			rawKey = []byte(*privKey)
		} else if *keyFile != "" {
			data, err := os.ReadFile(*keyFile)
			if err != nil {
				return fmt.Errorf("read key file: %w", err)
			}
			rawKey = data
		}

		// Local asymmetric Ed25519 signing
		if selectedAlgo == "ed25519" {
			var priv ed25519.PrivateKey
			var err error
			if len(rawKey) > 0 {
				priv, err = parseEd25519PrivateKey(rawKey)
				if err != nil {
					return fmt.Errorf("parse ed25519 private key: %w", err)
				}
			} else {
				_, priv, err = ed25519.GenerateKey(rand.Reader)
				if err != nil {
					return fmt.Errorf("generate ed25519 key pair: %w", err)
				}
			}

			token, lease, err := agentlease.IssueEd25519(priv, *agent, repoTarget, prefixes, disallowed, time.Duration(ttlSeconds)*time.Second)
			if err != nil {
				return fmt.Errorf("issue ed25519 lease: %w", err)
			}

			pubKey := priv.Public().(ed25519.PublicKey)
			pubKeyB64 := base64.RawURLEncoding.EncodeToString(pubKey)

			if *jsonOut {
				res := map[string]any{
					"id":                    lease.ID,
					"token":                 token,
					"algorithm":             "ed25519",
					"publicKey":             pubKeyB64,
					"agentId":               lease.AgentID,
					"repoId":                lease.RepoID,
					"allowedRefPrefixes":    lease.AllowedRefPrefixes,
					"disallowedRefPrefixes": lease.DisallowedRefPrefixes,
					"issuedAt":              lease.IssuedAt.Format(time.RFC3339),
					"expiresAt":             lease.ExpiresAt.Format(time.RFC3339),
				}
				out, _ := json.MarshalIndent(res, "", "  ")
				fmt.Fprintln(c.Stdout, string(out))
				return nil
			}

			fmt.Fprintf(c.Stdout, "Capability Lease Issued: %s\n", lease.ID)
			fmt.Fprintf(c.Stdout, "Algorithm:  ed25519\n")
			fmt.Fprintf(c.Stdout, "Agent:      %s\n", lease.AgentID)
			fmt.Fprintf(c.Stdout, "Repo:       %s\n", lease.RepoID)
			fmt.Fprintf(c.Stdout, "Expires:    %s\n", lease.ExpiresAt.Format(time.RFC3339))
			fmt.Fprintf(c.Stdout, "Public Key: %s\n", pubKeyB64)
			fmt.Fprintf(c.Stdout, "Token:      %s\n", token)
			return nil
		}

		// Local symmetric HMAC signing if key/secret is supplied
		if len(rawKey) > 0 {
			token, lease, err := agentlease.Issue(bytes.TrimSpace(rawKey), *agent, repoTarget, prefixes, disallowed, time.Duration(ttlSeconds)*time.Second)
			if err != nil {
				return fmt.Errorf("issue hmac lease: %w", err)
			}

			if *jsonOut {
				res := map[string]any{
					"id":                    lease.ID,
					"token":                 token,
					"algorithm":             "hmac-sha256",
					"agentId":               lease.AgentID,
					"repoId":                lease.RepoID,
					"allowedRefPrefixes":    lease.AllowedRefPrefixes,
					"disallowedRefPrefixes": lease.DisallowedRefPrefixes,
					"issuedAt":              lease.IssuedAt.Format(time.RFC3339),
					"expiresAt":             lease.ExpiresAt.Format(time.RFC3339),
				}
				out, _ := json.MarshalIndent(res, "", "  ")
				fmt.Fprintln(c.Stdout, string(out))
				return nil
			}

			fmt.Fprintf(c.Stdout, "Capability Lease Issued: %s\n", lease.ID)
			fmt.Fprintf(c.Stdout, "Algorithm:  hmac-sha256\n")
			fmt.Fprintf(c.Stdout, "Agent:      %s\n", lease.AgentID)
			fmt.Fprintf(c.Stdout, "Repo:       %s\n", lease.RepoID)
			fmt.Fprintf(c.Stdout, "Expires:    %s\n", lease.ExpiresAt.Format(time.RFC3339))
			fmt.Fprintf(c.Stdout, "Token:      %s\n", token)
			return nil
		}

		// Default daemon-issued capability lease via HTTP API
		reqBody := map[string]any{
			"agentId":         *agent,
			"repoId":          repoTarget,
			"allowedPrefixes": prefixes,
			"ttlSeconds":      ttlSeconds,
		}
		if len(disallowed) > 0 {
			reqBody["disallowedRefPrefixes"] = disallowed
		}

		data, code, err := c.doRequest(http.MethodPost, "/api/v1/agent/leases", reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}

		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}

		var resp struct {
			Token     string `json:"token"`
			ID        string `json:"id"`
			AgentID   string `json:"agentId"`
			RepoID    string `json:"repoId"`
			ExpiresAt string `json:"expiresAt"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}

		fmt.Fprintf(c.Stdout, "Capability Lease Issued: %s\n", resp.ID)
		fmt.Fprintf(c.Stdout, "Agent:     %s\n", resp.AgentID)
		fmt.Fprintf(c.Stdout, "Repo:      %s\n", resp.RepoID)
		fmt.Fprintf(c.Stdout, "Expires:   %s\n", resp.ExpiresAt)
		fmt.Fprintf(c.Stdout, "Token:     %s\n", resp.Token)
		return nil

	case "revoke":
		fs := flag.NewFlagSet("lease revoke", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		reason := fs.String("reason", "operator CLI revocation", "Reason for revoking capability lease")
		jsonOut := fs.Bool("json", false, "Output response in JSON format")

		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}

		if fs.NArg() == 0 {
			return errors.New("lease ID is required: refweir lease revoke <lease-id>")
		}
		leaseID := fs.Arg(0)

		reqBody := map[string]any{
			"reason": *reason,
		}

		data, code, err := c.doRequest(http.MethodPost, fmt.Sprintf("/api/v1/agent/leases/%s/revoke", leaseID), reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}

		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}

		fmt.Fprintf(c.Stdout, "Revoked capability lease %s\n", leaseID)
		return nil

	case "verify":
		fs := flag.NewFlagSet("lease verify", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		pubKeyFlag := fs.String("pubkey", "", "Ed25519 public key (hex, base64, or raw)")
		secretFlag := fs.String("secret", "", "HMAC signing secret string")
		keyFile := fs.String("key-file", "", "Path to public key or secret file")
		jsonOut := fs.Bool("json", false, "Output response in JSON format")

		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}

		if fs.NArg() == 0 {
			return errors.New("lease token is required: refweir lease verify <token>")
		}
		token := strings.TrimSpace(fs.Arg(0))

		var verificationKey []byte
		if *pubKeyFlag != "" {
			verificationKey = []byte(*pubKeyFlag)
		} else if *secretFlag != "" {
			verificationKey = []byte(*secretFlag)
		} else if *keyFile != "" {
			data, err := os.ReadFile(*keyFile)
			if err != nil {
				return fmt.Errorf("read key file: %w", err)
			}
			verificationKey = data
		}

		isEd := strings.HasPrefix(token, "ed25519.")
		algo := "hmac-sha256"
		if isEd {
			algo = "ed25519"
		}

		var lease *agentlease.Lease
		var verifyErr error
		signatureVerified := false

		if isEd {
			if len(verificationKey) > 0 {
				pub, err := parseEd25519PublicKey(verificationKey)
				if err != nil {
					if priv, privErr := parseEd25519PrivateKey(verificationKey); privErr == nil {
						pub = priv.Public().(ed25519.PublicKey)
						err = nil
					}
				}
				if err != nil {
					return fmt.Errorf("invalid ed25519 public key: %w", err)
				}
				lease, verifyErr = agentlease.VerifyEd25519(pub, token)
				if verifyErr == nil || errors.Is(verifyErr, agentlease.ErrExpiredLease) {
					signatureVerified = true
				}
			} else {
				parts := strings.Split(token, ".")
				if len(parts) == 3 {
					if payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1]); err == nil {
						var parsed agentlease.Lease
						if json.Unmarshal(payloadBytes, &parsed) == nil {
							lease = &parsed
							if time.Now().UTC().After(lease.ExpiresAt) {
								verifyErr = agentlease.ErrExpiredLease
							}
						}
					}
				}
			}
		} else {
			if len(verificationKey) > 0 {
				lease, verifyErr = agentlease.Verify(bytes.TrimSpace(verificationKey), token)
				if verifyErr == nil || errors.Is(verifyErr, agentlease.ErrExpiredLease) {
					signatureVerified = true
				}
			} else {
				parts := strings.Split(token, ".")
				if len(parts) == 2 {
					if payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0]); err == nil {
						var parsed agentlease.Lease
						if json.Unmarshal(payloadBytes, &parsed) == nil {
							lease = &parsed
							if time.Now().UTC().After(lease.ExpiresAt) {
								verifyErr = agentlease.ErrExpiredLease
							}
						}
					}
				}
			}
		}

		if lease == nil {
			return fmt.Errorf("invalid or malformed capability lease token: %v", verifyErr)
		}

		if len(verificationKey) > 0 && !signatureVerified {
			if verifyErr != nil {
				return fmt.Errorf("verification failed: %w", verifyErr)
			}
			return errors.New("verification failed: invalid lease signature")
		}

		expired := time.Now().UTC().After(lease.ExpiresAt)
		status := "VALID"
		if expired {
			status = "EXPIRED"
		}

		if *jsonOut {
			res := map[string]any{
				"valid":                 !expired && (len(verificationKey) == 0 || signatureVerified),
				"status":                status,
				"signatureVerified":     signatureVerified,
				"algorithm":             algo,
				"id":                    lease.ID,
				"agentId":               lease.AgentID,
				"repoId":                lease.RepoID,
				"allowedRefPrefixes":    lease.AllowedRefPrefixes,
				"disallowedRefPrefixes": lease.DisallowedRefPrefixes,
				"issuedAt":              lease.IssuedAt.Format(time.RFC3339),
				"expiresAt":             lease.ExpiresAt.Format(time.RFC3339),
				"expired":               expired,
			}
			out, _ := json.MarshalIndent(res, "", "  ")
			fmt.Fprintln(c.Stdout, string(out))
			if expired {
				return agentlease.ErrExpiredLease
			}
			return nil
		}

		fmt.Fprintf(c.Stdout, "Capability Lease Status: %s\n", status)
		fmt.Fprintf(c.Stdout, "Algorithm:  %s\n", algo)
		if signatureVerified {
			fmt.Fprintf(c.Stdout, "Signature:  VERIFIED\n")
		} else {
			fmt.Fprintf(c.Stdout, "Signature:  UNVERIFIED (provide --pubkey, --secret, or --key-file to verify cryptographic signature)\n")
		}
		fmt.Fprintf(c.Stdout, "Lease ID:   %s\n", lease.ID)
		fmt.Fprintf(c.Stdout, "Agent:      %s\n", lease.AgentID)
		fmt.Fprintf(c.Stdout, "Repo:       %s\n", lease.RepoID)
		fmt.Fprintf(c.Stdout, "Issued:     %s\n", lease.IssuedAt.Format(time.RFC3339))
		fmt.Fprintf(c.Stdout, "Expires:    %s\n", lease.ExpiresAt.Format(time.RFC3339))
		if len(lease.AllowedRefPrefixes) > 0 {
			fmt.Fprintf(c.Stdout, "Allowed:    %s\n", strings.Join(lease.AllowedRefPrefixes, ", "))
		}
		if len(lease.DisallowedRefPrefixes) > 0 {
			fmt.Fprintf(c.Stdout, "Disallowed: %s\n", strings.Join(lease.DisallowedRefPrefixes, ", "))
		}
		if expired {
			return agentlease.ErrExpiredLease
		}
		return nil

	default:
		return fmt.Errorf("unknown lease subcommand %q", args[0])
	}
}

func (c *CLI) runWebhook(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: refweir webhook <list|create|delete|ping|deliveries|redeliver>")
	}
	switch args[0] {
	case "list":
		fs := flag.NewFlagSet("webhook list", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		jsonOut := fs.Bool("json", false, "Output as JSON")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/hooks", owner, name)
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		var hooks []struct {
			ID          string `json:"id"`
			URL         string `json:"url"`
			Events      string `json:"events"`
			Active      bool   `json:"active"`
			InsecureSSL bool   `json:"insecure_ssl"`
			CreatedAt   string `json:"created_at"`
		}
		if err := json.Unmarshal(data, &hooks); err == nil && len(hooks) > 0 {
			for _, h := range hooks {
				status := "active"
				if !h.Active {
					status = "inactive"
				}
				fmt.Fprintf(c.Stdout, "%-16s  %-8s  %-24s  %s\n", h.ID, status, h.Events, h.URL)
			}
			return nil
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "create":
		fs := flag.NewFlagSet("webhook create", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		targetURL := fs.String("url", "", "Webhook target payload URL")
		secret := fs.String("secret", "", "Webhook HMAC secret")
		events := fs.String("events", "push", "Comma-separated list of events or *")
		insecureSSL := fs.Bool("insecure-ssl", false, "Skip SSL verification")
		active := fs.Bool("active", true, "Whether the webhook is active")
		jsonOut := fs.Bool("json", false, "Output as JSON")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if *targetURL == "" {
			return errors.New("--url is required")
		}
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		var evList []string
		for _, e := range strings.Split(*events, ",") {
			e = strings.TrimSpace(e)
			if e != "" {
				evList = append(evList, e)
			}
		}
		if len(evList) == 0 {
			evList = []string{"push"}
		}
		reqBody := map[string]any{
			"url":          *targetURL,
			"secret":       *secret,
			"content_type": "json",
			"events":       evList,
			"insecure_ssl": *insecureSSL,
			"active":       active,
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/hooks", owner, name)
		data, code, err := c.doRequest(http.MethodPost, path, reqBody)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		var created struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		}
		if err := json.Unmarshal(data, &created); err == nil && created.ID != "" {
			fmt.Fprintf(c.Stdout, "Created webhook %s -> %s\n", created.ID, created.URL)
			return nil
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "delete":
		fs := flag.NewFlagSet("webhook delete", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		jsonOut := fs.Bool("json", false, "Output as JSON")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: refweir webhook delete <id> [-R owner/name]")
		}
		hookID := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/hooks/%s", owner, name, hookID)
		data, code, err := c.doRequest(http.MethodDelete, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		fmt.Fprintf(c.Stdout, "Deleted webhook %s\n", hookID)
		return nil

	case "ping", "test":
		fs := flag.NewFlagSet("webhook ping", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		jsonOut := fs.Bool("json", false, "Output as JSON")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: refweir webhook ping <id> [-R owner/name]")
		}
		hookID := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/hooks/%s/tests", owner, name, hookID)
		data, code, err := c.doRequest(http.MethodPost, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		var resp struct {
			OK         bool   `json:"ok"`
			DeliveryID string `json:"delivery_id"`
			Status     string `json:"status"`
		}
		if err := json.Unmarshal(data, &resp); err == nil && resp.DeliveryID != "" {
			fmt.Fprintf(c.Stdout, "Ping dispatched (delivery %s, status: %s)\n", resp.DeliveryID, resp.Status)
			return nil
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "deliveries":
		fs := flag.NewFlagSet("webhook deliveries", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		jsonOut := fs.Bool("json", false, "Output as JSON")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 1 {
			return errors.New("usage: refweir webhook deliveries <id> [-R owner/name]")
		}
		hookID := fs.Arg(0)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/hooks/%s/deliveries", owner, name, hookID)
		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		var dels []struct {
			ID                 string `json:"id"`
			Event              string `json:"event"`
			Delivered          bool   `json:"delivered"`
			ResponseStatusCode int    `json:"response_status_code"`
			DurationMs         int    `json:"duration_ms"`
			CreatedAt          string `json:"created_at"`
		}
		if err := json.Unmarshal(data, &dels); err == nil && len(dels) > 0 {
			for _, d := range dels {
				status := fmt.Sprintf("%d", d.ResponseStatusCode)
				if d.Delivered {
					status = fmt.Sprintf("%d OK", d.ResponseStatusCode)
				}
				fmt.Fprintf(c.Stdout, "%-16s  %-12s  %-10s  %dms  %s\n", d.ID, d.Event, status, d.DurationMs, d.CreatedAt)
			}
			return nil
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	case "redeliver":
		fs := flag.NewFlagSet("webhook redeliver", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		repoFlag := fs.String("R", c.DefaultRepo, "Repository (owner/name)")
		repoLong := fs.String("repo", c.DefaultRepo, "Repository (owner/name)")
		jsonOut := fs.Bool("json", false, "Output as JSON")
		if err := fs.Parse(reorderFlags(args[1:])); err != nil {
			return err
		}
		if fs.NArg() < 2 {
			return errors.New("usage: refweir webhook redeliver <id> <delivery-id> [-R owner/name]")
		}
		hookID := fs.Arg(0)
		delID := fs.Arg(1)
		repoStr := *repoFlag
		if *repoLong != "" && *repoLong != c.DefaultRepo {
			repoStr = *repoLong
		}
		owner, name, err := splitRepo(repoStr)
		if err != nil {
			return err
		}
		path := fmt.Sprintf("/api/v1/repos/%s/%s/hooks/%s/deliveries/%s/redeliver", owner, name, hookID, delID)
		data, code, err := c.doRequest(http.MethodPost, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		var resp struct {
			OK           bool   `json:"ok"`
			RedeliveryID string `json:"redelivery_id"`
		}
		if err := json.Unmarshal(data, &resp); err == nil && resp.RedeliveryID != "" {
			fmt.Fprintf(c.Stdout, "Redelivery triggered successfully (new delivery ID: %s)\n", resp.RedeliveryID)
			return nil
		}
		fmt.Fprintln(c.Stdout, string(data))
		return nil

	default:
		return fmt.Errorf("unknown webhook subcommand %q. Run 'refweir webhook' for usage", args[0])
	}
}

func splitRepo(repo string) (string, string, error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repository target %q: expected owner/repo", repo)
	}
	return parts[0], parts[1], nil
}

func reorderFlags(args []string) []string {
	var flags, pos []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			flags = append(flags, arg)
			if !strings.Contains(arg, "=") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags = append(flags, args[i+1])
				i++
			}
		} else {
			pos = append(pos, arg)
		}
	}
	return append(flags, pos...)
}

type doctorCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "pass", "warn", "fail"
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (c *CLI) runDoctor(args []string) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(c.Stderr)
	jsonOut := fs.Bool("json", false, "Output report in JSON format")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var checks []doctorCheck
	allHealthy := true

	// 1. Daemon Connectivity & Health
	connCheck := doctorCheck{Name: "Daemon Connectivity"}
	healthPath := "/api/v1/admin/health"
	data, code, err := c.doRequest(http.MethodGet, healthPath, nil)
	if err != nil || code != http.StatusOK {
		data, code, err = c.doRequest(http.MethodGet, "/healthz", nil)
	}

	if err != nil {
		connCheck.Status = "fail"
		connCheck.Message = fmt.Sprintf("Failed to reach daemon at %s: %v", c.Endpoint, err)
		allHealthy = false
	} else if code >= 400 {
		connCheck.Status = "fail"
		connCheck.Message = fmt.Sprintf("Daemon responded with HTTP %d at %s", code, c.Endpoint)
		allHealthy = false
	} else {
		connCheck.Status = "pass"
		connCheck.Message = fmt.Sprintf("Connected to Refweir daemon at %s (HTTP %d)", c.Endpoint, code)
		var hInfo struct {
			ServerStatus       string `json:"serverStatus"`
			Version            string `json:"version"`
			Uptime             string `json:"uptime"`
			Goroutines         int    `json:"goroutines"`
			SqliteWalSizeBytes int64  `json:"sqliteWalSizeBytes"`
			AllocatedMemoryMb  string `json:"allocatedMemoryMb"`
		}
		if err := json.Unmarshal(data, &hInfo); err == nil && hInfo.Version != "" {
			connCheck.Details = fmt.Sprintf("version=%s, uptime=%s, goroutines=%d, wal=%dKB, memory=%sMB",
				hInfo.Version, hInfo.Uptime, hInfo.Goroutines, hInfo.SqliteWalSizeBytes/1024, hInfo.AllocatedMemoryMb)
		}
	}
	checks = append(checks, connCheck)

	// 2. Authentication
	authCheck := doctorCheck{Name: "Authentication"}
	if c.Token == "" {
		authCheck.Status = "warn"
		authCheck.Message = "No REFWEIR_TOKEN configured (anonymous access only)"
	} else {
		sessData, sessCode, sessErr := c.doRequest(http.MethodGet, "/api/v1/auth/session", nil)
		if sessErr == nil && sessCode == http.StatusOK {
			var sess struct {
				User struct {
					ID    string `json:"id"`
					Login string `json:"login"`
					Role  string `json:"role"`
				} `json:"user"`
			}
			if err := json.Unmarshal(sessData, &sess); err == nil && sess.User.Login != "" {
				authCheck.Status = "pass"
				authCheck.Message = fmt.Sprintf("Authenticated as %s (role: %s)", sess.User.Login, sess.User.Role)
			} else {
				authCheck.Status = "pass"
				authCheck.Message = "Authenticated with active token"
			}
		} else {
			authCheck.Status = "warn"
			authCheck.Message = fmt.Sprintf("Token present but session verification returned HTTP %d", sessCode)
		}
	}
	checks = append(checks, authCheck)

	// 3. Local Git
	gitCheck := doctorCheck{Name: "Local Git"}
	if gitOut, gitErr := exec.Command("git", "version").Output(); gitErr != nil {
		gitCheck.Status = "fail"
		gitCheck.Message = "git binary not found in PATH"
		allHealthy = false
	} else {
		gitCheck.Status = "pass"
		gitCheck.Message = strings.TrimSpace(string(gitOut))
	}
	checks = append(checks, gitCheck)

	// 4. Local SSH
	sshCheck := doctorCheck{Name: "Local SSH"}
	if sshOut, sshErr := exec.Command("ssh", "-V").CombinedOutput(); sshErr != nil && len(sshOut) == 0 {
		sshCheck.Status = "warn"
		sshCheck.Message = "ssh client not found in PATH"
	} else {
		sshCheck.Status = "pass"
		sshCheck.Message = strings.TrimSpace(string(sshOut))
	}
	checks = append(checks, sshCheck)

	if *jsonOut {
		report := map[string]any{
			"healthy":  allHealthy,
			"endpoint": c.Endpoint,
			"checks":   checks,
		}
		enc := json.NewEncoder(c.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}

	fmt.Fprintln(c.Stdout, "======================================================")
	fmt.Fprintln(c.Stdout, "              Refweir Diagnostic Report               ")
	fmt.Fprintln(c.Stdout, "======================================================")
	fmt.Fprintf(c.Stdout, "Endpoint: %s\n\n", c.Endpoint)

	for _, chk := range checks {
		symbol := "[✓]"
		if chk.Status == "warn" {
			symbol = "[!]"
		} else if chk.Status == "fail" {
			symbol = "[✗]"
		}
		fmt.Fprintf(c.Stdout, "%s %s: %s\n", symbol, chk.Name, chk.Message)
		if chk.Details != "" {
			fmt.Fprintf(c.Stdout, "    Details: %s\n", chk.Details)
		}
	}
	fmt.Fprintln(c.Stdout, "")
	if allHealthy {
		fmt.Fprintln(c.Stdout, "Result: All systems operational.")
	} else {
		fmt.Fprintln(c.Stdout, "Result: Diagnostic detected warnings or failures.")
	}
	fmt.Fprintln(c.Stdout, "======================================================")
	return nil
}

func (c *CLI) runCompletion(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("shell required. Usage: refweir completion <bash|zsh|fish>")
	}
	shell := strings.ToLower(args[0])
	switch shell {
	case "bash":
		return c.printBashCompletion()
	case "zsh":
		return c.printZshCompletion()
	case "fish":
		return c.printFishCompletion()
	default:
		return fmt.Errorf("unsupported shell %q. Supported shells: bash, zsh, fish", shell)
	}
}

func (c *CLI) printBashCompletion() error {
	script := `# bash completion for refweir
_refweir_completion() {
    local cur prev words cword
    _init_completion || return

    local commands="auth repo pr issue webhook lease app doctor completion audit cluster hook api conformance --version --help"

    if [ $cword -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${commands}" -- "$cur") )
        return 0
    fi

    case "${words[1]}" in
        repo)
            COMPREPLY=( $(compgen -W "list view get create mirror" -- "$cur") )
            ;;
        pr)
            COMPREPLY=( $(compgen -W "list view create close" -- "$cur") )
            ;;
        issue)
            COMPREPLY=( $(compgen -W "list view create close comment" -- "$cur") )
            ;;
        webhook)
            COMPREPLY=( $(compgen -W "list create delete ping deliveries redeliver" -- "$cur") )
            ;;
        lease)
            COMPREPLY=( $(compgen -W "issue revoke verify" -- "$cur") )
            ;;
        completion)
            COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
            ;;
        audit)
            COMPREPLY=( $(compgen -W "list export" -- "$cur") )
            ;;
        cluster)
            COMPREPLY=( $(compgen -W "status checkpoint snapshot promote" -- "$cur") )
            ;;
        hook)
            COMPREPLY=( $(compgen -W "pre-receive" -- "$cur") )
            ;;
        doctor)
            COMPREPLY=( $(compgen -W "--json" -- "$cur") )
            ;;
    esac
}
complete -F _refweir_completion refweir
`
	fmt.Fprint(c.Stdout, script)
	return nil
}

func (c *CLI) printZshCompletion() error {
	script := `#compdef refweir

_refweir() {
    local -a commands
    commands=(
        'auth:Manage authentication'
        'repo:Manage repositories'
        'pr:Manage pull requests'
        'issue:Manage issues'
        'webhook:Manage webhooks'
        'lease:Manage agent capability leases'
        'app:Manage apps'
        'doctor:Diagnose daemon connectivity and environment'
        'completion:Generate shell completion script'
        'audit:Inspect administrative audit trail'
        'cluster:Manage cluster replication and disaster recovery'
        'hook:Run Git hook handler'
        'api:Make authenticated request to REST API'
        'conformance:Run API conformance test suite'
    )

    _arguments -C \
        '--version[Show version number]' \
        '--help[Show help]' \
        '1: :->command' \
        '*:: :->args'

    case $state in
        command)
            _describe -t commands 'refweir command' commands
            ;;
    esac
}

_refweir "$@"
`
	fmt.Fprint(c.Stdout, script)
	return nil
}

func (c *CLI) printFishCompletion() error {
	script := `# fish completion for refweir
complete -c refweir -f
complete -c refweir -n "__fish_use_subcommand" -a "auth" -d "Manage authentication"
complete -c refweir -n "__fish_use_subcommand" -a "repo" -d "Manage repositories"
complete -c refweir -n "__fish_use_subcommand" -a "pr" -d "Manage pull requests"
complete -c refweir -n "__fish_use_subcommand" -a "issue" -d "Manage issues"
complete -c refweir -n "__fish_use_subcommand" -a "webhook" -d "Manage webhooks"
complete -c refweir -n "__fish_use_subcommand" -a "lease" -d "Manage agent capability leases"
complete -c refweir -n "__fish_use_subcommand" -a "app" -d "Manage apps"
complete -c refweir -n "__fish_use_subcommand" -a "doctor" -d "Diagnose daemon connectivity and environment"
complete -c refweir -n "__fish_use_subcommand" -a "completion" -d "Generate shell completion script"
complete -c refweir -n "__fish_use_subcommand" -a "audit" -d "Inspect administrative audit trail"
complete -c refweir -n "__fish_use_subcommand" -a "cluster" -d "Manage cluster replication and disaster recovery"
complete -c refweir -n "__fish_use_subcommand" -a "hook" -d "Run Git hook handler"
complete -c refweir -n "__fish_use_subcommand" -a "api" -d "Make authenticated REST request"
complete -c refweir -n "__fish_use_subcommand" -a "conformance" -d "Run API conformance suite"
`
	fmt.Fprint(c.Stdout, script)
	return nil
}

func (c *CLI) runHook(args []string) error {
	if len(args) > 0 && args[0] == "pre-receive" {
		in := c.Stdin
		if in == nil {
			in = os.Stdin
		}
		errOut := c.Stderr
		if errOut == nil {
			errOut = os.Stderr
		}
		return security.PreReceiveHook(in, errOut)
	}
	return errors.New("usage: refweir hook pre-receive")
}

func (c *CLI) runAudit(args []string) error {
	if len(args) == 0 {
		args = []string{"list"}
	}
	sub := args[0]
	subArgs := args[1:]
	if strings.HasPrefix(sub, "-") {
		sub = "list"
		subArgs = args
	}

	switch sub {
	case "list":
		fs := flag.NewFlagSet("audit list", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		actor := fs.String("actor", "", "Filter by actor identifier")
		action := fs.String("action", "", "Filter by action name")
		limit := fs.Int("limit", 30, "Maximum number of events to return")
		jsonOut := fs.Bool("json", false, "Output events in JSON format")

		if err := fs.Parse(reorderFlags(subArgs)); err != nil {
			return err
		}

		path := fmt.Sprintf("/api/v1/admin/audit-logs?limit=%d", *limit)
		if *actor != "" {
			path += "&actor=" + *actor
		}
		if *action != "" {
			path += "&action=" + *action
		}

		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}

		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}

		var resp struct {
			Events []struct {
				ID        string `json:"id"`
				Timestamp string `json:"timestamp"`
				Actor     string `json:"actor"`
				Action    string `json:"action"`
				Resource  string `json:"resource"`
				Severity  string `json:"severity"`
				Details   string `json:"details"`
			} `json:"events"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}

		if len(resp.Events) == 0 {
			fmt.Fprintln(c.Stdout, "No audit events found.")
			return nil
		}

		fmt.Fprintf(c.Stdout, "%-16s %-20s %-16s %-24s %-20s %s\n", "ID", "TIME", "ACTOR", "ACTION", "RESOURCE", "SEVERITY")
		for _, e := range resp.Events {
			tStr := e.Timestamp
			if len(tStr) > 19 {
				tStr = tStr[:19]
			}
			fmt.Fprintf(c.Stdout, "%-16s %-20s %-16s %-24s %-20s %s\n", e.ID, tStr, e.Actor, e.Action, e.Resource, e.Severity)
		}
		return nil

	case "export":
		fs := flag.NewFlagSet("audit export", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		since := fs.String("since", "", "Filter events after duration (e.g. 24h, 7d) or RFC3339 timestamp")
		until := fs.String("until", "", "Filter events before RFC3339 timestamp")
		actor := fs.String("actor", "", "Filter by actor identifier")
		action := fs.String("action", "", "Filter by action name")
		severity := fs.String("severity", "", "Filter by severity (info, warning, critical)")
		format := fs.String("format", "ndjson", "Export format: ndjson, csv, or json")
		output := fs.String("output", "", "Output file path (defaults to stdout)")
		oShort := fs.String("o", "", "Output file path shorthand")
		limit := fs.Int("limit", 10000, "Maximum number of events to export (0 for unlimited)")

		if err := fs.Parse(reorderFlags(subArgs)); err != nil {
			return err
		}

		outPath := *output
		if *oShort != "" {
			outPath = *oShort
		}

		path := fmt.Sprintf("/api/v1/admin/audit-logs/export?format=%s&limit=%d", *format, *limit)
		if *since != "" {
			path += "&since=" + *since
		}
		if *until != "" {
			path += "&until=" + *until
		}
		if *actor != "" {
			path += "&actor=" + *actor
		}
		if *action != "" {
			path += "&action=" + *action
		}
		if *severity != "" {
			path += "&severity=" + *severity
		}

		data, code, err := c.doRequest(http.MethodGet, path, nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}

		if outPath != "" {
			if err := os.WriteFile(outPath, data, 0o644); err != nil {
				return fmt.Errorf("write export file: %w", err)
			}
			fmt.Fprintf(c.Stdout, "Exported audit logs to %s (%d bytes)\n", outPath, len(data))
			return nil
		}

		_, _ = c.Stdout.Write(data)
		if *format == "json" && !bytes.HasSuffix(data, []byte("\n")) {
			fmt.Fprintln(c.Stdout)
		}
		return nil

	default:
		return fmt.Errorf("unknown audit subcommand %q. Run 'refweir audit <list|export>' for usage", sub)
	}
}

func (c *CLI) runCluster(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: refweir cluster <status|checkpoint|snapshot|promote>")
	}

	sub := args[0]
	subArgs := args[1:]

	switch sub {
	case "status":
		fs := flag.NewFlagSet("cluster status", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		jsonOut := fs.Bool("json", false, "Output cluster replication status in JSON format")
		if err := fs.Parse(subArgs); err != nil {
			return err
		}

		data, code, err := c.doRequest(http.MethodGet, "/api/v1/cluster/replication", nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}

		var st struct {
			NodeID     string `json:"nodeId"`
			Role       string `json:"role"`
			Health     string `json:"health"`
			Seq        uint64 `json:"seq"`
			ReplicaDir string `json:"replicaDir"`
			LastSync   string `json:"lastSync"`
		}
		if err := json.Unmarshal(data, &st); err != nil {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}

		fmt.Fprintln(c.Stdout, "Refweir Cluster Replication Status:")
		fmt.Fprintf(c.Stdout, "  Node ID:     %s\n", st.NodeID)
		fmt.Fprintf(c.Stdout, "  Role:        %s\n", st.Role)
		fmt.Fprintf(c.Stdout, "  Health:      %s\n", st.Health)
		if st.Seq > 0 {
			fmt.Fprintf(c.Stdout, "  Sequence:    %d\n", st.Seq)
		}
		if st.ReplicaDir != "" {
			fmt.Fprintf(c.Stdout, "  Replica Dir: %s\n", st.ReplicaDir)
		}
		if st.LastSync != "" {
			fmt.Fprintf(c.Stdout, "  Last Sync:   %s\n", st.LastSync)
		}
		return nil

	case "checkpoint":
		fs := flag.NewFlagSet("cluster checkpoint", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		jsonOut := fs.Bool("json", false, "Output in JSON format")
		if err := fs.Parse(subArgs); err != nil {
			return err
		}

		data, code, err := c.doRequest(http.MethodPost, "/api/v1/cluster/checkpoint", nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		fmt.Fprintln(c.Stdout, "SQLite WAL checkpointed successfully.")
		return nil

	case "snapshot":
		fs := flag.NewFlagSet("cluster snapshot", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		jsonOut := fs.Bool("json", false, "Output in JSON format")
		if err := fs.Parse(subArgs); err != nil {
			return err
		}

		data, code, err := c.doRequest(http.MethodPost, "/api/v1/cluster/snapshot", nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}

		var meta struct {
			Filename       string `json:"filename"`
			SizeBytes      int64  `json:"sizeBytes"`
			ChecksumSHA256 string `json:"checksumSha256"`
			SnapshotID     string `json:"snapshotId"`
		}
		_ = json.Unmarshal(data, &meta)
		name := meta.Filename
		if name == "" {
			name = meta.SnapshotID
		}
		fmt.Fprintf(c.Stdout, "Created atomic point-in-time snapshot: %s (%d bytes)\n", name, meta.SizeBytes)
		if meta.ChecksumSHA256 != "" {
			fmt.Fprintf(c.Stdout, "  SHA-256: %s\n", meta.ChecksumSHA256)
		}
		return nil

	case "promote":
		fs := flag.NewFlagSet("cluster promote", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		yes := fs.Bool("yes", false, "Confirm promotion without interactive prompt")
		yShort := fs.Bool("y", false, "Confirm promotion without interactive prompt (shorthand)")
		jsonOut := fs.Bool("json", false, "Output in JSON format")
		if err := fs.Parse(subArgs); err != nil {
			return err
		}

		if !*yes && !*yShort {
			return errors.New("confirmation required: pass --yes to promote this node to primary cluster role")
		}

		data, code, err := c.doRequest(http.MethodPost, "/api/v1/cluster/failover/promote", nil)
		if err != nil {
			return err
		}
		if code >= 400 {
			return fmt.Errorf("server error (%d): %s", code, string(data))
		}
		if *jsonOut {
			fmt.Fprintln(c.Stdout, string(data))
			return nil
		}
		fmt.Fprintln(c.Stdout, "Standby replica successfully promoted to primary.")
		return nil

	default:
		return fmt.Errorf("unknown cluster subcommand %q. Run 'refweir cluster' for usage", sub)
	}
}

func parseEd25519PrivateKey(raw []byte) (ed25519.PrivateKey, error) {
	trimmed := bytes.TrimSpace(raw)
	// 1. PEM decode check
	if block, _ := pem.Decode(trimmed); block != nil {
		if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			if edKey, ok := key.(ed25519.PrivateKey); ok {
				return edKey, nil
			}
		}
		trimmed = block.Bytes
	}

	// 2. Raw 64-byte private key
	if len(trimmed) == ed25519.PrivateKeySize {
		return ed25519.PrivateKey(trimmed), nil
	}
	// 3. Raw 32-byte seed
	if len(trimmed) == ed25519.SeedSize {
		return ed25519.NewKeyFromSeed(trimmed), nil
	}

	str := string(trimmed)
	// 4. Hex encoded (128 chars for 64-byte key or 64 chars for 32-byte seed)
	if len(str) == ed25519.PrivateKeySize*2 {
		if b, err := hex.DecodeString(str); err == nil && len(b) == ed25519.PrivateKeySize {
			return ed25519.PrivateKey(b), nil
		}
	}
	if len(str) == ed25519.SeedSize*2 {
		if b, err := hex.DecodeString(str); err == nil && len(b) == ed25519.SeedSize {
			return ed25519.NewKeyFromSeed(b), nil
		}
	}

	// 5. Base64 RawURL / Std
	for _, enc := range []*base64.Encoding{base64.RawURLEncoding, base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding} {
		if b, err := enc.DecodeString(str); err == nil {
			if len(b) == ed25519.PrivateKeySize {
				return ed25519.PrivateKey(b), nil
			}
			if len(b) == ed25519.SeedSize {
				return ed25519.NewKeyFromSeed(b), nil
			}
		}
	}

	return nil, errors.New("unsupported ed25519 private key format (expected 32-byte seed or 64-byte key in raw, hex, base64, or PKCS#8 PEM)")
}

func parseEd25519PublicKey(raw []byte) (ed25519.PublicKey, error) {
	trimmed := bytes.TrimSpace(raw)
	if block, _ := pem.Decode(trimmed); block != nil {
		if pub, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
			if edPub, ok := pub.(ed25519.PublicKey); ok {
				return edPub, nil
			}
		}
		trimmed = block.Bytes
	}
	if len(trimmed) == ed25519.PublicKeySize {
		return ed25519.PublicKey(trimmed), nil
	}
	str := string(trimmed)
	if len(str) == ed25519.PublicKeySize*2 {
		if b, err := hex.DecodeString(str); err == nil && len(b) == ed25519.PublicKeySize {
			return ed25519.PublicKey(b), nil
		}
	}
	for _, enc := range []*base64.Encoding{base64.RawURLEncoding, base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding} {
		if b, err := enc.DecodeString(str); err == nil && len(b) == ed25519.PublicKeySize {
			return ed25519.PublicKey(b), nil
		}
	}
	return nil, errors.New("unsupported ed25519 public key format (expected 32-byte key in raw, hex, base64, or PKIX PEM)")
}
