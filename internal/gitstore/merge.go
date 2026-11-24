package gitstore

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/object"
	"github.com/go-git/go-git/v6/utils/merkletrie"
)

// ErrMergeConflict is returned when a merge cannot be completed cleanly.
var ErrMergeConflict = errors.New("gitstore: merge conflict")

// PRFileChange represents a file modified in a pull request diff.
type PRFileChange struct {
	Filename         string `json:"filename"`
	PreviousFilename string `json:"previousFilename,omitempty"`
	Status           string `json:"status"` // added | removed | modified | renamed
	Additions        int    `json:"additions"`
	Deletions        int    `json:"deletions"`
	Changes          int    `json:"changes"`
	Patch            string `json:"patch,omitempty"`
}

// PRDiffStats contains the total line and file change counts and file list.
type PRDiffStats struct {
	Additions    int
	Deletions    int
	ChangedFiles int
	Files        []PRFileChange
}

// DiffCommits computes the differences between baseSHA and headSHA.
func (s *Store) DiffCommits(namespaceSlug, repoName, baseSHA, headSHA string) (*PRDiffStats, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, err
	}

	baseCommit, err := repo.CommitObject(plumbing.NewHash(baseSHA))
	if err != nil {
		return nil, fmt.Errorf("gitstore: base commit %s: %w", baseSHA, err)
	}
	headCommit, err := repo.CommitObject(plumbing.NewHash(headSHA))
	if err != nil {
		return nil, fmt.Errorf("gitstore: head commit %s: %w", headSHA, err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, err
	}
	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, err
	}

	changes, err := baseTree.Diff(headTree)
	if err != nil {
		return nil, err
	}

	stats := &PRDiffStats{}
	for _, ch := range changes {
		action, err := ch.Action()
		if err != nil {
			continue
		}
		from, to, err := ch.Files()
		if err != nil {
			continue
		}

		var fc PRFileChange
		switch action {
		case merkletrie.Insert:
			fc.Status = "added"
			if to != nil {
				fc.Filename = to.Name
			}
		case merkletrie.Delete:
			fc.Status = "removed"
			if from != nil {
				fc.Filename = from.Name
			}
		case merkletrie.Modify:
			fc.Status = "modified"
			if to != nil {
				fc.Filename = to.Name
			} else if from != nil {
				fc.Filename = from.Name
			}
			if from != nil && to != nil && from.Name != to.Name {
				fc.Status = "renamed"
				fc.PreviousFilename = from.Name
			}
		}

		if patch, err := ch.Patch(); err == nil && patch != nil {
			fc.Patch = patch.String()
			fileStats := patch.Stats()
			for _, fs := range fileStats {
				fc.Additions += fs.Addition
				fc.Deletions += fs.Deletion
			}
			fc.Changes = fc.Additions + fc.Deletions
		}

		stats.Additions += fc.Additions
		stats.Deletions += fc.Deletions
		stats.Files = append(stats.Files, fc)
	}
	stats.ChangedFiles = len(stats.Files)
	return stats, nil
}

// ListPRCommits returns the commits belonging to the PR (reachable from headSHA
// but not reachable from baseSHA), newest first.
func (s *Store) ListPRCommits(namespaceSlug, repoName, baseSHA, headSHA string, limit int) ([]CommitInfo, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return nil, err
	}

	headH := plumbing.NewHash(headSHA)
	baseH := plumbing.NewHash(baseSHA)

	if limit <= 0 {
		limit = 100
	}

	visited := make(map[plumbing.Hash]bool)
	// Mark base and all ancestors as visited/stop
	if baseCommit, err := repo.CommitObject(baseH); err == nil {
		_ = walkAncestors(repo, baseCommit, visited)
	}

	var commits []CommitInfo
	curr, err := repo.CommitObject(headH)
	if err != nil {
		return nil, fmt.Errorf("gitstore: head commit %s: %w", headSHA, err)
	}

	queue := []*object.Commit{curr}
	for len(queue) > 0 && len(commits) < limit {
		c := queue[0]
		queue = queue[1:]

		if visited[c.Hash] {
			continue
		}
		visited[c.Hash] = true

		parents := make([]string, 0, len(c.ParentHashes))
		for _, p := range c.ParentHashes {
			parents = append(parents, p.String())
			if !visited[p] {
				if parentCommit, err := repo.CommitObject(p); err == nil {
					queue = append(queue, parentCommit)
				}
			}
		}

		commits = append(commits, CommitInfo{
			SHA:            c.Hash.String(),
			TreeSHA:        c.TreeHash.String(),
			Message:        c.Message,
			Parents:        parents,
			AuthorName:     c.Author.Name,
			AuthorEmail:    c.Author.Email,
			AuthorDate:     c.Author.When.Format("2006-01-02T15:04:05Z07:00"),
			CommitterName:  c.Committer.Name,
			CommitterEmail: c.Committer.Email,
			CommitterDate:  c.Committer.When.Format("2006-01-02T15:04:05Z07:00"),
			Verification:   VerifyCommitSignature(c, nil),
		})
	}

	return commits, nil
}

