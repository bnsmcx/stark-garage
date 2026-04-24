package memory

import (
	"database/sql"
	"fmt"
	"testing"
	"time"
)

func TestPruneActiveToStale(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "old-no-hits", "old content")

	// Backdate updated_at to 90 days ago (past the 60-day threshold)
	old := time.Now().UTC().Add(-90 * 24 * time.Hour).Format(time.RFC3339)
	db.sql.Exec(`UPDATE memories SET updated_at = ?, hit_count = 0 WHERE key = 'old-no-hits'`, old)

	n, err := db.Prune(200)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if n == 0 {
		t.Fatal("Prune returned 0 transitions, want >= 1")
	}

	// Verify it's now stale
	var lifecycle string
	db.sql.QueryRow(`SELECT lifecycle FROM memories WHERE key = 'old-no-hits'`).Scan(&lifecycle)
	if lifecycle != LifecycleStale {
		t.Errorf("lifecycle = %q, want %q", lifecycle, LifecycleStale)
	}
}

func TestPruneActiveToValidated(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("bug_pattern", "debugger", "recurring-bug", "content")
	// Two hits is enough to validate.
	db.sql.Exec(`UPDATE memories SET hit_count = 2 WHERE key = 'recurring-bug'`)

	n, err := db.Prune(200)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if n < 1 {
		t.Errorf("Prune returned %d transitions, want >= 1", n)
	}

	var lifecycle string
	db.sql.QueryRow(`SELECT lifecycle FROM memories WHERE key = 'recurring-bug'`).Scan(&lifecycle)
	if lifecycle != LifecycleValidated {
		t.Errorf("lifecycle = %q, want %q", lifecycle, LifecycleValidated)
	}
}

func TestPruneValidatedStaysValidated(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("bug_pattern", "debugger", "already-validated", "content")
	db.sql.Exec(`UPDATE memories SET lifecycle = 'validated', hit_count = 5 WHERE key = 'already-validated'`)

	_, err := db.Prune(200)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	var lifecycle string
	db.sql.QueryRow(`SELECT lifecycle FROM memories WHERE key = 'already-validated'`).Scan(&lifecycle)
	if lifecycle != LifecycleValidated {
		t.Errorf("lifecycle = %q, want %q (validated should stay validated)", lifecycle, LifecycleValidated)
	}
}

func TestPruneActiveHitCountZeroNotValidated(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "fresh-entry", "content")
	// hit_count stays 0 — entry is fresh, should remain active, not validated

	_, err := db.Prune(200)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	var lifecycle string
	db.sql.QueryRow(`SELECT lifecycle FROM memories WHERE key = 'fresh-entry'`).Scan(&lifecycle)
	if lifecycle != LifecycleActive {
		t.Errorf("lifecycle = %q, want %q (hit_count=0 should stay active)", lifecycle, LifecycleActive)
	}
}

func TestPruneStaleToArchived(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "stale-entry", "stale content")

	// Set to stale and backdate past the 30-day threshold
	old := time.Now().UTC().Add(-45 * 24 * time.Hour).Format(time.RFC3339)
	db.sql.Exec(`UPDATE memories SET lifecycle = 'stale', updated_at = ? WHERE key = 'stale-entry'`, old)

	n, err := db.Prune(200)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if n == 0 {
		t.Fatal("Prune returned 0 transitions, want >= 1")
	}

	var lifecycle string
	db.sql.QueryRow(`SELECT lifecycle FROM memories WHERE key = 'stale-entry'`).Scan(&lifecycle)
	if lifecycle != LifecycleArchived {
		t.Errorf("lifecycle = %q, want %q", lifecycle, LifecycleArchived)
	}
}

func TestPruneActiveWithHitsGoesToValidatedNotStale(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "active-with-hits", "useful content")

	// Backdate past the stale cutoff, but give it hits so it qualifies for validated.
	old := time.Now().UTC().Add(-90 * 24 * time.Hour).Format(time.RFC3339)
	db.sql.Exec(`UPDATE memories SET updated_at = ?, hit_count = 5 WHERE key = 'active-with-hits'`, old)

	_, err := db.Prune(200)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}

	var lifecycle string
	db.sql.QueryRow(`SELECT lifecycle FROM memories WHERE key = 'active-with-hits'`).Scan(&lifecycle)
	if lifecycle != LifecycleValidated {
		t.Errorf("lifecycle = %q, want %q (hits should route to validated, not stale)", lifecycle, LifecycleValidated)
	}
}

func TestPruneOverflowArchival(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	// Create 5 active entries
	for i := 0; i < 5; i++ {
		db.Store("lesson", "pomo", fmt.Sprintf("entry-%d", i), fmt.Sprintf("content %d", i))
	}

	// Prune with maxActive=3 — should archive 2
	n, err := db.Prune(3)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if n != 2 {
		t.Errorf("Prune transitioned %d, want 2", n)
	}

	stats, _ := db.Stats()
	if stats.Active != 3 {
		t.Errorf("active count = %d, want 3", stats.Active)
	}
	if stats.Archived != 2 {
		t.Errorf("archived count = %d, want 2", stats.Archived)
	}
}

