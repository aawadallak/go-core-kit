package eventbroker

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aawadallak/go-core-kit/core/event"
	brokerjs "github.com/aawadallak/go-core-kit/plugin/broker/natsjetstream"
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

type DispatcherConfig struct {
	Endpoint   string
	StreamName string
	Subject    string
}

type Dispatcher struct {
	publisher *brokerjs.Publisher
}

var _ event.Dispatcher = (*Dispatcher)(nil)

func NewDispatcher(ctx context.Context, cfg DispatcherConfig) (*Dispatcher, error) {
	publisher, err := brokerjs.NewPublisher(ctx, brokerjs.PublisherConfig{
		Endpoint:   cfg.Endpoint,
		StreamName: cfg.StreamName,
		Subject:    cfg.Subject,
	})
	if err != nil {
		return nil, err
	}

	return &Dispatcher{publisher: publisher}, nil
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

	return d.publisher.Publish(ctx, brokerjs.PublishMessage{Message: env})
}

func (d *Dispatcher) Close() error {
	if d.publisher == nil {
		return nil
	}
	return d.publisher.Close()
}