func walkAncestors(repo *git.Repository, c *object.Commit, visited map[plumbing.Hash]bool) error {
	if visited[c.Hash] {
		return nil
	}
	visited[c.Hash] = true
	for _, p := range c.ParentHashes {
		if !visited[p] {
			if parent, err := repo.CommitObject(p); err == nil {
				_ = walkAncestors(repo, parent, visited)
			}
		}
	}
	return nil
}

// CheckMergeable evaluates whether baseRef and headRef can be merged cleanly
// without conflict. It returns the resulting tree SHA if clean, or conflicts if blocked.
func (s *Store) CheckMergeable(namespaceSlug, repoName, baseRef, headRef string) (mergeable bool, treeSHA string, conflicts []string, err error) {
	repoPath := s.repoPath(namespaceSlug, repoName)

	baseSHA, err := s.ResolveRef(namespaceSlug, repoName, baseRef)
	if err != nil {
		return false, "", nil, fmt.Errorf("resolve base ref %s: %w", baseRef, err)
	}
	headSHA, err := s.ResolveRef(namespaceSlug, repoName, headRef)
	if err != nil {
		return false, "", nil, fmt.Errorf("resolve head ref %s: %w", headRef, err)
	}

	cmd := exec.Command("git", "-C", repoPath, "merge-tree", "--write-tree", baseSHA.String(), headSHA.String()) // #nosec G204 -- git subprocess invoked with repository-scoped, validated paths
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	output := stdout.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if runErr == nil && len(lines) > 0 && (len(lines[0]) == 40 || len(lines[0]) == 64) {
		return true, lines[0], nil, nil
	}

	var conflictList []string
	for _, l := range lines {
		if strings.HasPrefix(l, "CONFLICT") {
			conflictList = append(conflictList, strings.TrimSpace(l))
		}
	}
	if len(conflictList) == 0 && runErr != nil {
		conflictList = append(conflictList, "Merge conflict detected")
	}

	return false, "", conflictList, nil
}

// Merge performs a 3-way merge or squash merge of headRef into baseRef and advances
// baseRef to the newly created commit.
func (s *Store) Merge(namespaceSlug, repoName, baseRef, headRef, mergeMethod string, author object.Signature, message string) (string, error) {
	repoLock := s.RepoLock(namespaceSlug, repoName)
	repoLock.Lock()
	defer repoLock.Unlock()

	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return "", err
	}

	baseSHA, err := s.ResolveRef(namespaceSlug, repoName, baseRef)
	if err != nil {
		return "", fmt.Errorf("resolve base %s: %w", baseRef, err)
	}
	headSHA, err := s.ResolveRef(namespaceSlug, repoName, headRef)
	if err != nil {
		return "", fmt.Errorf("resolve head %s: %w", headRef, err)
	}

	mergeable, treeSHA, conflicts, err := s.CheckMergeable(namespaceSlug, repoName, baseRef, headRef)
	if err != nil {
		return "", err
	}
	if !mergeable {
		return "", fmt.Errorf("%w: %s", ErrMergeConflict, strings.Join(conflicts, "; "))
	}

	treeHash := plumbing.NewHash(treeSHA)
	var parents []plumbing.Hash
	if mergeMethod == "squash" {
		parents = []plumbing.Hash{baseSHA}
	} else {
		parents = []plumbing.Hash{baseSHA, headSHA}
	}

	commitHash, err := CommitObject(repo, message, author, author, treeHash, parents)
	if err != nil {
		return "", fmt.Errorf("create merge commit: %w", err)
	}

	// CAS validation: verify base ref has not moved concurrently
	currentBaseSHA, err := s.ResolveRef(namespaceSlug, repoName, baseRef)
	if err != nil {
		return "", fmt.Errorf("verify base %s before update: %w", baseRef, err)
	}
	if currentBaseSHA != baseSHA {
		return "", fmt.Errorf("base branch %s was modified concurrently (expected %s, got %s)", baseRef, baseSHA, currentBaseSHA)
	}

	// Update base branch ref
	refName := plumbing.ReferenceName(normalizeBranchRef(baseRef))
	newRef := plumbing.NewHashReference(refName, commitHash)
	oldRef := plumbing.NewHashReference(refName, baseSHA)
	if err := repo.Storer.CheckAndSetReference(newRef, oldRef); err != nil {
		// Fallback to SetReference if CheckAndSetReference is not supported
		if err := repo.Storer.SetReference(newRef); err != nil {
			return "", fmt.Errorf("update branch %s: %w", baseRef, err)
		}
	}

	return commitHash.String(), nil
}

