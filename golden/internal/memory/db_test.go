package memory

import "testing"

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
