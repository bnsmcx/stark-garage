package memory

import "time"

// MemoryEntry is a single stored memory item with confidence scoring
// and lifecycle management.
type MemoryEntry struct {
	ID         int64      `json:"id"`
	Namespace  string     `json:"namespace"`
	Agent      string     `json:"agent"`
	Key        string     `json:"key"`
	Value      string     `json:"value"`
	Confidence float64    `json:"confidence"`
	HitCount   int        `json:"hitCount"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
	Lifecycle  string     `json:"lifecycle"`
	PromotedTo *string    `json:"promotedTo,omitempty"`
}

// LifecycleStats holds counts per lifecycle state.
type LifecycleStats struct {
	Active    int `json:"active"`
	Validated int `json:"validated"`
	Promoted  int `json:"promoted"`
	Stale     int `json:"stale"`
	Archived  int `json:"archived"`
	Total     int `json:"total"`
}

// NamespaceStats holds per-lifecycle counts scoped to a single namespace.
type NamespaceStats struct {
	Namespace string `json:"namespace"`
	Active    int    `json:"active"`
	Validated int    `json:"validated"`
	Promoted  int    `json:"promoted"`
	Stale     int    `json:"stale"`
	Archived  int    `json:"archived"`
	Total     int    `json:"total"`
}

// Valid lifecycle states.
const (
	LifecycleActive    = "active"
	LifecycleValidated = "validated"
	LifecyclePromoted  = "promoted"
	LifecycleStale     = "stale"
	LifecycleArchived  = "archived"
)

// Valid namespace values.
const (
	NamespaceBugPattern  = "bug_pattern"
	NamespaceSpecGap     = "spec_gap"
	NamespaceCalibration = "calibration"
	NamespaceRouting     = "routing"
	NamespaceLesson      = "lesson"
)
