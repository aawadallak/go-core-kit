package common

import (
	"context"
	"testing"
)

func TestWithRequestContext_RoundTrip(t *testing.T) {
	rc := &RequestContext{
		RequestID: "req-123",
		TraceID:   "trace-abc",
		SpanID:    "span-xyz",
	}

	ctx := WithRequestContext(context.Background(), rc)
	got := RequestContextFrom(ctx)

	if got == nil {
		t.Fatal("expected non-nil RequestContext")
	}

	if got.RequestID != rc.RequestID {
		t.Errorf("RequestID = %q, want %q", got.RequestID, rc.RequestID)
	}
	if got.TraceID != rc.TraceID {
		t.Errorf("TraceID = %q, want %q", got.TraceID, rc.TraceID)
	}
	if got.SpanID != rc.SpanID {
		t.Errorf("SpanID = %q, want %q", got.SpanID, rc.SpanID)
	}
}

func TestRequestContextFrom_EmptyContext(t *testing.T) {
	got := RequestContextFrom(context.Background())
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}
