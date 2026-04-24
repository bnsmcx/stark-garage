package memory

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// DB wraps *sql.DB with the toolbox memory schema.
type DB struct {
	sql *sql.DB
}

// Open opens (or creates) the memory database at the specified path.
// Creates the parent directory with 0700 permissions if it does not exist.
// Runs all schema DDL statements on every open (idempotent).
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create memory dir: %w", err)
	}

	// Ensure the DB file exists before opening so we can set permissions.
	f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory db file: %w", err)
	}
	f.Close()

	if err := os.Chmod(dbPath, 0600); err != nil {
		return nil, fmt.Errorf("failed to set db permissions: %w", err)
	}

	dsn := dbPath + "?_journal_mode=WAL&_busy_timeout=5000"

	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open memory db: %w", err)
	}

	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to initialize memory db: %w", err)
	}

	return db, nil
}

// OpenDefault opens the memory database at .claude/memory/toolbox.db
// relative to the current working directory.
func OpenDefault() (*DB, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	dbPath := filepath.Join(cwd, ".claude", "memory", "toolbox.db")
	return Open(dbPath)
}

// openInMemory opens an in-memory SQLite database and runs the full schema.
// Used by tests to avoid touching the filesystem.
func openInMemory() (*DB, error) {
	sqlDB, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory db: %w", err)
	}

	db := &DB{sql: sqlDB}
	if err := db.migrate(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to initialize in-memory db: %w", err)
	}

	return db, nil
}

// Close closes the underlying *sql.DB.
func (d *DB) Close() error {
	return d.sql.Close()
}

// migrate runs all schema DDL statements idempotently.
func (d *DB) migrate() error {
	ddl := []string{
		`CREATE TABLE IF NOT EXISTS memories (
			id INTEGER PRIMARY KEY,
			namespace TEXT NOT NULL,
			agent TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			confidence REAL DEFAULT 0.5,
			hit_count INTEGER DEFAULT 0,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			expires_at TEXT,
			lifecycle TEXT DEFAULT 'active'
		)`,

		`CREATE UNIQUE INDEX IF NOT EXISTS idx_memories_ns_key
			ON memories (namespace, key)`,

		`CREATE INDEX IF NOT EXISTS idx_memories_lifecycle
			ON memories (lifecycle)`,

		`CREATE INDEX IF NOT EXISTS idx_memories_ns_lifecycle
			ON memories (namespace, lifecycle)`,

		`CREATE VIRTUAL TABLE IF NOT EXISTS memories_fts
			USING fts5(
				key,
				value,
				content=memories,
				content_rowid=id
			)`,

		// FTS sync triggers: keep the FTS index in sync with the main table.

		`CREATE TRIGGER IF NOT EXISTS memories_fts_insert
			AFTER INSERT ON memories
		BEGIN
			INSERT INTO memories_fts(rowid, key, value)
				VALUES (new.id, new.key, new.value);
		END`,

		`CREATE TRIGGER IF NOT EXISTS memories_fts_update
			AFTER UPDATE ON memories
		BEGIN
			INSERT INTO memories_fts(memories_fts, rowid, key, value)
				VALUES ('delete', old.id, old.key, old.value);
			INSERT INTO memories_fts(rowid, key, value)
				VALUES (new.id, new.key, new.value);
		END`,

		`CREATE TRIGGER IF NOT EXISTS memories_fts_delete
			AFTER DELETE ON memories
		BEGIN
			INSERT INTO memories_fts(memories_fts, rowid, key, value)
				VALUES ('delete', old.id, old.key, old.value);
		END`,

		// Backfill FTS index (no-op if table is empty; handles upgrades).
		`INSERT INTO memories_fts(rowid, key, value)
			SELECT id, key, value FROM memories`,

		// Additive column for promote target. Added unconditionally; the
		// "duplicate column name" error is suppressed below for idempotency.
		`ALTER TABLE memories ADD COLUMN promoted_to TEXT`,

		// One-time backfill for rows promoted under the legacy concat scheme
		// (value was suffixed with " [promoted to: X]"). The IS NULL guard
		// makes this idempotent; the LIKE guard ensures we only touch rows
		// that match the specific corrupted shape.
		`UPDATE memories
		 SET
		   promoted_to = substr(
		     value,
		     instr(value, ' [promoted to: ') + length(' [promoted to: '),
		     length(value) - instr(value, ' [promoted to: ') - length(' [promoted to: ')
		   ),
		   value = substr(value, 1, instr(value, ' [promoted to: ') - 1)
		 WHERE lifecycle = 'promoted'
		   AND promoted_to IS NULL
		   AND value LIKE '% [promoted to: %]'`,
	}

	for _, stmt := range ddl {
		result, err := d.sql.Exec(stmt)
		if err != nil {
			// Suppress expected-idempotent errors ("already exists" for CREATE
			// statements; "duplicate column" for repeat ALTER ADD COLUMN). Log
			// the suppression so operators inspecting an Open() failure can
			// audit which DDL silently no-op'd.
			msg := err.Error()
			if strings.Contains(msg, "already exists") ||
				strings.Contains(msg, "duplicate column") {
				fmt.Fprintf(os.Stderr, "toolbox-memory: migration step is idempotent no-op: %s\n", msg)
				continue
			}
			return fmt.Errorf("migration failed: %w", err)
		}

		// Surface rows affected for the one-time legacy promote backfill so
		// operators know how many corrupted rows were repaired on upgrade.
		if strings.Contains(stmt, "promoted_to = substr") {
			if n, _ := result.RowsAffected(); n > 0 {
				fmt.Fprintf(os.Stderr, "toolbox-memory: backfilled %d legacy promoted entries (split value and promoted_to)\n", n)
			}
		}
	}

	return nil
}
