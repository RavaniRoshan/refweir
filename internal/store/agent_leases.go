package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// RevokeAgentLease records a revoked agent lease token.
func (db *DB) RevokeAgentLease(ctx context.Context, leaseID, agentID, repoID, reason string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	query := `
		INSERT INTO revoked_agent_leases (lease_id, agent_id, repo_id, revoked_at, reason)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(lease_id) DO UPDATE SET
			revoked_at = excluded.revoked_at,
			reason = excluded.reason
	`
	_, err := db.sql.ExecContext(ctx, query, leaseID, agentID, repoID, now, reason)
	if err != nil {
		return fmt.Errorf("revoke agent lease: %w", err)
	}
	return nil
}

// RevokeRepoAgentLeases marks all active capability leases for a specific repository as revoked.
func (db *DB) RevokeRepoAgentLeases(ctx context.Context, repoID, reason string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	markerID := fmt.Sprintf("repo:%s", repoID)
	query := `
		INSERT INTO revoked_agent_leases (lease_id, agent_id, repo_id, revoked_at, reason)
		VALUES (?, '*', ?, ?, ?)
		ON CONFLICT(lease_id) DO UPDATE SET
			revoked_at = excluded.revoked_at,
			reason = excluded.reason
	`
	_, err := db.sql.ExecContext(ctx, query, markerID, repoID, now, reason)
	if err != nil {
		return fmt.Errorf("revoke repo agent leases: %w", err)
	}
	return nil
}

// IsLeaseRevoked checks whether a specific lease ID or its target repository has been revoked.
func (db *DB) IsLeaseRevoked(ctx context.Context, leaseID, repoID string) (bool, error) {
	markerID := fmt.Sprintf("repo:%s", repoID)
	query := `SELECT 1 FROM revoked_agent_leases WHERE lease_id = ? OR lease_id = ? LIMIT 1`
	var exists int
	err := db.sql.QueryRowContext(ctx, query, leaseID, markerID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("check revoked lease: %w", err)
	}
	return true, nil
}

// IsAgentLeaseRevoked checks whether a lease ID has been marked as revoked.
func (db *DB) IsAgentLeaseRevoked(ctx context.Context, leaseID string) (bool, error) {
	return db.IsLeaseRevoked(ctx, leaseID, "")
}
