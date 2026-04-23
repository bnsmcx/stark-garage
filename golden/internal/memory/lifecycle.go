package memory

import (
	"fmt"
	"time"
)

// Prune performs lifecycle transitions on memory entries:
//
//  1. Active entries with hit_count >= 2 are transitioned to 'validated'.
//  2. Active entries with no hits (hit_count = 0) and updated_at older than 60 days
//     are transitioned to 'stale'.
//  3. Stale entries with updated_at older than 30 days are transitioned to 'archived'.
//  4. If the count of active entries exceeds maxActive, the lowest-hit active entries
//     are archived until the count is at or below maxActive.
//
// Returns the total number of transitions performed.
func (d *DB) Prune(maxActive int) (int, error) {
	if maxActive <= 0 {
		maxActive = 200
	}

	totalTransitions := 0
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// Step 1: active with hit_count >= 2 -> validated
	result, err := d.sql.Exec(
		`UPDATE memories
		 SET lifecycle = 'validated', updated_at = ?
		 WHERE lifecycle = 'active'
		   AND hit_count >= 2`,
		nowStr,
	)
	if err != nil {
		return 0, fmt.Errorf("prune: active->validated transition failed: %w", err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune: rows affected check failed: %w", err)
	}
	totalTransitions += int(n)

	// Step 2: active with no hits and stale for 60+ days -> stale
	staleCutoff := now.Add(-60 * 24 * time.Hour).Format(time.RFC3339)

	result, err = d.sql.Exec(
		`UPDATE memories
		 SET lifecycle = 'stale', updated_at = ?
		 WHERE lifecycle = 'active'
		   AND hit_count = 0
		   AND updated_at < ?`,
		nowStr, staleCutoff,
	)
	if err != nil {
		return totalTransitions, fmt.Errorf("prune: active->stale transition failed: %w", err)
	}
	n, err = result.RowsAffected()
	if err != nil {
		return totalTransitions, fmt.Errorf("prune: rows affected check failed: %w", err)
	}
	totalTransitions += int(n)

	// Step 3: stale for 30+ more days -> archived
	archiveCutoff := now.Add(-30 * 24 * time.Hour).Format(time.RFC3339)

	result, err = d.sql.Exec(
		`UPDATE memories
		 SET lifecycle = 'archived', updated_at = ?
		 WHERE lifecycle = 'stale'
		   AND updated_at < ?`,
		nowStr, archiveCutoff,
	)
	if err != nil {
		return totalTransitions, fmt.Errorf("prune: stale->archived transition failed: %w", err)
	}
	n, err = result.RowsAffected()
	if err != nil {
		return totalTransitions, fmt.Errorf("prune: rows affected check failed: %w", err)
	}
	totalTransitions += int(n)

	// Step 4: if active count > maxActive, archive lowest-hit active entries
	var activeCount int
	err = d.sql.QueryRow(
		`SELECT COUNT(*) FROM memories WHERE lifecycle = 'active'`,
	).Scan(&activeCount)
	if err != nil {
		return totalTransitions, fmt.Errorf("prune: count active failed: %w", err)
	}

	if activeCount > maxActive {
		excess := activeCount - maxActive
		result, err = d.sql.Exec(
			`UPDATE memories
			 SET lifecycle = 'archived', updated_at = ?
			 WHERE id IN (
			   SELECT id FROM memories
			   WHERE lifecycle = 'active'
			   ORDER BY hit_count ASC, confidence ASC, updated_at ASC
			   LIMIT ?
			 )`,
			nowStr, excess,
		)
		if err != nil {
			return totalTransitions, fmt.Errorf("prune: overflow archive failed: %w", err)
		}
		n, err = result.RowsAffected()
		if err != nil {
			return totalTransitions, fmt.Errorf("prune: rows affected check failed: %w", err)
		}
		totalTransitions += int(n)
	}

	return totalTransitions, nil
}

// Promote marks a memory entry as promoted and records where it was promoted to.
// The value is updated to include the promotion target and the lifecycle is set
// to 'promoted'. The entry's confidence is set to 1.0.
func (d *DB) Promote(namespace, key, promotedTo string) error {
	now := time.Now().UTC().Format(time.RFC3339)

	result, err := d.sql.Exec(
		`UPDATE memories
		 SET lifecycle = 'promoted',
		     confidence = 1.0,
		     updated_at = ?,
		     value = value || ' [promoted to: ' || ? || ']'
		 WHERE namespace = ? AND key = ?`,
		now, promotedTo, namespace, key,
	)
	if err != nil {
		return fmt.Errorf("promote failed: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("promote: rows affected check failed: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Stats returns counts per lifecycle state.
func (d *DB) Stats() (*LifecycleStats, error) {
	rows, err := d.sql.Query(
		`SELECT lifecycle, COUNT(*) FROM memories GROUP BY lifecycle`,
	)
	if err != nil {
		return nil, fmt.Errorf("stats failed: %w", err)
	}
	defer rows.Close()

	stats := &LifecycleStats{}
	for rows.Next() {
		var state string
		var count int
		if err := rows.Scan(&state, &count); err != nil {
			return nil, fmt.Errorf("stats scan failed: %w", err)
		}
		switch state {
		case LifecycleActive:
			stats.Active = count
		case LifecycleValidated:
			stats.Validated = count
		case LifecyclePromoted:
			stats.Promoted = count
		case LifecycleStale:
			stats.Stale = count
		case LifecycleArchived:
			stats.Archived = count
		}
		stats.Total += count
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("stats rows error: %w", err)
	}

	return stats, nil
}
