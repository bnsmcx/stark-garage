package memory

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound is returned when Get or Delete finds no matching entry.
var ErrNotFound = errors.New("memory entry not found")

// Store inserts or updates a memory entry for the given namespace and key.
// If an entry with the same (namespace, key) already exists, its value and
// updated_at are updated; hit_count and confidence are preserved.
// Returns the ID of the stored entry and any error.
func (d *DB) Store(namespace, agent, key, value string) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := d.sql.Exec(
		`INSERT INTO memories (namespace, agent, key, value, confidence, hit_count, created_at, updated_at, lifecycle)
		 VALUES (?, ?, ?, ?, 0.5, 0, ?, ?, 'active')
		 ON CONFLICT(namespace, key) DO UPDATE SET
		   value = excluded.value,
		   agent = excluded.agent,
		   updated_at = excluded.updated_at`,
		namespace, agent, key, value, now, now,
	)
	if err != nil {
		return 0, fmt.Errorf("store failed: %w", err)
	}

	// Retrieve the actual ID (may be the existing one on upsert).
	var id int64
	err = d.sql.QueryRow(
		`SELECT id FROM memories WHERE namespace = ? AND key = ?`,
		namespace, key,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("store: failed to retrieve id: %w", err)
	}

	return id, nil
}

// Get retrieves a single memory entry by namespace and key.
// Increments hit_count and updates updated_at on access.
// Returns nil, ErrNotFound if no entry exists.
func (d *DB) Get(namespace, key string) (*MemoryEntry, error) {
	row := d.sql.QueryRow(
		`SELECT id, namespace, agent, key, value, confidence, hit_count,
		        created_at, updated_at, expires_at, lifecycle
		 FROM memories
		 WHERE namespace = ? AND key = ?`,
		namespace, key,
	)

	entry, err := scanEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get failed: %w", err)
	}

	// Increment hit_count and updated_at. Confidence is intentionally not
	// mutated on read — it's set at write (0.5) and promote (1.0) only.
	now := time.Now().UTC().Format(time.RFC3339)
	_, _ = d.sql.Exec(
		`UPDATE memories
		 SET hit_count = hit_count + 1,
		     updated_at = ?
		 WHERE id = ?`,
		now, entry.ID,
	)

	return entry, nil
}

// Peek retrieves a single memory entry by namespace and key WITHOUT any side
// effects: hit_count, updated_at, and confidence are untouched. Use this for
// browsing or auditing; use Get for "I am consuming this entry" semantics.
// Returns nil, ErrNotFound if no entry exists.
func (d *DB) Peek(namespace, key string) (*MemoryEntry, error) {
	row := d.sql.QueryRow(
		`SELECT id, namespace, agent, key, value, confidence, hit_count,
		        created_at, updated_at, expires_at, lifecycle
		 FROM memories
		 WHERE namespace = ? AND key = ?`,
		namespace, key,
	)

	entry, err := scanEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("peek failed: %w", err)
	}
	return entry, nil
}

// List returns all memory entries for the given namespace where lifecycle
// is 'active' or 'validated', ordered by hit_count DESC, updated_at DESC.
// Returns an empty non-nil slice if none exist.
func (d *DB) List(namespace string) ([]MemoryEntry, error) {
	rows, err := d.sql.Query(
		`SELECT id, namespace, agent, key, value, confidence, hit_count,
		        created_at, updated_at, expires_at, lifecycle
		 FROM memories
		 WHERE namespace = ?
		   AND lifecycle IN ('active', 'validated')
		 ORDER BY hit_count DESC, updated_at DESC`,
		namespace,
	)
	if err != nil {
		return nil, fmt.Errorf("list failed: %w", err)
	}
	defer rows.Close()

	return scanEntries(rows)
}

// Delete removes a single memory entry by namespace and key.
// Returns ErrNotFound if no entry exists.
func (d *DB) Delete(namespace, key string) error {
	result, err := d.sql.Exec(
		`DELETE FROM memories WHERE namespace = ? AND key = ?`,
		namespace, key,
	)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete: rows affected check failed: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// scanEntry scans a single row into a MemoryEntry.
func scanEntry(row *sql.Row) (*MemoryEntry, error) {
	var e MemoryEntry
	var createdAt, updatedAt string
	var expiresAt sql.NullString

	err := row.Scan(
		&e.ID, &e.Namespace, &e.Agent, &e.Key, &e.Value,
		&e.Confidence, &e.HitCount,
		&createdAt, &updatedAt, &expiresAt, &e.Lifecycle,
	)
	if err != nil {
		return nil, err
	}

	if err := parseEntryTimes(&e, createdAt, updatedAt, expiresAt); err != nil {
		return nil, err
	}
	return &e, nil
}

// scanEntryFromRows scans a single row from sql.Rows into a MemoryEntry.
func scanEntryFromRows(rows *sql.Rows) (*MemoryEntry, error) {
	var e MemoryEntry
	var createdAt, updatedAt string
	var expiresAt sql.NullString

	err := rows.Scan(
		&e.ID, &e.Namespace, &e.Agent, &e.Key, &e.Value,
		&e.Confidence, &e.HitCount,
		&createdAt, &updatedAt, &expiresAt, &e.Lifecycle,
	)
	if err != nil {
		return nil, err
	}

	if err := parseEntryTimes(&e, createdAt, updatedAt, expiresAt); err != nil {
		return nil, err
	}
	return &e, nil
}

// scanEntries collects all rows into a non-nil slice of MemoryEntry.
func scanEntries(rows *sql.Rows) ([]MemoryEntry, error) {
	entries := make([]MemoryEntry, 0)
	for rows.Next() {
		e, err := scanEntryFromRows(rows)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		entries = append(entries, *e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return entries, nil
}

// parseEntryTimes parses time fields from their string representations.
func parseEntryTimes(e *MemoryEntry, createdAt, updatedAt string, expiresAt sql.NullString) error {
	t, err := parseTime(createdAt)
	if err != nil {
		return fmt.Errorf("parse created_at %q: %w", createdAt, err)
	}
	e.CreatedAt = t

	t, err = parseTime(updatedAt)
	if err != nil {
		return fmt.Errorf("parse updated_at %q: %w", updatedAt, err)
	}
	e.UpdatedAt = t

	if expiresAt.Valid && expiresAt.String != "" {
		t, err := parseTime(expiresAt.String)
		if err != nil {
			return fmt.Errorf("parse expires_at %q: %w", expiresAt.String, err)
		}
		e.ExpiresAt = &t
	}

	return nil
}

// parseTime tries RFC3339, then RFC3339Nano as a fallback.
func parseTime(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t, err = time.Parse(time.RFC3339Nano, s)
		if err != nil {
			return time.Time{}, err
		}
	}
	return t, nil
}
