package audit

import (
	"context"
	"sync"
	"time"

	"github.com/aawadallak/go-core-kit/core/logger"
)

type Orchestrator struct {
	Stream        chan Log
	once          sync.Once
	batch         []Log
	batchSize     int
	batchInterval time.Duration
	batchLock     sync.Mutex
	provider      Provider
}

// Close implements audit.Provider.
func (o *Orchestrator) Close(ctx context.Context) error {
	o.once.Do(func() {
		close(o.Stream)
	})

	if err := ctx.Err(); err != nil {
		return err
	}

	return nil
}

// Dispatch implements audit.Provider.
func (o *Orchestrator) Dispatch(ctx context.Context, log *Log) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		o.Stream <- *log
	}

	return nil
}

func (o *Orchestrator) Flush(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if len(o.batch) == 0 {
			return nil
		}

		o.batchLock.Lock()
		if err := o.provider.Flush(ctx, o.batch...); err != nil {
			return err
		}

		o.batch = make([]Log, 0, o.batchSize)
		o.batchLock.Unlock()
	}

	return nil
}

func (o *Orchestrator) start(ctx context.Context) {
	ticker := time.NewTicker(o.batchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Of(ctx).InfoS("Audit::Provider::dispatch", logger.
				WithValue("message", "context done"))
			return
		case log, ok := <-o.Stream:
			if !ok {
				if err := o.Flush(ctx); err != nil {
					logger.Of(ctx).ErrorS("Audit::Provider::FinalFlush",
						logger.WithValue("error", err))
				}
				return
			}

			o.batchLock.Lock()
			o.batch = append(o.batch, log)
			batchLen := len(o.batch)
			o.batchLock.Unlock()

			if batchLen >= o.batchSize {
				if err := o.Flush(ctx); err != nil {
					logger.Of(ctx).ErrorS(
						"Audit::Provider::Flush",
						logger.WithValue("error", err),
					)
				}
			}
		case <-ticker.C:
			if err := o.Flush(ctx); err != nil {
				logger.Of(ctx).ErrorS(
					"Audit::Provider::Flush",
					logger.WithValue("error", err),
				)
			}
		}
	}
}

func NewOrchestrator(ctx context.Context, opts ...Options) *Orchestrator {
	orchestrator := &Orchestrator{
		Stream:        make(chan Log),
		batchSize:     1,
		batchInterval: 10 * time.Second,
		provider:      NewStandardProvider(),
	}

	for _, opt := range opts {
		opt(orchestrator)
	}

	orchestrator.batch = make([]Log, 0, orchestrator.batchSize)

	go orchestrator.start(ctx)

	return orchestrator
}
