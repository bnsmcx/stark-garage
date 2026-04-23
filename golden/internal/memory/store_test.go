package memory

import (
	"testing"
)

func mustOpenInMemory(t *testing.T) *DB {
	t.Helper()
	db, err := openInMemory()
	if err != nil {
		t.Fatalf("openInMemory: %v", err)
	}
	return db
}

func TestStoreAndGet(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	id, err := db.Store("lesson", "pomo", "test-pattern", `{"wrong":"x","right":"y"}`)
	if err != nil {
		t.Fatalf("Store: %v", err)
	}
	if id == 0 {
		t.Fatal("Store returned id 0")
	}

	entry, err := db.Get("lesson", "test-pattern")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if entry.Namespace != "lesson" {
		t.Errorf("namespace = %q, want %q", entry.Namespace, "lesson")
	}
	if entry.Agent != "pomo" {
		t.Errorf("agent = %q, want %q", entry.Agent, "pomo")
	}
	if entry.Key != "test-pattern" {
		t.Errorf("key = %q, want %q", entry.Key, "test-pattern")
	}
	if entry.Value != `{"wrong":"x","right":"y"}` {
		t.Errorf("value = %q, want JSON", entry.Value)
	}
	if entry.Lifecycle != LifecycleActive {
		t.Errorf("lifecycle = %q, want %q", entry.Lifecycle, LifecycleActive)
	}
	if entry.Confidence != 0.5 {
		t.Errorf("initial confidence = %f, want 0.5", entry.Confidence)
	}
}

func TestStoreUpsert(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	id1, _ := db.Store("lesson", "pomo", "key1", "value1")
	id2, _ := db.Store("lesson", "pomo", "key1", "value2")

	if id1 != id2 {
		t.Errorf("upsert returned different IDs: %d vs %d", id1, id2)
	}

	entry, _ := db.Get("lesson", "key1")
	if entry.Value != "value2" {
		t.Errorf("value after upsert = %q, want %q", entry.Value, "value2")
	}
}

func TestGetIncrementsHitCount(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "key1", "value1")

	// First Get — hit_count goes from 0 to 1
	db.Get("lesson", "key1")
	// Second Get — hit_count goes from 1 to 2
	db.Get("lesson", "key1")

	// Read raw to check hit_count (Get increments, so we check via list)
	entries, _ := db.List("lesson")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	// After 2 Gets, hit_count should be 2
	if entries[0].HitCount < 2 {
		t.Errorf("hit_count = %d after 2 Gets, want >= 2", entries[0].HitCount)
	}
}

func TestGetDoesNotBumpConfidence(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "key1", "value1")
	// Initial confidence is 0.5 and must stay that way across repeated Gets.
	db.Get("lesson", "key1")
	db.Get("lesson", "key1")
	db.Get("lesson", "key1")

	entries, _ := db.List("lesson")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Confidence != 0.5 {
		t.Errorf("confidence = %f after 3 Gets, want 0.5 (no auto-bump)", entries[0].Confidence)
	}
}

func TestPeekNoSideEffects(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	_, err := db.Store("bug_pattern", "debugger", "peek-target", "v")
	if err != nil {
		t.Fatalf("Store: %v", err)
	}

	before, err := db.Peek("bug_pattern", "peek-target")
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}

	// Multiple peeks must not touch hit_count, updated_at, or confidence.
	db.Peek("bug_pattern", "peek-target")
	db.Peek("bug_pattern", "peek-target")
	after, _ := db.Peek("bug_pattern", "peek-target")

	if after.HitCount != before.HitCount {
		t.Errorf("HitCount changed: before=%d after=%d", before.HitCount, after.HitCount)
	}
	if after.Confidence != before.Confidence {
		t.Errorf("Confidence changed: before=%f after=%f", before.Confidence, after.Confidence)
	}
	if !after.UpdatedAt.Equal(before.UpdatedAt) {
		t.Errorf("UpdatedAt changed: before=%v after=%v", before.UpdatedAt, after.UpdatedAt)
	}
}

func TestPeekNotFound(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	_, err := db.Peek("lesson", "nonexistent")
	if err != ErrNotFound {
		t.Errorf("Peek nonexistent: got %v, want ErrNotFound", err)
	}
}

func TestGetNotFound(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	_, err := db.Get("lesson", "nonexistent")
	if err != ErrNotFound {
		t.Errorf("Get nonexistent: got %v, want ErrNotFound", err)
	}
}

func TestList(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "a", "va")
	db.Store("lesson", "pomo", "b", "vb")
	db.Store("bug_pattern", "debugger", "c", "vc")

	entries, err := db.List("lesson")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("List(lesson) returned %d entries, want 2", len(entries))
	}

	entries, err = db.List("bug_pattern")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("List(bug_pattern) returned %d entries, want 1", len(entries))
	}
}

func TestListExcludesArchivedAndPromoted(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "active-one", "v1")
	db.Store("lesson", "pomo", "will-archive", "v2")
	db.Store("lesson", "pomo", "will-promote", "v3")

	// Manually set lifecycles
	db.sql.Exec(`UPDATE memories SET lifecycle = 'archived' WHERE key = 'will-archive'`)
	db.sql.Exec(`UPDATE memories SET lifecycle = 'promoted' WHERE key = 'will-promote'`)

	entries, _ := db.List("lesson")
	if len(entries) != 1 {
		t.Errorf("List returned %d entries, want 1 (only active)", len(entries))
	}
	if len(entries) > 0 && entries[0].Key != "active-one" {
		t.Errorf("expected 'active-one', got %q", entries[0].Key)
	}
}

func TestDelete(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "key1", "value1")

	err := db.Delete("lesson", "key1")
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = db.Get("lesson", "key1")
	if err != ErrNotFound {
		t.Errorf("after Delete, Get returned %v, want ErrNotFound", err)
	}
}

func TestDeleteNotFound(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	err := db.Delete("lesson", "nonexistent")
	if err != ErrNotFound {
		t.Errorf("Delete nonexistent: got %v, want ErrNotFound", err)
	}
}
