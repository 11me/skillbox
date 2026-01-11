// Package storage provides SQL helper types for complex queries.
// Place in: internal/storage/common.go
package storage

import (
	"fmt"
	"strings"
)

// =============================================================================
// PostgreSQL Array Operations
// =============================================================================

// SqlArrayContains checks if array column contains all specified values.
// PostgreSQL: column @> ARRAY[values]::type
type SqlArrayContains struct {
	Field     string
	Values    []any
	ValueType string // e.g., "uuid[]", "text[]", "int[]"
}

// NewSqlArrayContains creates a new array contains condition.
func NewSqlArrayContains(field string, values []any, valueType string) SqlArrayContains {
	return SqlArrayContains{Field: field, Values: values, ValueType: valueType}
}

// ToSql implements sq.Sqlizer interface.
func (s SqlArrayContains) ToSql() (string, []any, error) {
	if len(s.Values) == 0 {
		return "TRUE", nil, nil
	}

	placeholders := make([]string, len(s.Values))
	for i := range s.Values {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	sql := fmt.Sprintf("%s @> ARRAY[%s]::%s",
		s.Field,
		strings.Join(placeholders, ","),
		s.ValueType,
	)

	return sql, s.Values, nil
}

// SqlArrayIsContainedBy checks if array column is contained by specified values.
// PostgreSQL: column <@ ARRAY[values]::type
type SqlArrayIsContainedBy struct {
	Field     string
	Values    []any
	ValueType string
}

// NewSqlArrayIsContainedBy creates a new array "is contained by" condition.
func NewSqlArrayIsContainedBy(field string, values []any, valueType string) SqlArrayIsContainedBy {
	return SqlArrayIsContainedBy{Field: field, Values: values, ValueType: valueType}
}

// ToSql implements sq.Sqlizer interface.
func (s SqlArrayIsContainedBy) ToSql() (string, []any, error) {
	if len(s.Values) == 0 {
		return "TRUE", nil, nil
	}

	placeholders := make([]string, len(s.Values))
	for i := range s.Values {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	sql := fmt.Sprintf("%s <@ ARRAY[%s]::%s",
		s.Field,
		strings.Join(placeholders, ","),
		s.ValueType,
	)

	return sql, s.Values, nil
}

// SqlArrayOverlap checks if array column overlaps with specified values.
// PostgreSQL: column && ARRAY[values]::type
type SqlArrayOverlap struct {
	Field     string
	Values    []any
	ValueType string
}

// NewSqlArrayOverlap creates a new array overlap condition.
func NewSqlArrayOverlap(field string, values []any, valueType string) SqlArrayOverlap {
	return SqlArrayOverlap{Field: field, Values: values, ValueType: valueType}
}

// ToSql implements sq.Sqlizer interface.
func (s SqlArrayOverlap) ToSql() (string, []any, error) {
	if len(s.Values) == 0 {
		return "FALSE", nil, nil
	}

	placeholders := make([]string, len(s.Values))
	for i := range s.Values {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
	}

	sql := fmt.Sprintf("%s && ARRAY[%s]::%s",
		s.Field,
		strings.Join(placeholders, ","),
		s.ValueType,
	)

	return sql, s.Values, nil
}

// =============================================================================
// Subquery Operations
// =============================================================================

// InSubQuery creates an IN (subquery) condition.
// PostgreSQL: field IN (SELECT ...)
type InSubQuery struct {
	Field    string
	Subquery string
	Args     []any
}

// NewInSubQuery creates a new IN subquery condition.
func NewInSubQuery(field, subquery string, args ...any) InSubQuery {
	return InSubQuery{Field: field, Subquery: subquery, Args: args}
}

// ToSql implements sq.Sqlizer interface.
func (s InSubQuery) ToSql() (string, []any, error) {
	sql := fmt.Sprintf("%s IN (%s)", s.Field, s.Subquery)
	return sql, s.Args, nil
}

// NotInSubQuery creates a NOT IN (subquery) condition.
type NotInSubQuery struct {
	Field    string
	Subquery string
	Args     []any
}

// NewNotInSubQuery creates a new NOT IN subquery condition.
func NewNotInSubQuery(field, subquery string, args ...any) NotInSubQuery {
	return NotInSubQuery{Field: field, Subquery: subquery, Args: args}
}

// ToSql implements sq.Sqlizer interface.
func (s NotInSubQuery) ToSql() (string, []any, error) {
	sql := fmt.Sprintf("%s NOT IN (%s)", s.Field, s.Subquery)
	return sql, s.Args, nil
}

// =============================================================================
// JSON Operations
// =============================================================================

// SqlJSONContains checks if JSONB column contains specified value.
// PostgreSQL: column @> 'value'::jsonb
type SqlJSONContains struct {
	Field string
	Value string // JSON string
}

// NewSqlJSONContains creates a new JSONB contains condition.
func NewSqlJSONContains(field, value string) SqlJSONContains {
	return SqlJSONContains{Field: field, Value: value}
}

// ToSql implements sq.Sqlizer interface.
func (s SqlJSONContains) ToSql() (string, []any, error) {
	sql := fmt.Sprintf("%s @> $1::jsonb", s.Field)
	return sql, []any{s.Value}, nil
}

// =============================================================================
// Usage in getCondition
// =============================================================================

// Example usage in repository:
//
//	func (r *gameRepo) getGameCondition(filter *GameFilter) []sq.Sqlizer {
//	    conditions := make([]sq.Sqlizer, 0)
//
//	    if filter == nil {
//	        return conditions
//	    }
//
//	    // Regular equality
//	    if len(filter.ID) > 0 {
//	        conditions = append(conditions, sq.Eq{"g.id": filter.ID})
//	    }
//
//	    // Array overlap (any match)
//	    if len(filter.Label) > 0 {
//	        labels := make([]any, len(filter.Label))
//	        for i, l := range filter.Label {
//	            labels[i] = l
//	        }
//	        conditions = append(conditions, NewSqlArrayOverlap("g.labels", labels, "text[]"))
//	    }
//
//	    // Array contains (all must match)
//	    if len(filter.RequiredFeatures) > 0 {
//	        features := make([]any, len(filter.RequiredFeatures))
//	        for i, f := range filter.RequiredFeatures {
//	            features[i] = f
//	        }
//	        conditions = append(conditions, NewSqlArrayContains("g.features", features, "text[]"))
//	    }
//
//	    return conditions
//	}
