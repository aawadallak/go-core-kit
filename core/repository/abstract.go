package repository

import "context"

// AbstractRepositoryEntity represents an entity that can be stored in a repository.
// It requires implementing a GetID method to uniquely identify the entity.
type AbstractRepositoryEntity interface {
	GetID() uint
}

// AbstractRepository defines the common CRUD operations for a repository.
// It is a generic interface that works with any type T that implements AbstractRepositoryEntity.
type AbstractRepository[T AbstractRepositoryEntity] interface {
	// FindOne retrieves a single entity matching the provided entity's criteria.
	FindOne(ctx context.Context, entity *T) (*T, error)

	// Find retrieves all entities matching the provided entity's criteria.
	Find(ctx context.Context, entity *T) ([]*T, error)

	// Save creates a new entity in the repository.
	Save(ctx context.Context, entity *T) (*T, error)

	// Update modifies an existing entity in the repository.
	Update(ctx context.Context, entity *T) (*T, error)

	// Delete removes an entity from the repository.
	Delete(ctx context.Context, entity *T) error

	// Tx executes a function within a transaction context.
	// If the function returns an error, the transaction is rolled back.
	Tx(ctx context.Context, fn func(ctx context.Context) error) error
}

// AbstractPaginatedRepository defines operations for paginated data retrieval.
// It is a generic interface that works with any types T (result type) and E (query type).
type AbstractPaginatedRepository[T any, E any] interface {
	// FindAll retrieves a paginated list of entities based on the provided pagination query.
	FindAll(ctx context.Context, entity *PaginationQuery[E]) (*Pagination[T], error)
}
