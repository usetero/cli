package api

// OrderDirection specifies the direction for ordering results.
// This is a common type used across all list operations.
type OrderDirection string

const (
	OrderDirectionAsc  OrderDirection = "ASC"
	OrderDirectionDesc OrderDirection = "DESC"
)