// PrepareMergeRef creates a preview merge commit and updates refs/pull/<number>/merge.
func (s *Store) PrepareMergeRef(namespaceSlug, repoName, baseRef, headRef string, pullNumber int, author object.Signature) (string, error) {
	repo, err := s.Open(namespaceSlug, repoName)
	if err != nil {
		return "", err
	}

	baseSHA, err := s.ResolveRef(namespaceSlug, repoName, baseRef)
	if err != nil {
		return "", fmt.Errorf("resolve base %s: %w", baseRef, err)
	}
	headSHA, err := s.ResolveRef(namespaceSlug, repoName, headRef)
	if err != nil {
		return "", fmt.Errorf("resolve head %s: %w", headRef, err)
	}

	mergeable, treeSHA, conflicts, err := s.CheckMergeable(namespaceSlug, repoName, baseRef, headRef)
	if err != nil {
		return "", err
	}
	if !mergeable {
		return "", fmt.Errorf("%w: %s", ErrMergeConflict, strings.Join(conflicts, "; "))
	}

	msg := fmt.Sprintf("Merge %s into %s (preview for PR #%d)", headSHA.String()[:8], baseRef, pullNumber)
	treeHash := plumbing.NewHash(treeSHA)
	parents := []plumbing.Hash{baseSHA, headSHA}

	commitHash, err := CommitObject(repo, msg, author, author, treeHash, parents)
	if err != nil {
		return "", fmt.Errorf("create preview merge commit: %w", err)
	}

	previewRef := plumbing.ReferenceName(fmt.Sprintf("refs/pull/%d/merge", pullNumber))
	if err := repo.Storer.SetReference(plumbing.NewHashReference(previewRef, commitHash)); err != nil {
		return "", fmt.Errorf("set preview ref: %w", err)
	}

	return commitHash.String(), nil
}

// SyncForkResult reports the outcome of synchronizing a fork branch from upstream.
type SyncForkResult struct {
	MergeType  string `json:"merge_type"` // "none" | "fast-forward" | "merge"
	Message    string `json:"message"`
	BaseBranch string `json:"base_branch"`
	CommitSHA  string `json:"commit_sha,omitempty"`
}

