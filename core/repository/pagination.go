package repository

// PaginationQuery represents a query for paginated data retrieval.
// It contains parameters for pagination, ordering, and filtering.
type PaginationQuery[T any] struct {
	// Page specifies the current page number (1-based)
	Page int
	// PerPage specifies the number of items per page
	PerPage int
	// Order specifies how the results should be sorted
	Order OrderBy
	// Query contains the filtering criteria
	Query *T
}

// OrderBy defines the sorting parameters for query results.
type OrderBy struct {
	// Direction specifies the sort direction (e.g., "asc" or "desc")
	Direction string
	// Field specifies which field to sort by
	Field string
}

// Pagination represents the paginated response structure.
type Pagination[T any] struct {
	// Page is the current page number
	Page int `json:"page"`
	// Pages is the total number of pages
	Pages int `json:"pages"`
	// Total is the total number of items across all pages
	Total int `json:"total"`
	// Data contains the items for the current page
	Data []T `json:"data"`
}
