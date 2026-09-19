package db

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"sort"

	migrations "github.com/StellarAsset-Lab/nexus-app/db/migrations"
)

const createMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	filename    TEXT PRIMARY KEY,
	applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
)`

// migrationLockKey is an arbitrary constant used as a Postgres advisory lock
// key so concurrent Migrate() callers (multiple service replicas starting
// at once, or — as this repository's own test suite discovered — parallel
// `go test` processes against a fresh database) serialize instead of
// racing on CREATE TABLE/CREATE TYPE DDL, which Postgres does not make safe
// under full concurrency.
const migrationLockKey = 72_727_272

// Migrate applies every embedded *.sql migration that hasn't already been
// recorded in schema_migrations, in filename order, each inside its own
// transaction. It never partially applies a migration: a failure rolls back
// that migration and returns immediately, leaving the schema at the last
// successfully applied migration.
//
// The whole run is serialized against other concurrent Migrate() callers
// via a session-level Postgres advisory lock, held on one dedicated
// connection for the duration.
func Migrate(ctx context.Context, conn *sql.DB) error {
	session, err := conn.Conn(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer session.Close()

	if _, err := session.ExecContext(ctx, "SELECT pg_advisory_lock($1)", migrationLockKey); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = session.ExecContext(context.Background(), "SELECT pg_advisory_unlock($1)", migrationLockKey)
	}()

	if _, err := session.ExecContext(ctx, createMigrationsTable); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	applied, err := appliedMigrations(ctx, session)
	if err != nil {
		return fmt.Errorf("load applied migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read embedded migrations: %w", err)
	}

	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			filenames = append(filenames, entry.Name())
		}
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		if applied[filename] {
			continue
		}

		contents, err := migrations.FS.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		if err := applyMigration(ctx, session, filename, string(contents)); err != nil {
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}
	}

	return nil
}

func appliedMigrations(ctx context.Context, session *sql.Conn) (map[string]bool, error) {
	rows, err := session.QueryContext(ctx, "SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return nil, err
		}
		applied[filename] = true
	}
	return applied, rows.Err()
}

func applyMigration(ctx context.Context, session *sql.Conn, filename, sqlText string) error {
	tx, err := session.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, sqlText); err != nil {
		return fmt.Errorf("execute migration sql: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (filename) VALUES ($1)", filename); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit()
}
