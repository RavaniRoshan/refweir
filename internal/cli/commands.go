package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// CLI is the runner for the origin CLI client.
type CLI struct {
	Endpoint    string
	Token       string
	DefaultRepo string
	HTTPClient  *http.Client
	Stdout      io.Writer
	Stderr      io.Writer
}

// New creates a new CLI instance configured from the environment.
func New() *CLI {
	endpoint := os.Getenv("REFWEIR_ENDPOINT")
	if endpoint == "" {
		endpoint = os.Getenv("CURSOR_ORIGIN_ENDPOINT")
	}
	if endpoint == "" {
		endpoint = "http://localhost:8080/v1/origin"
	}
	endpoint = strings.TrimRight(endpoint, "/")

	token := os.Getenv("REFWEIR_TOKEN")
	if token == "" {
		token = os.Getenv("CURSOR_AUTH_TOKEN")
	}
	if token == "" {
		token = os.Getenv("CURSOR_API_KEY")
	}

	defaultRepo := os.Getenv("REFWEIR_REPO")
	if defaultRepo == "" {
		defaultRepo = os.Getenv("ORIGIN_REPO")
	}

	return &CLI{
		Endpoint:    endpoint,
		Token:       token,
		DefaultRepo: defaultRepo,
		HTTPClient:  http.DefaultClient,
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
  refweir lease              Manage agent capability leases (issue, revoke)
  refweir app                Manage apps
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
	url := c.Endpoint + path
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
		fmt.Fprintln(c.Stdout, "Set CURSOR_AUTH_TOKEN or CURSOR_API_KEY in your environment.")
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
		return errors.New("usage: refweir lease <issue|revoke>")
	}
	switch args[0] {
	case "issue":
		fs := flag.NewFlagSet("lease issue", flag.ContinueOnError)
		fs.SetOutput(c.Stderr)
		agent := fs.String("agent", "agent:cli", "Agent identifier (e.g. agent:cursor, agent:claude)")
		repoFlag := fs.String("R", c.DefaultRepo, "Target repository owner/name")
		repoLong := fs.String("repo", "", "Target repository owner/name")
		prefix := fs.String("prefix", "refs/heads/agent/*", "Comma-separated allowed ref prefixes")
		ttlStr := fs.String("ttl", "30m", "TTL duration (e.g. 30m, 1h, 900s)")
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

		reqBody := map[string]any{
			"agentId":         *agent,
			"repoId":          repoTarget,
			"allowedPrefixes": prefixes,
			"ttlSeconds":      ttlSeconds,
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
