package audit

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestNewHTTPLog(t *testing.T) {
	log := NewHTTPLog("trace-1", "user-42", "POST", "/api/items", 201, "10.0.0.1")

	if log.TraceID != "trace-1" {
		t.Errorf("expected TraceID %q, got %q", "trace-1", log.TraceID)
	}
	if log.UserID != "user-42" {
		t.Errorf("expected UserID %q, got %q", "user-42", log.UserID)
	}
	if log.Action != "POST /api/items" {
		t.Errorf("expected Action %q, got %q", "POST /api/items", log.Action)
	}
	if log.Timestamp.IsZero() {
		t.Error("expected non-zero Timestamp")
	}

	assertMeta := func(key string, want any) {
		t.Helper()
		got, ok := log.Metadata[key]
		if !ok {
			t.Errorf("expected Metadata key %q to exist", key)
			return
		}
		if got != want {
			t.Errorf("Metadata[%q] = %v, want %v", key, got, want)
		}
	}

	assertMeta("method", "POST")
	assertMeta("endpoint", "/api/items")
	assertMeta("status_code", 201)
	assertMeta("ip", "10.0.0.1")
}

// spyProvider records flushed logs for test verification.
type spyProvider struct {
	mu      sync.Mutex
	flushed []Log
}

func (s *spyProvider) Close(_ context.Context) error { return nil }

func (s *spyProvider) Flush(_ context.Context, entries ...Log) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flushed = append(s.flushed, entries...)
	return nil
}

func (s *spyProvider) logs() []Log {
	s.mu.Lock()
	defer s.mu.Unlock()
	dst := make([]Log, len(s.flushed))
	copy(dst, s.flushed)
	return dst
}

func TestOrchestratorDispatchAndFlush_BatchSize(t *testing.T) {
	spy := &spyProvider{}
	ctx := t.Context()

	orch := NewOrchestrator(ctx,
		WithBatchSize(2),
		WithBatchInterval(10*time.Minute), // long interval so only batch size triggers
		WithProvider(spy),
	)

	log1 := Log{Action: "action-1", Timestamp: time.Now()}
	log2 := Log{Action: "action-2", Timestamp: time.Now()}

	orch.Stream <- log1
	orch.Stream <- log2

	// Wait briefly for the orchestrator goroutine to flush.
	time.Sleep(200 * time.Millisecond)

	got := spy.logs()
	if len(got) != 2 {
		t.Fatalf("expected 2 flushed logs, got %d", len(got))
	}
	if got[0].Action != "action-1" {
		t.Errorf("expected first log Action %q, got %q", "action-1", got[0].Action)
	}
	if got[1].Action != "action-2" {
		t.Errorf("expected second log Action %q, got %q", "action-2", got[1].Action)
	}
}

func TestOrchestratorDispatchAndFlush_Interval(t *testing.T) {
	spy := &spyProvider{}
	ctx := t.Context()

	orch := NewOrchestrator(ctx,
		WithBatchSize(100), // large batch so interval triggers first
		WithBatchInterval(100*time.Millisecond),
		WithProvider(spy),
	)

	log1 := Log{Action: "interval-action", Timestamp: time.Now()}
	orch.Stream <- log1

	// Wait for the interval tick to trigger a flush.
	time.Sleep(300 * time.Millisecond)

	got := spy.logs()
	if len(got) != 1 {
		t.Fatalf("expected 1 flushed log via interval, got %d", len(got))
	}
	if got[0].Action != "interval-action" {
		t.Errorf("expected Action %q, got %q", "interval-action", got[0].Action)
	}
}