func TestPromote(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "promote-me", "original value")

	err := db.Promote("lesson", "promote-me", "CLAUDE.md Development Philosophy")
	if err != nil {
		t.Fatalf("Promote: %v", err)
	}

	var lifecycle, value string
	var confidence float64
	var promotedTo sql.NullString
	db.sql.QueryRow(
		`SELECT lifecycle, confidence, value, promoted_to FROM memories WHERE key = 'promote-me'`,
	).Scan(&lifecycle, &confidence, &value, &promotedTo)

	if lifecycle != LifecyclePromoted {
		t.Errorf("lifecycle = %q, want %q", lifecycle, LifecyclePromoted)
	}
	if confidence != 1.0 {
		t.Errorf("confidence = %f, want 1.0", confidence)
	}
	if value != "original value" {
		t.Errorf("value mutated: got %q, want %q", value, "original value")
	}
	if !promotedTo.Valid || promotedTo.String != "CLAUDE.md Development Philosophy" {
		t.Errorf("promoted_to = %v, want \"CLAUDE.md Development Philosophy\"", promotedTo)
	}
}

func TestPromotePreservesValue(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	original := `{"rule":"do X","why":"because Y"}`
	db.Store("bug_pattern", "debugger", "json-entry", original)

	if err := db.Promote("bug_pattern", "json-entry", "CLAUDE.md"); err != nil {
		t.Fatalf("Promote: %v", err)
	}

	entry, err := db.Peek("bug_pattern", "json-entry")
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if entry.Value != original {
		t.Errorf("value mutated by Promote:\n  got:  %q\n  want: %q", entry.Value, original)
	}
	if entry.PromotedTo == nil || *entry.PromotedTo != "CLAUDE.md" {
		t.Errorf("PromotedTo = %v, want &\"CLAUDE.md\"", entry.PromotedTo)
	}
	if entry.Lifecycle != LifecyclePromoted {
		t.Errorf("lifecycle = %q, want %q", entry.Lifecycle, LifecyclePromoted)
	}
}

func TestPromoteNotFound(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	err := db.Promote("lesson", "nonexistent", "CLAUDE.md")
	if err != ErrNotFound {
		t.Errorf("Promote nonexistent: got %v, want ErrNotFound", err)
	}
}

func TestStats(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "a", "v1")
	db.Store("lesson", "pomo", "b", "v2")
	db.Store("lesson", "pomo", "c", "v3")
	db.sql.Exec(`UPDATE memories SET lifecycle = 'validated' WHERE key = 'b'`)
	db.sql.Exec(`UPDATE memories SET lifecycle = 'archived' WHERE key = 'c'`)

	stats, err := db.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}

	if stats.Active != 1 {
		t.Errorf("active = %d, want 1", stats.Active)
	}
	if stats.Validated != 1 {
		t.Errorf("validated = %d, want 1", stats.Validated)
	}
	if stats.Archived != 1 {
		t.Errorf("archived = %d, want 1", stats.Archived)
	}
	if stats.Total != 3 {
		t.Errorf("total = %d, want 3", stats.Total)
	}
}

func TestNamespaces(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("bug_pattern", "debugger", "b1", "v")
	db.Store("bug_pattern", "debugger", "b2", "v")
	db.Store("lesson", "pomo", "l1", "v")
	db.Store("calibration", "planner", "c1", "v")

	db.sql.Exec(`UPDATE memories SET lifecycle = 'validated' WHERE key = 'b2'`)
	db.sql.Exec(`UPDATE memories SET lifecycle = 'archived' WHERE key = 'c1'`)

	out, err := db.Namespaces()
	if err != nil {
		t.Fatalf("Namespaces: %v", err)
	}

	if len(out) != 3 {
		t.Fatalf("expected 3 namespaces, got %d: %+v", len(out), out)
	}

	// Results are sorted alphabetically.
	if out[0].Namespace != "bug_pattern" {
		t.Errorf("out[0].Namespace = %q, want %q", out[0].Namespace, "bug_pattern")
	}
	if out[1].Namespace != "calibration" {
		t.Errorf("out[1].Namespace = %q, want %q", out[1].Namespace, "calibration")
	}
	if out[2].Namespace != "lesson" {
		t.Errorf("out[2].Namespace = %q, want %q", out[2].Namespace, "lesson")
	}

	// bug_pattern: 1 active + 1 validated = 2 total
	if out[0].Active != 1 || out[0].Validated != 1 || out[0].Total != 2 {
		t.Errorf("bug_pattern = %+v, want Active=1 Validated=1 Total=2", out[0])
	}
	// calibration: 1 archived = 1 total
	if out[1].Archived != 1 || out[1].Total != 1 {
		t.Errorf("calibration = %+v, want Archived=1 Total=1", out[1])
	}
	// lesson: 1 active = 1 total
	if out[2].Active != 1 || out[2].Total != 1 {
		t.Errorf("lesson = %+v, want Active=1 Total=1", out[2])
	}
}

func TestNamespacesEmpty(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	out, err := db.Namespaces()
	if err != nil {
		t.Fatalf("Namespaces: %v", err)
	}
	if out == nil {
		t.Error("Namespaces returned nil slice, want empty non-nil slice")
	}
	if len(out) != 0 {
		t.Errorf("expected 0 namespaces, got %d", len(out))
	}
}

func TestStatsEmpty(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	stats, err := db.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.Total != 0 {
		t.Errorf("total = %d, want 0", stats.Total)
	}
}
