// Package forge provides centralized coordination for repository lifecycles,
// forks, branch synchronization, and change proposals for refweir.dev.
package forge

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-git/go-git/v6/plumbing/object"
	"refweir.dev/internal/audit"
	"refweir.dev/internal/events"
	"refweir.dev/internal/gitstore"
	"refweir.dev/internal/store"
	"refweir.dev/pkg/idgen"
	"refweir.dev/pkg/slug"
)

var (
	// ErrRepoNotFound indicates the target repository could not be located in metadata storage.
	ErrRepoNotFound = errors.New("repository not found")

	// ErrRepoAlreadyExists indicates a repository with the requested namespace and name exists.
	ErrRepoAlreadyExists = errors.New("repository already exists")

	// ErrInvalidName indicates the provided repository name or namespace violates naming rules.
	ErrInvalidName = errors.New("invalid repository or owner name")

	// ErrCannotSelfFork indicates an attempt to fork a repository into the exact same owner and name.
	ErrCannotSelfFork = errors.New("cannot fork repository into itself")

	// ErrNotAFork indicates the repository is not configured as a downstream fork.
	ErrNotAFork = errors.New("repository is not a fork")

	// ErrParentNotFound indicates the upstream parent repository could not be located.
	ErrParentNotFound = errors.New("upstream parent repository not found")

	// ErrMergeConflict indicates the merge operation could not be completed cleanly.
	ErrMergeConflict = gitstore.ErrMergeConflict
)

// Config holds runtime configuration options for the forge engine.
type Config struct {
	PublicURL string
}

// Engine coordinates Git disk operations, relational metadata transactions,
// audit trails, and event dispatches behind a single deep interface.
type Engine struct {
	db  *store.DB
	git *gitstore.Store
	bus *events.Bus
	cfg Config
}

// New creates an initialized forge engine.
func New(db *store.DB, git *gitstore.Store, bus *events.Bus, cfg Config) *Engine {
	if cfg.PublicURL == "" {
		cfg.PublicURL = "http://localhost:8080"
	}
	if bus == nil {
		bus = events.NewBus(db, nil)
	}
	return &Engine{
		db:  db,
		git: git,
		bus: bus,
		cfg: cfg,
	}
}

