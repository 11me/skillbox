// Package models provides common types for filtering.
// Place in: internal/models/common.go
package models

import "time"

// =============================================================================
// Date Filtering
// =============================================================================

// DateFilter represents a date range filter.
// Both From and To are optional (nil = unbounded).
type DateFilter struct {
	From *time.Time
	To   *time.Time
}

// DateInterval represents a concrete date interval (both required).
type DateInterval struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

// =============================================================================
// Numeric Range Filtering
// =============================================================================

// OptionalIntRange represents an optional integer range.
type OptionalIntRange struct {
	Min *int
	Max *int
}

// OptionalInt64Range represents an optional int64 range.
type OptionalInt64Range struct {
	Min *int64
	Max *int64
}

// OptionalFloatRange represents an optional float64 range.
type OptionalFloatRange struct {
	Min *float64
	Max *float64
}

// IntRange represents a concrete integer range (both required).
type IntRange struct {
	Min int
	Max int
}

// =============================================================================
// Money Range Filtering
// =============================================================================

// OptionalMoneyRange represents an optional money range.
// Used for price filtering.
type OptionalMoneyRange struct {
	Min *Money `json:"min"`
	Max *Money `json:"max"`
}

// =============================================================================
// Helper Functions
// =============================================================================

// ptr returns a pointer to the value.
// Useful for constructing filters inline.
func ptr[T any](v T) *T {
	return &v
}

// =============================================================================
// Usage Examples
// =============================================================================

// Example filter using common types:
//
//	type OrderFilter struct {
//	    ID        []string
//	    Status    []string
//	    CreatedAt *DateFilter        // Date range
//	    Total     *OptionalMoneyRange // Price range
//	    Quantity  *OptionalIntRange   // Quantity range
//	}
//
// Usage:
//
//	filter := &OrderFilter{
//	    Status:    []string{"pending", "processing"},
//	    CreatedAt: &DateFilter{From: ptr(time.Now().AddDate(0, -1, 0))},
//	    Total:     &OptionalMoneyRange{Min: ptr(Money{Amount: 1000})},
//	}
