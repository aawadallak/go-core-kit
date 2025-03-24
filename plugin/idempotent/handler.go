package idem

import (
	"context"

	"github.com/aawadallak/go-core-kit/core/idempotent"
)

// Wrap executes a function idempotently using a key to ensure consistent results.
// If the key exists in the repository, it returns the cached result. Otherwise,
// it executes the provided hook function, stores the result, and returns it.
//
// Type Parameters:
//
//	T - The type of the value being processed and returned
//
// Parameters:
//
//	ctx - Context for cancellation and deadlines
//	key - Unique identifier for the idempotent operation
//	fn  - Hook function to execute if the key is not found
//
// Returns:
//
//	T     - The result of the operation (cached or newly computed)
//	error - Any error encountered during execution or storage
//
// Errors:
//
//	Decoder errors     - If decoding the cached payload fails
//	Hook function errors - If the provided fn returns an error
//	Encoder errors     - If encoding the result fails
//	Repository Save errors - If storing the new item fails
func (h *Handler[T]) Wrap(ctx context.Context, key string, fn idempotent.Hook[T]) (T, error) {
	// Attempt to retrieve the existing item from the repository
	item, err := h.repository.Find(ctx, key)
	if err == nil {
		return h.decoder(item.Payload)
	}

	// If not found, execute the provided hook function
	val, err := fn(ctx)
	if err != nil {
		// Return the zero value of T and the error if the hook fails
		return val, err
	}

	// Encode the result for storage
	payload, err := h.encoder(val)
	if err != nil {
		// If encoding fails, return the value anyway with no error to maintain idempotency
		return val, nil
	}

	// Create an event item with the key and encoded payload
	idemItem := idempotent.EventItem{
		Key:     key,
		Payload: payload,
	}

	// Store the item in the repository
	if err := h.repository.Save(ctx, idemItem); err != nil {
		// If saving fails, still return the value to ensure the caller gets the result
		return val, err
	}

	// Return the successfully computed and stored value
	return val, nil
}