// SyncFork synchronizes a branch in a forked repository from its upstream parent repository.
func (s *Store) SyncFork(upstreamNS, upstreamName, forkNS, forkName, branch string, author object.Signature) (*SyncForkResult, error) {
	if branch == "" {
		branch = "main"
	}

	srcKey := upstreamNS + "/" + upstreamName
	targetKey := forkNS + "/" + forkName

	srcLock := s.RepoLock(upstreamNS, upstreamName)
	targetLock := s.RepoLock(forkNS, forkName)

	if srcKey < targetKey {
		srcLock.RLock()
		defer srcLock.RUnlock()
		targetLock.Lock()
		defer targetLock.Unlock()
	} else {
		targetLock.Lock()
		defer targetLock.Unlock()
		srcLock.RLock()
		defer srcLock.RUnlock()
	}

	srcPath := s.repoPath(upstreamNS, upstreamName)
	forkPath := s.repoPath(forkNS, forkName)
	if srcPath == "" || forkPath == "" {
		return nil, fmt.Errorf("gitstore: invalid repository slugs")
	}

	// 1. Fetch upstream branch into the fork repository's remotes
	refSpec := fmt.Sprintf("%s:refs/remotes/upstream/%s", branch, branch)
	cmd := exec.Command("git", "-C", forkPath, "fetch", srcPath, refSpec) // #nosec G204 -- paths and branch are validated
	var fetchErrBuf bytes.Buffer
	cmd.Stderr = &fetchErrBuf
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("gitstore: fetch upstream branch %s: %w: %s", branch, err, fetchErrBuf.String())
	}

	forkRepo, err := s.Open(forkNS, forkName)
	if err != nil {
		return nil, err
	}

	upstreamRefName := plumbing.ReferenceName("refs/remotes/upstream/" + branch)
	upstreamRef, err := forkRepo.Reference(upstreamRefName, true)
	if err != nil {
		return nil, fmt.Errorf("gitstore: resolve upstream ref: %w", err)
	}
	upstreamCommitHash := upstreamRef.Hash()

	forkBranchRefName := plumbing.ReferenceName("refs/heads/" + branch)
	forkRef, err := forkRepo.Reference(forkBranchRefName, true)
	if err != nil {
		// If branch doesn't exist in fork yet, create it pointing to upstream
		newRef := plumbing.NewHashReference(forkBranchRefName, upstreamCommitHash)
		if err := forkRepo.Storer.SetReference(newRef); err != nil {
			return nil, fmt.Errorf("gitstore: set branch ref: %w", err)
		}
		return &SyncForkResult{
			MergeType:  "fast-forward",
			Message:    "Successfully created branch from upstream",
			BaseBranch: fmt.Sprintf("upstream:%s", branch),
			CommitSHA:  upstreamCommitHash.String(),
		}, nil
	}

	forkCommitHash := forkRef.Hash()
	if forkCommitHash == upstreamCommitHash {
		return &SyncForkResult{
			MergeType:  "none",
			Message:    "This branch is already up to date with the upstream repository",
			BaseBranch: fmt.Sprintf("upstream:%s", branch),
			CommitSHA:  forkCommitHash.String(),
		}, nil
	}

	// Check if fork commit is ancestor of upstream commit (fast-forward candidate)
	forkCommit, err := forkRepo.CommitObject(forkCommitHash)
	if err != nil {
		return nil, err
	}
	upstreamCommit, err := forkRepo.CommitObject(upstreamCommitHash)
	if err != nil {
		return nil, err
	}

	isAncestor, err := forkCommit.IsAncestor(upstreamCommit)
	if err == nil && isAncestor {
		// Fast-forward
		newRef := plumbing.NewHashReference(forkBranchRefName, upstreamCommitHash)
		if err := forkRepo.Storer.SetReference(newRef); err != nil {
			return nil, fmt.Errorf("gitstore: advance ref: %w", err)
		}
		return &SyncForkResult{
			MergeType:  "fast-forward",
			Message:    "Successfully fast-forwarded to upstream",
			BaseBranch: fmt.Sprintf("upstream:%s", branch),
			CommitSHA:  upstreamCommitHash.String(),
		}, nil
	}

	// If not fast-forward, attempt 3-way merge
	mergeable, treeSHA, conflicts, err := s.CheckMergeable(forkNS, forkName, "refs/heads/"+branch, "refs/remotes/upstream/"+branch)
	if err != nil {
		return nil, err
	}
	if !mergeable {
		return nil, fmt.Errorf("%w: %s", ErrMergeConflict, strings.Join(conflicts, "; "))
	}

	treeHash := plumbing.NewHash(treeSHA)
	msg := fmt.Sprintf("Merge upstream %s/%s branch '%s' into %s", upstreamNS, upstreamName, branch, branch)
	mergeCommitHash, err := CommitObject(forkRepo, msg, author, author, treeHash, []plumbing.Hash{forkCommitHash, upstreamCommitHash})
	if err != nil {
		return nil, fmt.Errorf("gitstore: commit merge: %w", err)
	}

	newRef := plumbing.NewHashReference(forkBranchRefName, mergeCommitHash)
	if err := forkRepo.Storer.SetReference(newRef); err != nil {
		return nil, fmt.Errorf("gitstore: update branch ref after merge: %w", err)
	}

	return &SyncForkResult{
		MergeType:  "merge",
		Message:    "Successfully merged upstream into fork",
		BaseBranch: fmt.Sprintf("upstream:%s", branch),
		CommitSHA:  mergeCommitHash.String(),
	}, nil
}
