// Package store is refweir.dev's metadata store: a pure-Go SQLite database
// (modernc.org/sqlite, no cgo) holding namespaces, repos, grants, pull
// requests, checks, rulesets, apps, and every other Refweir entity that is not
// itself a git object. Git object data lives in internal/gitstore.
package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"time"

	_ "modernc.org/sqlite" // pure-Go SQLite driver, registers name "sqlite"
)

// DB wraps the metadata database and exposes the underlying *sql.DB for
// repository/query code in later phases.
type DB struct {
	sql *sql.DB
}

// Open opens (creating if necessary) the SQLite database at path, applies
// pragmas, and runs migrations. Use ":memory:" for an ephemeral database.
func Open(path string) (*DB, error) {
	// modernc's driver expects a file: DSN with pragmas for durability, performance in WAL mode, and connection-scoped foreign keys.
	dsn := "file:" + path + "?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-16000)&_pragma=foreign_keys(1)&_pragma=auto_vacuum(INCREMENTAL)&_txlock=immediate"
	if path == ":memory:" {
		dsn = "file::memory:?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_txlock=immediate"
	}

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("store: open %s: %w", path, err)
	}
	// Configure connection pooling: in-memory requires 1 connection to preserve state,
	// while on-disk WAL mode supports concurrent readers with busy_timeout queuing.
	if path == ":memory:" {
		sqlDB.SetMaxOpenConns(1)
	} else {
		maxConns := max(4, runtime.NumCPU()*2)
		sqlDB.SetMaxOpenConns(maxConns)
		sqlDB.SetMaxIdleConns(max(2, runtime.NumCPU()))
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)
	}

	if _, err := sqlDB.Exec(`PRAGMA foreign_keys = ON; PRAGMA auto_vacuum = INCREMENTAL;`); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("store: apply pragmas: %w", err)
	}

	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("store: migrate: %w", err)
	}
	return db, nil
}

// Close closes the underlying database.
func (db *DB) Close() error {
	if db == nil || db.sql == nil {
		return nil
	}
	return db.sql.Close()
}

// Ping verifies database connectivity and responsiveness.
func (db *DB) Ping(ctx context.Context) error {
	if db == nil || db.sql == nil {
		return errors.New("store: database not initialized")
	}
	return db.sql.PingContext(ctx)
}

// IncrementalVacuum reclaims unused freelist pages back to the operating system.
// If maxPages > 0, it frees up to maxPages pages; if <= 0, it frees all freelist pages.
func (db *DB) IncrementalVacuum(ctx context.Context, maxPages int) error {
	if db == nil || db.sql == nil {
		return errors.New("store: database not initialized")
	}
	query := "PRAGMA incremental_vacuum"
	if maxPages > 0 {
		query = fmt.Sprintf("PRAGMA incremental_vacuum(%d)", maxPages)
	}
	_, err := db.sql.ExecContext(ctx, query)
	return err
}

// SQL exposes the raw database for query code in later phases. Prefer typed
// helpers as they are added; this is the escape hatch.
func (db *DB) SQL() *sql.DB { return db.sql }

// Dir returns the directory that should hold git object stores for repos backed
// by this metadata DB (i.e. the parent of the sqlite file). For ":memory:" it
// returns "".
func (db *DB) Dir() string { return "" }
