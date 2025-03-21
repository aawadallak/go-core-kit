package idempotent

import (
	"github.com/aawadallak/go-core-kit/core/idempotent"
)

type Handler[T any] struct {
	encoder    idempotent.Encoder
	decoder    idempotent.Decoder[T]
	repository idempotent.Repository
}

type Option[T any] func(*Handler[T])

func WithEncoder[T any](encoder idempotent.Encoder) Option[T] {
	return func(h *Handler[T]) {
		h.encoder = encoder
	}
}

func WithDecoder[T any](decoder idempotent.Decoder[T]) Option[T] {
	return func(h *Handler[T]) {
		h.decoder = decoder
	}
}

func WithRepository[T any](repository idempotent.Repository) Option[T] {
	return func(h *Handler[T]) {
		h.repository = repository
	}
}

func NewHandler[T any](opts ...Option[T]) (*Handler[T], error) {
	h := &Handler[T]{
		encoder:    idempotent.NewJSONEncoder(),
		decoder:    idempotent.NewJSONDecoder[T](),
		repository: nil, // This should be required in practice
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.repository == nil {
		return nil, ErrRepository
	}

	return h, nil
}
