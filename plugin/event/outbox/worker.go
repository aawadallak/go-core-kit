package outbox

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aawadallak/go-core-kit/core/logger"
	cevent "github.com/aawadallak/go-core-kit/core/event"
)

const (
	defaultBatchSize    = 50
	defaultPollInterval = 1 * time.Second
	defaultMaxRetries   = 10
	defaultCleanupAge   = 7 * 24 * time.Hour
	cleanupInterval     = 1 * time.Hour
)

type fetcher interface {
	FetchPending(ctx context.Context, batchSize int, eventNames []string) ([]Entry, error)
	MarkSent(ctx context.Context, id uint) error
	MarkFailed(ctx context.Context, id uint, errMsg string) error
	MarkExhausted(ctx context.Context, id uint) error
	CleanupSent(ctx context.Context, olderThan time.Duration) (int64, error)
}

type WorkerDependencies struct {
	Repo         fetcher
	Dispatcher   cevent.Dispatcher
	EventNames   []string
	PollInterval time.Duration
	BatchSize    int
}

type Worker struct {
	repo         fetcher
	dispatcher   cevent.Dispatcher
	eventNames   []string
	pollInterval time.Duration
	batchSize    int
}

func NewWorker(deps WorkerDependencies) *Worker {
	if deps.PollInterval == 0 {
		deps.PollInterval = defaultPollInterval
	}
	if deps.BatchSize == 0 {
		deps.BatchSize = defaultBatchSize
	}
	return &Worker{
		repo:         deps.Repo,
		dispatcher:   deps.Dispatcher,
		eventNames:   deps.EventNames,
		pollInterval: deps.PollInterval,
		batchSize:    deps.BatchSize,
	}
}

func (w *Worker) Start(ctx context.Context) {
	pollTicker := time.NewTicker(w.pollInterval)
	cleanupTicker := time.NewTicker(cleanupInterval)
	defer pollTicker.Stop()
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pollTicker.C:
			w.processBatch(ctx)
		case <-cleanupTicker.C:
			w.cleanup(ctx)
		}
	}
}

func (w *Worker) processBatch(ctx context.Context) {
	entries, err := w.repo.FetchPending(ctx, w.batchSize, w.eventNames)
	if err != nil {
		logger.Of(ctx).ErrorS("outbox.worker.fetch_failed", logger.WithValue("error", err))
		return
	}

	for i := range entries {
		if entries[i].RetryCount >= defaultMaxRetries {
			if err := w.repo.MarkExhausted(ctx, entries[i].ID); err != nil {
				logger.Of(ctx).ErrorS("outbox.worker.mark_exhausted_failed",
					logger.WithValue("entry_id", entries[i].ID),
					logger.WithValue("error", err))
			}
			logger.Of(ctx).ErrorS("outbox.worker.event_exhausted",
				logger.WithValue("entry_id", entries[i].ID),
				logger.WithValue("event_name", entries[i].EventName),
				logger.WithValue("retry_count", entries[i].RetryCount))
			continue
		}

		record := &cevent.Record{
			ID:            entries[i].EventID,
			Name:          entries[i].EventName,
			Version:       entries[i].EventVersion,
			Metadata:      json.RawMessage(entries[i].Payload),
			CorrelationID: entries[i].CorrelationID,
			RequestID:     entries[i].RequestID,
			TraceID:       entries[i].TraceID,
			SpanID:        entries[i].SpanID,
			Timestamp:     entries[i].CreatedAt,
		}

		if err := w.dispatcher.Dispatch(ctx, record); err != nil {
			if markErr := w.repo.MarkFailed(ctx, entries[i].ID, err.Error()); markErr != nil {
				logger.Of(ctx).ErrorS("outbox.worker.mark_failed_failed",
					logger.WithValue("entry_id", entries[i].ID),
					logger.WithValue("error", markErr))
			}
			logger.Of(ctx).WarnS("outbox.worker.dispatch_failed",
				logger.WithValue("entry_id", entries[i].ID),
				logger.WithValue("event_name", entries[i].EventName),
				logger.WithValue("retry_count", entries[i].RetryCount),
				logger.WithValue("error", err))
			continue
		}

		if err := w.repo.MarkSent(ctx, entries[i].ID); err != nil {
			logger.Of(ctx).ErrorS("outbox.worker.mark_sent_failed",
				logger.WithValue("entry_id", entries[i].ID),
				logger.WithValue("error", err))
		}
	}
}

func (w *Worker) cleanup(ctx context.Context) {
	deleted, err := w.repo.CleanupSent(ctx, defaultCleanupAge)
	if err != nil {
		logger.Of(ctx).ErrorS("outbox.worker.cleanup_failed", logger.WithValue("error", err))
		return
	}
	if deleted > 0 {
		logger.Of(ctx).InfoS("outbox.worker.cleanup_done", logger.WithValue("deleted", deleted))
	}
}
