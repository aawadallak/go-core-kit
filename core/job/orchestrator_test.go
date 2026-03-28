package job

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aawadallak/go-core-kit/pkg/common"
)

type testHandler struct {
	typ string
	fn  func(ctx context.Context, job *Job) error
}

func (h testHandler) Process(ctx context.Context, job *Job) error {
	return h.fn(ctx, job)
}

func (h testHandler) GetType() string {
	return h.typ
}

type testRepo struct {
	mu        sync.Mutex
	jobs      []*Job
	updated   []*Job
	getNextN  atomic.Int64
	updateJob atomic.Int64
}

func (r *testRepo) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (r *testRepo) AppendToQueue(ctx context.Context, jobs ...Job) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range jobs {
		jobCopy := jobs[i]
		r.jobs = append(r.jobs, &jobCopy)
	}
	return nil
}

func (r *testRepo) GetNextJob(ctx context.Context) (*Job, error) {
	r.getNextN.Add(1)

	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.jobs) == 0 {
		return nil, common.ErrResourceNotFound
	}

	next := r.jobs[0]
	r.jobs = r.jobs[1:]
	return next, nil
}

func (r *testRepo) UpdateJob(ctx context.Context, job *Job) error {
	r.updateJob.Add(1)

	r.mu.Lock()
	defer r.mu.Unlock()

	jobCopy := *job
	r.updated = append(r.updated, &jobCopy)
	return nil
}

func (r *testRepo) getNextCount() int64 {
	return r.getNextN.Load()
}

func TestOrchestratorStopAndWait_WaitsInflightJob(t *testing.T) {
	repo := &testRepo{
		jobs: []*Job{{Type: "slow", Status: JobStatusPending}},
	}

	started := make(chan struct{})
	release := make(chan struct{})
	handler := testHandler{
		typ: "slow",
		fn: func(ctx context.Context, job *Job) error {
			close(started)
			<-release
			return nil
		},
	}

	orchestrator := NewOrchestrator(OrchestratorParams{
		Repo:     repo,
		Interval: 5 * time.Millisecond,
		Handlers: []Handler{handler},
	})

	runCtx := t.Context()
	go func() { _ = orchestrator.Run(runCtx) }()

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("job handler did not start")
	}

	stopDone := make(chan error, 1)
	go func() {
		stopDone <- orchestrator.StopAndWait(context.Background())
	}()

	select {
	case err := <-stopDone:
		t.Fatalf("StopAndWait returned before in-flight job finished: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(release)

	select {
	case err := <-stopDone:
		if err != nil {
			t.Fatalf("StopAndWait returned error: %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("StopAndWait timed out after releasing in-flight job")
	}
}

func TestOrchestratorStopAndWait_DoesNotAcceptNewWorkAfterStop(t *testing.T) {
	repo := &testRepo{}
	for range 200 {
		repo.jobs = append(repo.jobs, &Job{Type: "work", Status: JobStatusPending})
	}

	var processed atomic.Int64
	handler := testHandler{
		typ: "work",
		fn: func(ctx context.Context, job *Job) error {
			processed.Add(1)
			time.Sleep(10 * time.Millisecond)
			return nil
		},
	}

	orchestrator := NewOrchestrator(OrchestratorParams{
		Repo:     repo,
		Interval: 1 * time.Millisecond,
		Handlers: []Handler{handler},
	})

	runCtx := t.Context()
	go func() { _ = orchestrator.Run(runCtx) }()

	deadline := time.Now().Add(500 * time.Millisecond)
	for processed.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if processed.Load() == 0 {
		t.Fatal("expected at least one processed job before stopping")
	}

	if err := orchestrator.StopAndWait(context.Background()); err != nil {
		t.Fatalf("StopAndWait returned error: %v", err)
	}

	processedAfterStop := processed.Load()
	getNextAfterStop := repo.getNextCount()

	time.Sleep(50 * time.Millisecond)

	if got := processed.Load(); got != processedAfterStop {
		t.Fatalf("processed jobs increased after StopAndWait returned: before=%d after=%d", processedAfterStop, got)
	}
	if got := repo.getNextCount(); got != getNextAfterStop {
		t.Fatalf("GetNextJob calls increased after StopAndWait returned: before=%d after=%d", getNextAfterStop, got)
	}
}

func TestOrchestratorProcessJobs_PanicMarksJobAsFailed(t *testing.T) {
	repo := &testRepo{
		jobs: []*Job{{Type: "panic", Status: JobStatusPending}},
	}

	handler := testHandler{
		typ: "panic",
		fn: func(ctx context.Context, job *Job) error {
			panic("boom")
		},
	}

	orchestrator := NewOrchestrator(OrchestratorParams{
		Repo:     repo,
		Interval: 10 * time.Millisecond,
		Handlers: []Handler{handler},
	})

	err := orchestrator.ProcessJobs(context.Background())
	if err != nil {
		t.Fatalf("ProcessJobs returned unexpected error: %v", err)
	}

	repo.mu.Lock()
	defer repo.mu.Unlock()

	if len(repo.updated) != 1 {
		t.Fatalf("expected exactly one updated job, got %d", len(repo.updated))
	}

	updated := repo.updated[0]
	if updated.Status != JobStatusFailed {
		t.Fatalf("expected failed status, got %s", updated.Status)
	}
	if updated.Error == "" {
		t.Fatal("expected non-empty job error after panic")
	}
	if !strings.Contains(updated.Error, "panic") || !strings.Contains(updated.Error, "boom") {
		t.Fatalf("expected panic details in job error, got: %s", updated.Error)
	}
}

func TestOrchestratorProcessJobs_NoJobsIsNotError(t *testing.T) {
	repo := &testRepo{}
	orchestrator := NewOrchestrator(OrchestratorParams{
		Repo:     repo,
		Interval: 10 * time.Millisecond,
		Handlers: []Handler{},
	})

	err := orchestrator.ProcessJobs(context.Background())
	if err != nil && !errors.Is(err, common.ErrResourceNotFound) {
		t.Fatalf("expected nil error when no jobs are available, got: %v", err)
	}
}
