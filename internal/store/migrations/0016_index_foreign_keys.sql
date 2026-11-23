-- Migration 0016: Index foreign keys across metadata tables to eliminate full
-- table scans on cascading parent deletions and relational checks under SQLite WAL.

CREATE INDEX IF NOT EXISTS idx_namespaces_owner_user_id ON namespaces (owner_user_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_parent_pr_id ON pull_requests (parent_pr_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_author_user_id ON pull_requests (author_user_id);
CREATE INDEX IF NOT EXISTS idx_pr_comments_thread_id ON pr_comments (thread_id);
CREATE INDEX IF NOT EXISTS idx_pr_comments_in_reply_to ON pr_comments (in_reply_to);
CREATE INDEX IF NOT EXISTS idx_pr_comments_author_user_id ON pr_comments (author_user_id);
CREATE INDEX IF NOT EXISTS idx_pr_comments_repo_id ON pr_comments (repo_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_dismissed_by_user_id ON pr_reviews (dismissed_by_user_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_author_user_id ON pr_reviews (author_user_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviews_repo_id ON pr_reviews (repo_id);
CREATE INDEX IF NOT EXISTS idx_pr_labels_label_id ON pr_labels (label_id);
CREATE INDEX IF NOT EXISTS idx_check_suites_actor_user_id ON check_suites (actor_user_id);
CREATE INDEX IF NOT EXISTS idx_check_runs_actor_user_id ON check_runs (actor_user_id);
CREATE INDEX IF NOT EXISTS idx_check_runs_rerequested_by_user_id ON check_runs (rerequested_by_user_id);
CREATE INDEX IF NOT EXISTS idx_installations_namespace_id ON installations (namespace_id);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_installation_id ON webhook_deliveries (installation_id);
CREATE INDEX IF NOT EXISTS idx_access_tokens_api_key_id ON access_tokens (api_key_id);
CREATE INDEX IF NOT EXISTS idx_pr_threads_resolved_by_user_id ON pr_threads (resolved_by_user_id);
CREATE INDEX IF NOT EXISTS idx_pr_threads_repo_id ON pr_threads (repo_id);
CREATE INDEX IF NOT EXISTS idx_pr_comment_reactions_user_id ON pr_comment_reactions (user_id);
CREATE INDEX IF NOT EXISTS idx_pr_requested_reviewers_group_id ON pr_requested_reviewers (group_id);
CREATE INDEX IF NOT EXISTS idx_pr_requested_reviewers_user_id ON pr_requested_reviewers (user_id);
CREATE INDEX IF NOT EXISTS idx_pr_auto_merge_enabled_by_user_id ON pr_auto_merge (enabled_by_user_id);
CREATE INDEX IF NOT EXISTS idx_installation_repos_repo_id ON installation_repos (repo_id);
CREATE INDEX IF NOT EXISTS idx_issues_author_user_id ON issues (author_user_id);
CREATE INDEX IF NOT EXISTS idx_issue_comments_author_user_id ON issue_comments (author_user_id);
CREATE INDEX IF NOT EXISTS idx_issue_comments_repo_id ON issue_comments (repo_id);
