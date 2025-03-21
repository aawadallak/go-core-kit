package idempotent

import (
	"context"

	"github.com/aawadallak/go-core-kit/core/idempotent"
)

func (h *Handler[T]) Wrap(ctx context.Context, key string, fn idempotent.Hook[T]) (T, error) {
	item, err := h.repository.Find(ctx, key)
	if err == nil {
		return h.decoder(item.Payload)
	}

	val, err := fn(ctx)
	if err != nil {
		return val, err
	}

	payload, err := h.encoder(val)
	if err != nil {
		return val, nil
	}

	idemItem := idempotent.EventItem{
		Key:     key,
		Payload: payload,
	}

	if err := h.repository.Save(ctx, idemItem); err != nil {
		return val, err
	}

	return val, nil
}
