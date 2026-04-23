package memory

import (
	"database/sql"
	"testing"
)

func TestOpenInMemory(t *testing.T) {
	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory failed: %v", err)
	}
	defer db.Close()

	// Verify schema: insert and query should work
	_, err = db.Store("test_ns", "test_agent", "k1", "v1")
	if err != nil {
		t.Fatalf("store after open failed: %v", err)
	}
}

func TestMigrateIdempotent(t *testing.T) {
	db, err := openInMemory()
	if err != nil {
		t.Fatalf("first open failed: %v", err)
	}
	// Run migrate again — should be idempotent
	if err := db.migrate(); err != nil {
		t.Fatalf("second migrate failed: %v", err)
	}
	// Third time for good measure
	if err := db.migrate(); err != nil {
		t.Fatalf("third migrate failed: %v", err)
	}
	db.Close()
}

func TestPromoteBackfill(t *testing.T) {
	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory: %v", err)
	}
	defer db.Close()

	// Seed a row with the old corrupted shape: value ends with
	// " [promoted to: X]" and promoted_to IS NULL.
	_, err = db.sql.Exec(
		`INSERT INTO memories
		   (namespace, agent, key, value, confidence, hit_count,
		    created_at, updated_at, lifecycle, promoted_to)
		 VALUES
		   ('bug_pattern', 'debugger', 'legacy', 'x [promoted to: CLAUDE.md]',
		    1.0, 5, '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z',
		    'promoted', NULL)`,
	)
	if err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}

	// Re-run migrate — backfill should split the suffix into promoted_to.
	if err := db.migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var value string
	var promotedTo sql.NullString
	err = db.sql.QueryRow(
		`SELECT value, promoted_to FROM memories WHERE key = 'legacy'`,
	).Scan(&value, &promotedTo)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if value != "x" {
		t.Errorf("backfill left value = %q, want %q", value, "x")
	}
	if !promotedTo.Valid || promotedTo.String != "CLAUDE.md" {
		t.Errorf("backfill set promoted_to = %v, want \"CLAUDE.md\"", promotedTo)
	}
}

func TestPromoteBackfillIdempotent(t *testing.T) {
	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory: %v", err)
	}
	defer db.Close()

	// Seed a properly promoted row (not the old shape).
	_, err = db.sql.Exec(
		`INSERT INTO memories
		   (namespace, agent, key, value, confidence, hit_count,
		    created_at, updated_at, lifecycle, promoted_to)
		 VALUES
		   ('bug_pattern', 'debugger', 'ok', '{"a":"b"}',
		    1.0, 0, '2024-01-01T00:00:00Z', '2024-01-01T00:00:00Z',
		    'promoted', 'CLAUDE.md')`,
	)
	if err != nil {
		t.Fatalf("seed clean row: %v", err)
	}

	// Re-run migrate multiple times — must not touch clean rows.
	for range 3 {
		if err := db.migrate(); err != nil {
			t.Fatalf("migrate: %v", err)
		}
	}

	var value, promotedTo string
	db.sql.QueryRow(
		`SELECT value, promoted_to FROM memories WHERE key = 'ok'`,
	).Scan(&value, &promotedTo)
	if value != `{"a":"b"}` {
		t.Errorf("clean row value mutated: %q", value)
	}
	if promotedTo != "CLAUDE.md" {
		t.Errorf("clean row promoted_to changed: %q", promotedTo)
	}
}

func TestOpenFilePath(t *testing.T) {
	dir := t.TempDir()
	dbPath := dir + "/test.db"

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open(%q) failed: %v", dbPath, err)
	}
	defer db.Close()

	// Should be able to store
	_, err = db.Store("ns", "agent", "key", "value")
	if err != nil {
		t.Fatalf("store after file-based open failed: %v", err)
	}

	// Close and reopen — data should persist
	db.Close()

	db2, err := Open(dbPath)
	if err != nil {
		t.Fatalf("reopen failed: %v", err)
	}
	defer db2.Close()

	entry, err := db2.Get("ns", "key")
	if err != nil {
		t.Fatalf("get after reopen failed: %v", err)
	}
	if entry.Value != "value" {
		t.Errorf("got value %q, want %q", entry.Value, "value")
	}
}
