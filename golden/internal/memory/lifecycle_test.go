package memory

import (
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

func TestPruneDoesNotTouchActiveWithHits(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "active-with-hits", "useful content")

	// Backdate but give it hits
	old := time.Now().UTC().Add(-90 * 24 * time.Hour).Format(time.RFC3339)
	db.sql.Exec(`UPDATE memories SET updated_at = ?, hit_count = 5 WHERE key = 'active-with-hits'`, old)

	n, _ := db.Prune(200)
	if n != 0 {
		t.Errorf("Prune transitioned %d entries, want 0 (entry has hits)", n)
	}

	var lifecycle string
	db.sql.QueryRow(`SELECT lifecycle FROM memories WHERE key = 'active-with-hits'`).Scan(&lifecycle)
	if lifecycle != LifecycleActive {
		t.Errorf("lifecycle = %q, want %q (should not transition)", lifecycle, LifecycleActive)
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

	var lifecycle string
	var confidence float64
	var value string
	db.sql.QueryRow(
		`SELECT lifecycle, confidence, value FROM memories WHERE key = 'promote-me'`,
	).Scan(&lifecycle, &confidence, &value)

	if lifecycle != LifecyclePromoted {
		t.Errorf("lifecycle = %q, want %q", lifecycle, LifecyclePromoted)
	}
	if confidence != 1.0 {
		t.Errorf("confidence = %f, want 1.0", confidence)
	}
	if value == "original value" {
		t.Error("value not updated with promotion target")
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
