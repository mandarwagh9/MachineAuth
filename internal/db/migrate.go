package db

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// RunMigrations applies pending SQL migrations for PostgresDB.
// For JSON backend this is a no-op.
func RunMigrations(database Database) error {
	pdb, ok := database.(*PostgresDB)
	if !ok {
		// JSON backend — no migrations needed.
		return nil
	}
	return pdb.runMigrations()
}

func (p *PostgresDB) runMigrations() error {
	ctx := context.Background()

	// Ensure schema_migrations table exists.
	_, err := p.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			name       VARCHAR(255) NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// Get already-applied versions.
	rows, err := p.pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return fmt.Errorf("scan migration version: %w", err)
		}
		applied[v] = true
	}

	// Read embedded migration files.
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type migration struct {
		version int
		name    string
		sql     string
	}

	var pending []migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		// Extract version from filename: 001_initial.sql → 1
		parts := strings.SplitN(entry.Name(), "_", 2)
		if len(parts) < 2 {
			continue
		}
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		if applied[version] {
			continue
		}
		data, err := migrationFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}
		pending = append(pending, migration{
			version: version,
			name:    entry.Name(),
			sql:     string(data),
		})
	}

	// Sort by version.
	sort.Slice(pending, func(i, j int) bool {
		return pending[i].version < pending[j].version
	})

	// Apply each pending migration in a transaction.
	for _, m := range pending {
		if err := applyMigration(ctx, p.pool, m.version, m.name, m.sql); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.name, err)
		}
		log.Printf("applied migration %s", m.name)
	}

	return nil
}

func applyMigration(ctx context.Context, pool *pgxpool.Pool, version int, name, sql string) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, sql); err != nil {
		return fmt.Errorf("exec sql: %w", err)
	}

	if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version, name) VALUES ($1, $2)`, version, name); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit(ctx)
}
