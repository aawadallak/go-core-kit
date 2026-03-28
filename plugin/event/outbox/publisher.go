package outbox

import (
	"context"
	"encoding/json"
	"time"

	cevent "github.com/aawadallak/go-core-kit/core/event"
	"github.com/aawadallak/go-core-kit/common"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type inserter interface {
	Insert(ctx context.Context, entry *Entry) error
}

type Publisher struct {
	repo inserter
}

var _ cevent.Publisher = (*Publisher)(nil)

func NewPublisher(repo inserter) *Publisher {
	return &Publisher{repo: repo}
}

func (p *Publisher) Publish(ctx context.Context, metadata cevent.Metadata) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	entry := &Entry{
		EventID:      uuid.New().String(),
		EventName:    metadata.EventType(),
		EventVersion: metadata.EventVersion(),
		Payload:      payload,
		Status:       EntryStatusPending,
		CreatedAt:    time.Now(),
	}

	if rc := common.RequestContextFrom(ctx); rc != nil {
		entry.RequestID = rc.RequestID
		entry.TraceID = rc.TraceID
		entry.SpanID = rc.SpanID
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		entry.TraceID = span.SpanContext().TraceID().String()
		entry.SpanID = span.SpanContext().SpanID().String()
	}

	return p.repo.Insert(ctx, entry)
}
