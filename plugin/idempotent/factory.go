package idempotent

import (
	"errors"

	"github.com/aawadallak/go-core-kit/core/idempotent"
)

// ErrRepository is returned when a repository is not provided to the Handler.
var ErrRepository = errors.New("repository is required")

// Handler implements the idempotent.Handler interface for managing idempotent operations.
// It uses an encoder to serialize results, a decoder to deserialize stored payloads,
// and a repository to persist and retrieve event data.
type Handler[T any] struct {
	// encoder converts the result of a Hook into a byte slice for storage.
	encoder idempotent.Encoder

	// decoder converts a stored byte slice back into the original result type T.
	decoder idempotent.Decoder[T]

	// repository provides persistence for idempotent events, allowing the handler
	// to check for prior executions and store new results.
	repository idempotent.Repository
}

// Option defines a function type for configuring a Handler instance.
// It allows customization of the encoder, decoder, and repository via the
// functional options pattern.
type Option[T any] func(*Handler[T])

// WithEncoder returns an Option that sets the encoder for the Handler.
// This allows customization of how results are serialized into byte slices.
//
// Parameters:
//   - encoder: The Encoder to use for serializing Hook results.
func WithEncoder[T any](encoder idempotent.Encoder) Option[T] {
	return func(h *Handler[T]) {
		h.encoder = encoder
	}
}

// WithDecoder returns an Option that sets the decoder for the Handler.
// This allows customization of how stored payloads are deserialized into type T.
//
// Parameters:
//   - decoder: The Decoder to use for deserializing stored payloads.
func WithDecoder[T any](decoder idempotent.Decoder[T]) Option[T] {
	return func(h *Handler[T]) {
		h.decoder = decoder
	}
}

// WithRepository returns an Option that sets the repository for the Handler.
// This specifies the storage backend for persisting and retrieving idempotent events.
//
// Parameters:
//   - repository: The Repository to use for event storage and retrieval.
func WithRepository[T any](repository idempotent.Repository) Option[T] {
	return func(h *Handler[T]) {
		h.repository = repository
	}
}

// NewHandler creates a new Handler instance with the specified options.
// It initializes the Handler with default JSON codecs and requires a repository
// to be set via options (or it fails). The functional options pattern allows
// flexible configuration of the handler’s components.
//
// Parameters:
//   - opts: Variadic list of Option functions to configure the Handler.
//
// Returns:
//   - *Handler[T]: A configured Handler instance.
//   - error: An error if the repository is not provided (ErrRepository).
func NewHandler[T any](opts ...Option[T]) (*Handler[T], error) {
	// Initialize with default JSON encoder and decoder
	h := &Handler[T]{
		encoder:    idempotent.NewJSONEncoder(),    // Default to JSON encoding
		decoder:    idempotent.NewJSONDecoder[T](), // Default to JSON decoding
		repository: nil,                            // Repository must be set via options
	}

	// Apply each provided option to customize the Handler
	for _, opt := range opts {
		opt(h)
	}

	// Ensure a repository is provided, as it’s required for idempotency
	if h.repository == nil {
		return nil, ErrRepository // ErrRepository should be defined elsewhere
	}

	return h, nil
}
