-- Migration 0017: Composite indexes for sorted query plans.
-- Eliminates temporary B-tree allocations on pull requests, review threads, comments, and reactions.

CREATE INDEX IF NOT EXISTS idx_prs_repo_state_number ON pull_requests(repo_id, state, number DESC);
CREATE INDEX IF NOT EXISTS idx_prs_repo_number ON pull_requests(repo_id, number DESC);
CREATE INDEX IF NOT EXISTS idx_pr_comments_pr_created ON pr_comments(pull_request_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_pr_created ON pr_reviews(pull_request_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_reactions_comment_created ON pr_comment_reactions(comment_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_labels_ns_name ON labels(namespace_id, name ASC);
