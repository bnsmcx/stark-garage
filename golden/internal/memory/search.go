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
//
// When raw is false (default for CLI callers), unquoted tokens containing
// FTS5-special characters (- : ( ) * ") are auto-quoted so identifiers
// like "alt-screen" are treated as phrase literals instead of parsing as
// operator expressions. Pass raw=true to supply FTS5 operator syntax.
func (d *DB) Search(namespace, query string, limit int, raw bool) ([]MemoryEntry, error) {
	if query == "" {
		return []MemoryEntry{}, nil
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	ftsQuery := query
	if !raw {
		ftsQuery = sanitizeFTSQuery(query)
	}

	rows, err := d.sql.Query(
		`SELECT m.id, m.namespace, m.agent, m.key, m.value,
		        m.confidence, m.hit_count,
		        m.created_at, m.updated_at, m.expires_at, m.lifecycle, m.promoted_to
		 FROM memories_fts
		 JOIN memories m ON m.id = memories_fts.rowid
		 WHERE memories_fts MATCH ?
		   AND m.namespace = ?
		   AND m.lifecycle IN ('active', 'validated')
		 ORDER BY m.hit_count DESC, m.updated_at DESC, bm25(memories_fts)
		 LIMIT ?`,
		ftsQuery, namespace, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("fts query rejected — try quoting tokens with hyphens, or use --raw for FTS operator syntax: %w", err)
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

// sanitizeFTSQuery quotes each whitespace-separated token containing one of the
// FTS5 special characters (- : ( ) * ") so that e.g. "alt-screen" is treated as
// a phrase literal instead of "alt NOT screen" (which detonates as an unknown
// column reference). Tokens already enclosed in "..." pass through unchanged.
// Embedded double quotes inside a raw token are escaped to "" per FTS5 grammar.
func sanitizeFTSQuery(q string) string {
	var out strings.Builder
	i := 0
	first := true
	for i < len(q) {
		c := q[i]
		// Collapse runs of whitespace.
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}

		if !first {
			out.WriteByte(' ')
		}
		first = false

		// Already-quoted span: consume through the closing quote.
		if c == '"' {
			out.WriteByte('"')
			i++
			for i < len(q) {
				if q[i] == '"' {
					out.WriteByte('"')
					i++
					break
				}
				out.WriteByte(q[i])
				i++
			}
			continue
		}

		// Unquoted token: grab up to next whitespace.
		start := i
		for i < len(q) {
			b := q[i]
			if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
				break
			}
			i++
		}
		tok := q[start:i]
		if ftsTokenNeedsQuoting(tok) {
			out.WriteByte('"')
			for _, ch := range tok {
				if ch == '"' {
					out.WriteString(`""`)
				} else {
					out.WriteRune(ch)
				}
			}
			out.WriteByte('"')
		} else {
			out.WriteString(tok)
		}
	}
	return out.String()
}

func ftsTokenNeedsQuoting(tok string) bool {
	for _, c := range tok {
		switch c {
		case '-', ':', '(', ')', '*', '"':
			return true
		}
	}
	return false
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
