package subscriber

import (
	"context"
	"errors"
	"sync"

	"github.com/aawadallak/go-core-kit/core/broker"
)

// Errors
var (
	ErrUnexpectedSchema = errors.New("unexpected schema")
	ErrSubscriberClosed = errors.New("subscriber closed")
)

// Hook defines a function type for processing messages
type Hook[T any] func(ctx context.Context, val T) error

// Handler manages subscription and message processing
type Handler[T any] struct {
	subscribe broker.Subscriber
	handler   Hook[T]
	wg        sync.WaitGroup
	stop      chan struct{}
}

// NewHandler creates a new Handler instance
func NewHandler[T any](subscribe broker.Subscriber, fn Hook[T]) *Handler[T] {
	return &Handler[T]{
		subscribe: subscribe,
		handler:   fn,
		stop:      make(chan struct{}),
	}
}

// Start begins message processing with configurable workers
func (h *Handler[T]) Start(ctx context.Context, workerCount int) error {
	if workerCount < 1 {
		workerCount = 1
	}

	h.wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go h.worker(ctx)
	}

	return nil
}

// Stop gracefully shuts down the handler
func (h *Handler[T]) Stop() {
	close(h.stop)
	h.wg.Wait()
}

func (h *Handler[T]) worker(ctx context.Context) {
	defer h.wg.Done()

	for {
		select {
		case <-h.stop:
			return
		case <-ctx.Done():
			return
		default:
			msg, err := h.subscribe.Subscribe(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}
				continue
			}

			val, ok := msg.Payload().(*T)
			if !ok {
				continue
			}

			if err := h.handler(ctx, *val); err != nil {
				continue
			}

			if err := h.subscribe.Commit(ctx, msg); err != nil {
				continue
			}
		}
	}
}
