package idempotent

import "context"

// EventItem represents an idempotent event stored in the repository.
// It encapsulates a unique key and its associated payload, which can be
// used to track and replay operations for idempotency.
type EventItem struct {
	// Key is a unique identifier for the idempotent event.
	// It ensures that the same operation can be recognized and handled
	// consistently across multiple invocations.
	Key string

	// Payload contains the serialized data associated with the event.
	// This could be the result of an operation or metadata needed to
	// reconstruct the operation’s outcome.
	Payload []byte
}

// Hook defines a function type that represents the core operation to be
// executed in an idempotent manner. It takes a context for cancellation
// and timeout support and returns a result of type T or an error.
// This is typically the business logic that needs idempotency guarantees.
type Hook[T any] func(ctx context.Context) (T, error)

// Handler defines the interface for an idempotent handler.
// It provides a method to wrap a Hook with idempotency logic, ensuring
// that the same operation (identified by a key) produces consistent results
// without unintended side effects on repeated calls.
type Handler[T any] interface {
	// Wrap executes the provided Hook function with idempotency guarantees.
	// It uses the given key to check if the operation has been performed before,
	// returning the cached result if available, or executing and storing the result
	// if not. The context is passed through to the Hook for cancellation and deadlines.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeouts, and request-scoped values.
	//   - key: Unique identifier for the idempotent operation.
	//   - fn: The Hook function to execute idempotently.
	//
	// Returns:
	//   - T: The result of the operation.
	//   - error: Any error encountered during execution or retrieval.
	Wrap(ctx context.Context, key string, fn Hook[T]) (T, error)
}

// Repository defines the interface for persisting and retrieving idempotent events.
// It abstracts the storage layer, allowing the idempotent handler to work with
// different databases (e.g., PostgreSQL, MySQL, in-memory) without modification.
type Repository interface {
	// Find retrieves an EventItem from the repository by its key.
	// This is used to check if an operation has already been completed.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeouts, and request-scoped values.
	//   - key: The unique identifier of the event to retrieve.
	//
	// Returns:
	//   - *EventItem: The stored event item, or nil if not found.
	//   - error: Any error encountered during retrieval (e.g., not found, database error).
	Find(ctx context.Context, key string) (*EventItem, error)

	// Save stores an EventItem in the repository.
	// This persists the result of an operation so it can be reused on subsequent calls.
	//
	// Parameters:
	//   - ctx: Context for cancellation, timeouts, and request-scoped values.
	//   - item: The EventItem to store, containing the key and payload.
	//
	// Returns:
	//   - error: Any error encountered during storage (e.g., duplicate key, database error).
	Save(ctx context.Context, item EventItem) error
}
