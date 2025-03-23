package idem

import (
	"github.com/aawadallak/go-core-kit/core/idempotent"
)

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

// NewHandler creates a new Handler instance with the specified options.
// It initializes the Handler with default JSON codecs and requires a repository
// to be set via options (or it fails). The functional options pattern allows
// flexible configuration of the handler’s components.
//
// Parameters:
//   - repository: The Repository to use for storing and retrieving idempotent events.
//   - opts: Variadic list of Option functions to configure the Handler.
//
// Returns:
//   - *Handler[T]: A configured Handler instance.
//   - error: An error if the repository is not provided (ErrRepository).
func NewHandler[T any](repository idempotent.Repository, opts ...Option[T]) (*Handler[T], error) {
	h := &Handler[T]{
		encoder:    idempotent.NewGzipJSONEncoder(),
		decoder:    idempotent.NewGzipJSONDecoder[T](),
		repository: repository,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h, nil
}
