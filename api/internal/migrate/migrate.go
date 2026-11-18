// Package migrate applies embedded SQL migrations at startup.
//
// Rules:
//   - files run in filename order (0001_..., 0002_...), each once
//   - each file runs in its own transaction: it fully applies or fully rolls back
//   - a Postgres advisory lock stops two API instances migrating at the same time
//   - applied files are recorded with a checksum; editing one later is an error
package migrate

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"log"
	"sort"
)

const advisoryLockID = 727274 // arbitrary app-wide constant

func Up(ctx context.Context, db *sql.DB, files fs.FS) error {
	// The advisory lock is per session, so hold it on one dedicated connection.
	lockConn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("migrate: get connection: %w", err)
	}
	defer lockConn.Close()

	if _, err := lockConn.ExecContext(ctx, `SELECT pg_advisory_lock($1)`, advisoryLockID); err != nil {
		return fmt.Errorf("migrate: take lock: %w", err)
	}
	defer lockConn.ExecContext(context.Background(), `SELECT pg_advisory_unlock($1)`, advisoryLockID) //nolint:errcheck

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    text PRIMARY KEY,
			checksum   text NOT NULL,
			applied_at timestamptz NOT NULL DEFAULT now()
		)`); err != nil {
		return fmt.Errorf("migrate: create schema_migrations: %w", err)
	}

	applied, err := loadApplied(ctx, db)
	if err != nil {
		return err
	}

	names, err := fs.Glob(files, "*.sql")
	if err != nil {
		return fmt.Errorf("migrate: list files: %w", err)
	}
	sort.Strings(names)

	for _, name := range names {
		body, err := fs.ReadFile(files, name)
		if err != nil {
			return fmt.Errorf("migrate: read %s: %w", name, err)
		}
		sum := checksum(body)

		if prev, ok := applied[name]; ok {
			if prev != sum {
				return fmt.Errorf("migrate: %s was changed after it was applied; add a new migration instead", name)
			}
			continue
		}

		if err := apply(ctx, db, name, string(body), sum); err != nil {
			return err
		}
		log.Printf("migrate: applied %s", name)
	}
	return nil
}

func loadApplied(ctx context.Context, db *sql.DB) (map[string]string, error) {
	rows, err := db.QueryContext(ctx, `SELECT version, checksum FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("migrate: read schema_migrations: %w", err)
	}
	defer rows.Close()

	applied := map[string]string{}
	for rows.Next() {
		var v, c string
		if err := rows.Scan(&v, &c); err != nil {
			return nil, err
		}
		applied[v] = c
	}
	return applied, rows.Err()
}

func apply(ctx context.Context, db *sql.DB, name, body, sum string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("migrate: begin %s: %w", name, err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op after a successful Commit

	// No arguments, so the driver uses the simple protocol and accepts many statements.
	if _, err := tx.ExecContext(ctx, body); err != nil {
		return fmt.Errorf("migrate: run %s: %w", name, err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)`, name, sum); err != nil {
		return fmt.Errorf("migrate: record %s: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migrate: commit %s: %w", name, err)
	}
	return nil
}

func checksum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
