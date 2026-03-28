package eventbroker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aawadallak/go-core-kit/core/event"
)

type Envelope struct {
	EventName     string          `json:"event_name"`
	Version       int             `json:"version"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload"`
	EventID       string          `json:"event_id,omitempty"`
	RequestID     string          `json:"request_id,omitempty"`
	TraceID       string          `json:"trace_id,omitempty"`
	SpanID        string          `json:"span_id,omitempty"`
}

type Dispatcher struct {
	transport Transport
	subject   string
}

var _ event.Dispatcher = (*Dispatcher)(nil)

func NewDispatcher(transport Transport, subject string) *Dispatcher {
	return &Dispatcher{
		transport: transport,
		subject:   subject,
	}
}

func (d *Dispatcher) Dispatch(ctx context.Context, evt *event.Record) error {
	env := Envelope{
		EventName:     evt.Name,
		Version:       evt.Version,
		CorrelationID: evt.CorrelationID,
		OccurredAt:    evt.Timestamp.UTC(),
		Payload:       evt.Metadata,
		EventID:       evt.ID,
		RequestID:     evt.RequestID,
		TraceID:       evt.TraceID,
		SpanID:        evt.SpanID,
	}

	data, err := json.Marshal(env)
	if err != nil {
		return err
	}

	return d.transport.Publish(ctx, d.subject, data)
}
