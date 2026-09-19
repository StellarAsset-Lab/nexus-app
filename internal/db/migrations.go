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

// Migrate applies every embedded *.sql migration that hasn't already been
// recorded in schema_migrations, in filename order, each inside its own
// transaction. It never partially applies a migration: a failure rolls back
// that migration and returns immediately, leaving the schema at the last
// successfully applied migration.
func Migrate(ctx context.Context, conn *sql.DB) error {
	if _, err := conn.ExecContext(ctx, createMigrationsTable); err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	applied, err := appliedMigrations(ctx, conn)
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

		if err := applyMigration(ctx, conn, filename, string(contents)); err != nil {
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}
	}

	return nil
}

func appliedMigrations(ctx context.Context, conn *sql.DB) (map[string]bool, error) {
	rows, err := conn.QueryContext(ctx, "SELECT filename FROM schema_migrations")
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

func applyMigration(ctx context.Context, conn *sql.DB, filename, sqlText string) error {
	tx, err := conn.BeginTx(ctx, nil)
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
