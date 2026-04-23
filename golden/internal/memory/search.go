package memory

import (
	"fmt"
	"strings"
	"time"
)

// Search performs a BM25-ranked full-text search over stored memories.
// Scoped to the given namespace; only returns entries with lifecycle
// 'active' or 'validated'. Results are ordered by hit_count DESC,
// then updated_at DESC, with BM25 relevance as tiebreaker.
//
// limit <= 0 defaults to 10; limit > 100 is clamped to 100.
// Returns an empty non-nil slice when no matches are found.
// Increments hit_count for all returned entries.
func (d *DB) Search(namespace, query string, limit int) ([]MemoryEntry, error) {
	if query == "" {
		return []MemoryEntry{}, nil
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	rows, err := d.sql.Query(
		`SELECT m.id, m.namespace, m.agent, m.key, m.value,
		        m.confidence, m.hit_count,
		        m.created_at, m.updated_at, m.expires_at, m.lifecycle
		 FROM memories_fts
		 JOIN memories m ON m.id = memories_fts.rowid
		 WHERE memories_fts MATCH ?
		   AND m.namespace = ?
		   AND m.lifecycle IN ('active', 'validated')
		 ORDER BY m.hit_count DESC, m.updated_at DESC, bm25(memories_fts)
		 LIMIT ?`,
		query, namespace, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	entries, err := scanEntries(rows)
	if err != nil {
		return nil, err
	}

	// Increment hit_count for returned entries.
	if len(entries) > 0 {
		ids := make([]int64, len(entries))
		for i, e := range entries {
			ids[i] = e.ID
		}
		inClause, args := buildInt64InClause(ids)
		now := time.Now().UTC().Format(time.RFC3339)
		args = append([]interface{}{now}, args...)
		_, _ = d.sql.Exec(
			fmt.Sprintf(
				`UPDATE memories SET hit_count = hit_count + 1, updated_at = ? WHERE id IN (%s)`,
				inClause,
			),
			args...,
		)
	}

	return entries, nil
}

// buildInt64InClause builds a SQL IN clause placeholder string and args for the given IDs.
func buildInt64InClause(ids []int64) (string, []interface{}) {
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}
	return strings.Join(placeholders, ", "), args
}
