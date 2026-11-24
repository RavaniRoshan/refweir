-- Migration 0018: Repository Forks and Lineage Tracking
ALTER TABLE repos ADD COLUMN parent_repo_id TEXT REFERENCES repos(id) ON DELETE SET NULL;
ALTER TABLE repos ADD COLUMN is_fork INTEGER NOT NULL DEFAULT 0;
CREATE INDEX IF NOT EXISTS idx_repos_parent_repo_id ON repos(parent_repo_id);
