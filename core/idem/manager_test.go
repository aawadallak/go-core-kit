package idem_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aawadallak/go-core-kit/core/idem"
	"github.com/aawadallak/go-core-kit/common"
	"github.com/aawadallak/go-core-kit/plugin/idem/inmem"
)

func newManager() *idem.Manager {
	return idem.NewManager(inmem.NewStore(), inmem.NewLocker(), idem.JSONCodec{})
}

func TestHandleExecutesOnceAndReusesCompletedOutcome(t *testing.T) {
	ctx := context.Background()
	m := newManager()

	var calls atomic.Int32
	fn := func(context.Context) (noJSONTagsOutcome, error) { //nolint:unparam // test helper
		calls.Add(1)
		return noJSONTagsOutcome{DonationID: 1498, GrossCents: 5056, StreamerID: 411}, nil
	}

	first, err := idem.Handle(ctx, m, idem.HandleRequest[noJSONTagsOutcome]{
		Key:   "donation:1498",
		Owner: "worker-a",
		TTL:   5 * time.Second,
		Run: func(ctx context.Context) (noJSONTagsOutcome, error) {
			return fn(ctx)
		},
	})
	if err != nil {
		t.Fatalf("first handle returned error: %v", err)
	}
	if !first.Executed || first.Reused {
		t.Fatalf("expected first execution to run (executed=%v reused=%v)", first.Executed, first.Reused)
	}
	if first.Status != idem.StatusCompleted {
		t.Fatalf("expected completed status, got %s", first.Status)
	}

	second, err := idem.Handle(ctx, m, idem.HandleRequest[noJSONTagsOutcome]{
		Key: "donation:1498",
		Run: func(ctx context.Context) (noJSONTagsOutcome, error) {
			return fn(ctx)
		},
	})
	if err != nil {
		t.Fatalf("second handle returned error: %v", err)
	}
	if second.Executed || !second.Reused {
		t.Fatalf("expected second execution to reuse terminal state (executed=%v reused=%v)", second.Executed, second.Reused)
	}
	if second.Status != idem.StatusCompleted {
		t.Fatalf("expected completed status on reuse, got %s", second.Status)
	}
	if second.Value == nil {
		t.Fatalf("expected reusable outcome, got nil")
	}
	if second.Value.GrossCents != 5056 {
		t.Fatalf("expected reusable outcome with amount=5056, got %#v", second.Value)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("expected handler to run once, got %d", got)
	}
}

type dropErr struct{}

func (dropErr) Error() string { return "drop this message" }
func (dropErr) FailureMode() common.FailureMode {
	return common.FailureModeDrop
}

func TestHandleMarksDroppedWhenFailureModeIsDrop(t *testing.T) {
	ctx := context.Background()
	m := newManager()

	_, err := idem.Handle(ctx, m, idem.HandleRequest[string]{
		Key: "donation:drop",
		Run: func(context.Context) (string, error) {
			return "", dropErr{}
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	rec, ok, getErr := m.Get(ctx, "donation:drop")
	if getErr != nil {
		t.Fatalf("get error: %v", getErr)
	}
	if !ok {
		t.Fatal("expected record")
	}
	if rec.Status != idem.StatusDropped {
		t.Fatalf("expected dropped status, got %s", rec.Status)
	}
}

func TestHandleMarksFailedForRegularErrors(t *testing.T) {
	ctx := context.Background()
	m := newManager()

	_, err := idem.Handle(ctx, m, idem.HandleRequest[string]{
		Key: "donation:failed",
		Run: func(context.Context) (string, error) {
			return "", errors.New("temporary downstream issue")
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}

	rec, ok, getErr := m.Get(ctx, "donation:failed")
	if getErr != nil {
		t.Fatalf("get error: %v", getErr)
	}
	if !ok {
		t.Fatal("expected record")
	}
	if rec.Status != idem.StatusFailed {
		t.Fatalf("expected failed status, got %s", rec.Status)
	}
}

type noJSONTagsOutcome struct {
	DonationID int
	GrossCents int64
	StreamerID int
}

func TestHandleWithMsgPackCodecRoundtrip(t *testing.T) {
	ctx := context.Background()
	m := idem.NewManager(inmem.NewStore(), inmem.NewLocker(), idem.MsgPackCodec{})

	fn := func(context.Context) (noJSONTagsOutcome, error) {
		return noJSONTagsOutcome{DonationID: 1498, GrossCents: 5147, StreamerID: 411}, nil
	}

	first, err := idem.Handle(ctx, m, idem.HandleRequest[noJSONTagsOutcome]{
		Key: "donation:msgpack",
		Run: func(ctx context.Context) (noJSONTagsOutcome, error) {
			return fn(ctx)
		},
	})
	if err != nil {
		t.Fatalf("first handle returned error: %v", err)
	}
	if first.Value == nil {
		t.Fatal("expected first value")
	}

	replay, err := idem.Handle(ctx, m, idem.HandleRequest[noJSONTagsOutcome]{
		Key: "donation:msgpack",
		Run: func(ctx context.Context) (noJSONTagsOutcome, error) {
			return fn(ctx)
		},
	})
	if err != nil {
		t.Fatalf("replay handle returned error: %v", err)
	}
	if replay.Value == nil {
		t.Fatal("expected replay value")
	}

	if replay.Value == nil {
		t.Fatal("expected replay value")
	}
	if replay.Value.DonationID != 1498 || replay.Value.GrossCents != 5147 || replay.Value.StreamerID != 411 {
		t.Fatalf("unexpected replay value: %+v", replay.Value)
	}
}
