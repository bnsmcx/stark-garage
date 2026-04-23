package memory

import (
	"fmt"
	"testing"
)

func TestSanitizeFTSQuery(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"plain", "nil pointer", "nil pointer"},
		{"hyphen", "alt-screen", `"alt-screen"`},
		{"mixed", "alt-screen logger", `"alt-screen" logger`},
		{"already quoted", `"alt-screen"`, `"alt-screen"`},
		{"quoted then plain", `"alt-screen" logger`, `"alt-screen" logger`},
		{"embedded quote", `foo"bar-baz`, `"foo""bar-baz"`},
		{"empty", "", ""},
		{"whitespace only", "   ", ""},
		{"colon", "ns:key", `"ns:key"`},
		{"parens", "foo(bar)", `"foo(bar)"`},
		{"star", "wild*card", `"wild*card"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := sanitizeFTSQuery(tc.in); got != tc.want {
				t.Errorf("sanitizeFTSQuery(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestSearchHyphenatedToken(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("bug_pattern", "debugger", "alt-screen-bug", "this entry mentions alt-screen in the logger output")
	db.Store("bug_pattern", "debugger", "unrelated", "nothing to see here")

	results, err := db.Search("bug_pattern", "alt-screen logger", 10, false)
	if err != nil {
		t.Fatalf("Search hyphenated: %v", err)
	}
	if len(results) < 1 {
		t.Fatalf("expected >= 1 result for 'alt-screen logger', got %d", len(results))
	}
	found := false
	for _, r := range results {
		if r.Key == "alt-screen-bug" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'alt-screen-bug' in results, got %+v", results)
	}
}

func TestSearchRawPassesThrough(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("bug_pattern", "debugger", "a-key", "alpha value")
	db.Store("bug_pattern", "debugger", "b-key", "beta value")

	// raw=true passes FTS operator syntax unchanged; "alpha NOT beta" via "alpha -beta".
	results, err := db.Search("bug_pattern", "alpha", 10, true)
	if err != nil {
		t.Fatalf("Search raw: %v", err)
	}
	if len(results) < 1 {
		t.Errorf("expected >= 1 result for raw 'alpha', got %d", len(results))
	}
}

func TestSearchBasic(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("bug_pattern", "debugger", "nil-pointer-reset", "state not reconstructed after destroy")
	db.Store("bug_pattern", "debugger", "race-condition-map", "concurrent map access without mutex")
	db.Store("lesson", "pomo", "unrelated-lesson", "something else entirely")

	results, err := db.Search("bug_pattern", "nil pointer", 10, false)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search returned 0 results, want >= 1")
	}
	if results[0].Key != "nil-pointer-reset" {
		t.Errorf("first result key = %q, want %q", results[0].Key, "nil-pointer-reset")
	}
}

func TestSearchNamespaceScoping(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("bug_pattern", "debugger", "nil-bug", "nil pointer in handler")
	db.Store("lesson", "pomo", "nil-lesson", "nil pointer lesson")

	// Search in bug_pattern namespace should not return lesson entries
	results, _ := db.Search("bug_pattern", "nil pointer", 10, false)
	for _, r := range results {
		if r.Namespace != "bug_pattern" {
			t.Errorf("search in bug_pattern returned entry from namespace %q", r.Namespace)
		}
	}
}

func TestSearchExcludesArchived(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "active-lesson", "active content about nil pointers")
	db.Store("lesson", "pomo", "archived-lesson", "archived content about nil pointers")
	db.sql.Exec(`UPDATE memories SET lifecycle = 'archived' WHERE key = 'archived-lesson'`)

	results, _ := db.Search("lesson", "nil pointers", 10, false)
	for _, r := range results {
		if r.Key == "archived-lesson" {
			t.Error("Search returned archived entry")
		}
	}
}

func TestSearchIncrementsHitCount(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("bug_pattern", "debugger", "nil-bug", "nil pointer in reset handler")

	db.Search("bug_pattern", "nil pointer", 10, false)

	// Check hit_count was incremented
	entries, _ := db.List("bug_pattern")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].HitCount < 1 {
		t.Errorf("hit_count = %d after Search, want >= 1", entries[0].HitCount)
	}
}

func TestSearchDoesNotBumpConfidence(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("bug_pattern", "debugger", "nil-bug", "nil pointer in reset handler")

	// Initial confidence is 0.5. Multiple searches must not inflate it.
	db.Search("bug_pattern", "nil pointer", 10, false)
	db.Search("bug_pattern", "nil pointer", 10, false)
	db.Search("bug_pattern", "nil pointer", 10, false)

	entries, _ := db.List("bug_pattern")
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Confidence != 0.5 {
		t.Errorf("confidence = %f after 3 Searches, want 0.5 (no auto-bump)", entries[0].Confidence)
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	db.Store("lesson", "pomo", "some-key", "some value")

	results, err := db.Search("lesson", "", 10, false)
	if err != nil {
		t.Fatalf("Search empty query: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("empty query returned %d results, want 0", len(results))
	}
}

func TestSearchLimitClamping(t *testing.T) {
	db := mustOpenInMemory(t)
	defer db.Close()

	// Store enough entries to test limit
	for i := 0; i < 5; i++ {
		db.Store("lesson", "pomo", fmt.Sprintf("key-%d", i), fmt.Sprintf("content about testing %d", i))
	}

	// Default limit (0 → 10)
	results, _ := db.Search("lesson", "testing", 0, false)
	if len(results) > 10 {
		t.Errorf("default limit returned %d results, want <= 10", len(results))
	}
}