// CreateRepoRequest encapsulates parameters for initializing a new repository.
type CreateRepoRequest struct {
	Owner               string `json:"owner"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	DefaultBranch       string `json:"default_branch"`
	Visibility          string `json:"visibility"`
	InitReadme          bool   `json:"init_readme"`
	GitignoreTemplate   string `json:"gitignore_template"`
	LicenseTemplate     string `json:"license_template"`
	AllowMergeCommit    *bool  `json:"allow_merge_commit,omitempty"`
	AllowSquashMerge    *bool  `json:"allow_squash_merge,omitempty"`
	DeleteBranchOnMerge *bool  `json:"delete_branch_on_merge,omitempty"`
	Actor               string `json:"actor"`
	ActorType           string `json:"actor_type"`
	CallerUserID        string `json:"caller_user_id,omitempty"`
}

// CreateRepository coordinates namespace creation, bare Git repository initialization,
// database persistence, audit logging, and webhook event dispatching.
func (e *Engine) CreateRepository(ctx context.Context, req CreateRepoRequest) (*store.Repo, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: name is required", ErrInvalidName)
	}
	owner := strings.TrimSpace(req.Owner)
	if owner == "" {
		owner = "refweir"
	}

	if err := slug.Validate(owner); err != nil {
		return nil, fmt.Errorf("%w: invalid owner slug %q: %v", ErrInvalidName, owner, err)
	}
	if err := slug.Validate(name); err != nil {
		return nil, fmt.Errorf("%w: invalid repo name %q: %v", ErrInvalidName, name, err)
	}

	// Check for existing repository
	if _, err := e.db.GetRepo(ctx, owner, name); err == nil {
		return nil, fmt.Errorf("%w: %s/%s", ErrRepoAlreadyExists, owner, name)
	}

	branch := strings.TrimSpace(req.DefaultBranch)
	if branch == "" {
		branch = "main"
	}

	visibility := strings.TrimSpace(req.Visibility)
	if visibility == "" {
		visibility = "public"
	}

	// Ensure namespace exists
	ns, err := e.db.GetNamespaceBySlug(ctx, owner)
	if err != nil {
		ns = store.Namespace{
			ID:          idgen.NewNS(),
			Slug:        owner,
			Name:        owner,
			Type:        "user",
			OwnerUserID: req.CallerUserID,
		}
		_ = e.db.CreateNamespace(ctx, ns)
	}

	// Initialize bare repo on disk
	if _, err := e.git.Init(owner, name, branch); err != nil {
		return nil, fmt.Errorf("init bare repo on disk: %w", err)
	}

	// Build public clone and web URLs
	publicBase := strings.TrimRight(e.cfg.PublicURL, "/")
	cloneHTTP := fmt.Sprintf("%s/%s/%s.git", publicBase, owner, name)
	webURL := fmt.Sprintf("%s/%s/%s", publicBase, owner, name)

	allowMerge := true
	if req.AllowMergeCommit != nil {
		allowMerge = *req.AllowMergeCommit
	}
	allowSquash := true
	if req.AllowSquashMerge != nil {
		allowSquash = *req.AllowSquashMerge
	}
	deleteOnMerge := false
	if req.DeleteBranchOnMerge != nil {
		deleteOnMerge = *req.DeleteBranchOnMerge
	}

	repoRecord := store.Repo{
		ID:                  idgen.NewRepo(),
		NamespaceID:         ns.ID,
		Name:                name,
		FullName:            owner + "/" + name,
		Description:         req.Description,
		DefaultBranch:       branch,
		Visibility:          visibility,
		CloneURL:            cloneHTTP,
		WebURL:              webURL,
		AllowMergeCommit:    allowMerge,
		AllowSquashMerge:    allowSquash,
		DeleteBranchOnMerge: deleteOnMerge,
	}

	if err := e.db.CreateRepo(ctx, repoRecord); err != nil {
		return nil, fmt.Errorf("record repo: %w", err)
	}

	if req.CallerUserID != "" {
		_ = e.db.EnsureNamespaceAdminGrant(ctx, ns.ID, req.CallerUserID)
	}

	actor := req.Actor
	if actor == "" {
		actor = "api"
	}
	actorType := req.ActorType
	if actorType == "" {
		actorType = "user"
	}

	e.bus.Publish(ctx, events.DomainEvent{
		RepoID: repoRecord.ID,
		Event:  "repository",
		Payload: map[string]any{
			"action":     "created",
			"repository": repoRecord,
			"sender": map[string]any{
				"login": actor,
			},
		},
		Actor:     actor,
		ActorType: actorType,
		Action:    "repo.create",
		Resource:  repoRecord.FullName,
		Severity:  audit.SeverityInfo,
		Details:   fmt.Sprintf("Created repository %s", repoRecord.FullName),
	})

	return &repoRecord, nil
}

// ForkRepoRequest encapsulates parameters for creating a downstream repository fork.
type ForkRepoRequest struct {
	SourceOwner     string `json:"source_owner"`
	SourceName      string `json:"source_name"`
	TargetNamespace string `json:"target_namespace"`
	TargetName      string `json:"target_name"`
	Actor           string `json:"actor"`
	ActorType       string `json:"actor_type"`
	CallerUserID    string `json:"caller_user_id,omitempty"`
}

// ForkRepository coordinates source verification, self-fork prevention, namespace provisioning,
// bare Git cloning on disk, database recording, audit logging, and webhook event dispatching.
func (e *Engine) ForkRepository(ctx context.Context, req ForkRepoRequest) (*store.Repo, error) {
	sourceOwner := strings.TrimSpace(req.SourceOwner)
	sourceName := strings.TrimSpace(req.SourceName)

	sourceRepo, err := e.db.GetRepo(ctx, sourceOwner, sourceName)
	if err != nil {
		return nil, fmt.Errorf("%w: %s/%s", ErrRepoNotFound, sourceOwner, sourceName)
	}

	targetOrg := strings.TrimSpace(req.TargetNamespace)
	if targetOrg == "" {
		targetOrg = "refweir"
	}
	forkName := strings.TrimSpace(req.TargetName)
	if forkName == "" {
		forkName = sourceRepo.Name
	}

	if err := slug.Validate(targetOrg); err != nil {
		return nil, fmt.Errorf("%w: invalid target namespace %q: %v", ErrInvalidName, targetOrg, err)
	}
	if err := slug.Validate(forkName); err != nil {
		return nil, fmt.Errorf("%w: invalid fork name %q: %v", ErrInvalidName, forkName, err)
	}

	if targetOrg == sourceOwner && forkName == sourceRepo.Name {
		return nil, ErrCannotSelfFork
	}

	// Verify destination does not already exist
	if _, err := e.db.GetRepo(ctx, targetOrg, forkName); err == nil {
		return nil, fmt.Errorf("%w: %s/%s", ErrRepoAlreadyExists, targetOrg, forkName)
	}

	// Ensure target namespace exists
	targetNS, err := e.db.GetNamespaceBySlug(ctx, targetOrg)
	if err != nil {
		targetNS = store.Namespace{
			ID:          idgen.NewNS(),
			Slug:        targetOrg,
			Name:        targetOrg,
			Type:        "user",
			OwnerUserID: req.CallerUserID,
		}
		_ = e.db.CreateNamespace(ctx, targetNS)
	}

	// Perform bare git clone on disk
	if _, err := e.git.Fork(sourceOwner, sourceRepo.Name, targetOrg, forkName); err != nil {
		return nil, fmt.Errorf("fork git repository on disk: %w", err)
	}

	targetFullName := targetOrg + "/" + forkName
	publicBase := strings.TrimRight(e.cfg.PublicURL, "/")
	cloneHTTP := fmt.Sprintf("%s/%s/%s.git", publicBase, targetOrg, forkName)
	webURL := fmt.Sprintf("%s/%s/%s", publicBase, targetOrg, forkName)

	forkRecord := store.Repo{
		ID:                  idgen.NewRepo(),
		NamespaceID:         targetNS.ID,
		Name:                forkName,
		FullName:            targetFullName,
		Description:         sourceRepo.Description,
		DefaultBranch:       sourceRepo.DefaultBranch,
		Visibility:          sourceRepo.Visibility,
		CloneURL:            cloneHTTP,
		WebURL:              webURL,
		AllowMergeCommit:    sourceRepo.AllowMergeCommit,
		AllowSquashMerge:    sourceRepo.AllowSquashMerge,
		DeleteBranchOnMerge: sourceRepo.DeleteBranchOnMerge,
		ParentRepoID:        sourceRepo.ID,
		IsFork:              true,
	}

	if err := e.db.CreateRepo(ctx, forkRecord); err != nil {
		return nil, fmt.Errorf("record fork: %w", err)
	}

	if req.CallerUserID != "" {
		_ = e.db.EnsureNamespaceAdminGrant(ctx, targetNS.ID, req.CallerUserID)
	}

	actor := req.Actor
	if actor == "" {
		actor = "api"
	}
	actorType := req.ActorType
	if actorType == "" {
		actorType = "user"
	}

	e.bus.Publish(ctx, events.DomainEvent{
		RepoID: sourceRepo.ID,
		Event:  "fork",
		Payload: map[string]any{
			"action":     "forkee",
			"forkee":     forkRecord,
			"repository": sourceRepo,
			"sender": map[string]any{
				"login": actor,
			},
		},
		Actor:     actor,
		ActorType: actorType,
		Action:    "repo.fork",
		Resource:  targetFullName,
		Severity:  audit.SeverityInfo,
		Details:   fmt.Sprintf("Forked repository %s/%s to %s", sourceOwner, sourceName, targetFullName),
	})

	return &forkRecord, nil
}

// SyncForkRequest encapsulates parameters for synchronizing a fork from upstream.
type SyncForkRequest struct {
	Owner     string `json:"owner"`
	Name      string `json:"name"`
	Branch    string `json:"branch"`
	Actor     string `json:"actor"`
	ActorType string `json:"actor_type"`
	Email     string `json:"email"`
}

// SyncFork coordinates fork validation, upstream parent lookup, fast-forward or
// 3-way conflict-detecting merge, audit logging, and webhook event dispatching.
func (e *Engine) SyncFork(ctx context.Context, req SyncForkRequest) (*gitstore.SyncForkResult, error) {
	owner := strings.TrimSpace(req.Owner)
	name := strings.TrimSpace(req.Name)

	repo, err := e.db.GetRepo(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("%w: %s/%s", ErrRepoNotFound, owner, name)
	}

	if !repo.IsFork || repo.ParentRepoID == "" {
		return nil, fmt.Errorf("%w: %s/%s", ErrNotAFork, owner, name)
	}

	parentRepo, err := e.db.GetRepoByID(ctx, repo.ParentRepoID)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrParentNotFound, repo.ParentRepoID, err)
	}

	branch := strings.TrimSpace(req.Branch)
	if branch == "" {
		branch = repo.DefaultBranch
		if branch == "" {
			branch = "main"
		}
	}

	actor := req.Actor
	if actor == "" {
		actor = "api"
	}
	actorType := req.ActorType
	if actorType == "" {
		actorType = "user"
	}
	email := req.Email
	if email == "" {
		email = "noreply@refweir.local"
	}

	sig := object.Signature{
		Name:  actor,
		Email: email,
		When:  time.Now(),
	}

	parentParts := strings.Split(parentRepo.FullName, "/")
	parentNS := parentParts[0]

	res, err := e.git.SyncFork(parentNS, parentRepo.Name, owner, name, branch, sig)
	if err != nil {
		if errors.Is(err, gitstore.ErrMergeConflict) {
			return nil, ErrMergeConflict
		}
		return nil, fmt.Errorf("sync fork: %w", err)
	}

	e.bus.Publish(ctx, events.DomainEvent{
		RepoID: repo.ID,
		Event:  "fork_sync",
		Payload: map[string]any{
			"action":            "synchronized",
			"merge_type":        res.MergeType,
			"base_branch":       res.BaseBranch,
			"commit_sha":        res.CommitSHA,
			"repository":        repo,
			"parent_repository": parentRepo,
			"sender": map[string]any{
				"login": actor,
			},
		},
		Actor:     actor,
		ActorType: actorType,
		Action:    "repo.fork_sync",
		Resource:  repo.FullName,
		Severity:  audit.SeverityInfo,
		Details:   fmt.Sprintf("Synchronized fork %s from upstream %s (branch %s, type %s)", repo.FullName, parentRepo.FullName, branch, res.MergeType),
	})

	return res, nil
}

// ListForks returns child repositories forked from the specified parent.
func (e *Engine) ListForks(ctx context.Context, owner, name string, limit int) ([]store.Repo, error) {
	sourceRepo, err := e.db.GetRepo(ctx, owner, name)
	if err != nil {
		return nil, fmt.Errorf("%w: %s/%s", ErrRepoNotFound, owner, name)
	}
	if limit <= 0 {
		limit = 30
	}
	forks, err := e.db.ListForks(ctx, sourceRepo.ID, limit)
	if err != nil {
		return nil, err
	}
	if forks == nil {
		forks = []store.Repo{}
	}
	return forks, nil
}
