package worker

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrWorkerClosed = errors.New("worker: closed")

// AbstractWorker provides a base implementation of the Worker interface.
type AbstractWorker struct {
	once         sync.Once
	done         chan struct{}
	provider     TaskProvider
	abortOnError bool
	synchronous  bool
	isClosed     bool
	mu           sync.Mutex
}

// NewAbstractWorker creates a new AbstractWorker with the given provider and options.
func NewAbstractWorker(provider TaskProvider, opts ...Option) *AbstractWorker {
	w := &AbstractWorker{
		done:     make(chan struct{}),
		provider: provider,
	}

	for _, opt := range opts {
		opt(w)
	}

	return w
}

// Start begins the worker's execution, either synchronously or asynchronously based on configuration.
func (w *AbstractWorker) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isClosed {
		return ErrWorkerClosed
	}

	select {
	case <-w.done:
		return context.Canceled
	default:
	}

	if starter, ok := w.provider.(OnStarter); ok {
		if err := starter.OnStart(ctx); err != nil {
			return err
		}
	}

	interval := w.provider.Interval()
	if interval <= 0 {
		interval = NoInterval
	}

	run := func() {
		defer func() {
			w.once.Do(func() {
				if closer, ok := w.provider.(OnShutdowner); ok {
					if err := closer.OnShutdown(ctx); err != nil {
						return
					}
				}
				w.isClosed = true
			})
		}()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.done:
				return
			case <-ticker.C:
				w.executeHandler(ctx)
			}
		}
	}

	if w.synchronous {
		run()
	} else {
		go run()
	}

	return nil
}

// Close shuts down the worker, ensuring cleanup is performed only once.
func (w *AbstractWorker) Close(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-w.done:
		return nil
	default:
		var closeErr error
		w.once.Do(func() {
			close(w.done)
			if closer, ok := w.provider.(OnShutdowner); ok {
				if err := closer.OnShutdown(ctx); err != nil {
					closeErr = err
				}
			}
			w.isClosed = true
		})
		return closeErr
	}
}

// executeHandler runs the provider's execution logic with optional pre/post hooks.
func (w *AbstractWorker) executeHandler(ctx context.Context) {
	if before, ok := w.provider.(OnPreExecutor); ok {
		if err := before.OnPreExecute(ctx); err != nil {
			if w.abortOnError {
				close(w.done)
			}
			return
		}
	}

	if err := w.provider.Execute(ctx); err != nil {
		if w.abortOnError {
			close(w.done)
		}
		return
	}

	if after, ok := w.provider.(OnPostExecutor); ok {
		if err := after.OnPostExecute(ctx); err != nil {
			if w.abortOnError {
				close(w.done)
			}
		}
	}
}
