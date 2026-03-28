package job

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aawadallak/go-core-kit/core/logger"
	"github.com/aawadallak/go-core-kit/pkg/common"
)

type Orchestrator struct {
	Repo     Repository
	Handlers map[string]Handler
	Interval time.Duration
	stopChan chan struct{}
	stopOnce sync.Once
	stopping atomic.Bool
	workers  sync.WaitGroup
	runDone  chan struct{}
}

type OrchestratorParams struct {
	Repo     Repository
	Interval time.Duration
	Handlers []Handler
}

func NewOrchestrator(params OrchestratorParams) *Orchestrator {
	handlerMap := make(map[string]Handler)
	for _, h := range params.Handlers {
		handlerMap[h.GetType()] = h
	}
	return &Orchestrator{
		Repo:     params.Repo,
		Handlers: handlerMap,
		Interval: params.Interval,
		stopChan: make(chan struct{}),
		runDone:  make(chan struct{}),
	}
}

func (o *Orchestrator) Run(ctx context.Context) error {
	defer close(o.runDone)

	ticker := time.NewTicker(o.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Of(ctx).DebugS("Orchestrator stopped due to context cancellation")
			return ctx.Err()
		case <-o.stopChan:
			logger.Of(ctx).DebugS("Orchestrator stopped")
			return nil
		case <-ticker.C:
			if o.stopping.Load() {
				continue
			}

			o.workers.Add(1)
			go func() {
				defer o.workers.Done()
				if o.stopping.Load() {
					return
				}

				txFn := func(ctx context.Context) error {
					return o.ProcessJobs(ctx)
				}

				if err := o.Repo.WithTransaction(ctx, txFn); err != nil {
					logger.Of(ctx).ErrorS("Failed to process jobs", logger.WithValue("error", err))
				}
			}()
		}
	}
}

func (o *Orchestrator) Stop() {
	o.stopOnce.Do(func() {
		o.stopping.Store(true)
		close(o.stopChan)
	})
}

func (o *Orchestrator) StopAndWait(ctx context.Context) error {
	o.Stop()

	select {
	case <-o.runDone:
	case <-ctx.Done():
		return ctx.Err()
	}

	workersDone := make(chan struct{})
	go func() {
		defer close(workersDone)
		o.workers.Wait()
	}()

	select {
	case <-workersDone:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (o *Orchestrator) ProcessJobs(ctx context.Context) error {
	job, err := o.Repo.GetNextJob(ctx)
	if err != nil {
		if errors.Is(err, common.ErrResourceNotFound) {
			return nil // No more jobs to process
		}

		return err
	}

	if job == nil {
		logger.Of(ctx).DebugS("No jobs available")
		return nil // No more jobs to process
	}

	job.Error = ""

	// Find the handler for the job type
	handler, ok := o.Handlers[job.Type]
	if !ok {
		job.Status = JobStatusFailed
		job.Error = "no handler for job type: " + job.Type
		return o.Repo.UpdateJob(ctx, job)
	}

	logger.Of(ctx).DebugS("Processing job",
		logger.WithValue("correlation_id", job.CorrelationID),
		logger.WithValue("type", job.Type))

	handlerErr := func() (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				err = fmt.Errorf("job handler panic: %v", recovered)
			}
		}()

		return handler.Process(ctx, job)
	}()

	if handlerErr != nil {
		logger.Of(ctx).ErrorS("Job handler failed",
			logger.WithValue("correlation_id", job.CorrelationID),
			logger.WithValue("type", job.Type),
			logger.WithValue("error", handlerErr))
		job.Error = fmt.Sprintf("Job processing failed: %v", handlerErr)
		job.Status = JobStatusFailed
	}

	if job.Error == "" {
		job.Status = JobStatusCompleted
		logger.Of(ctx).DebugS("Job processed",
			logger.WithValue("correlation_id", job.CorrelationID),
			logger.WithValue("status", job.Status),
		)
	}

	if err := o.Repo.UpdateJob(ctx, job); err != nil {
		return err
	}

	return nil
}
