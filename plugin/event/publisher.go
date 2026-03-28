// Package event provides event functionality.
package event

import (
	"context"

	"github.com/aawadallak/go-core-kit/common"
	cevent "github.com/aawadallak/go-core-kit/core/event"
	"go.opentelemetry.io/otel/trace"
)

type Publisher struct {
	dispatcher cevent.Dispatcher
}

var _ cevent.Publisher = (*Publisher)(nil)

func (ep *Publisher) Publish(ctx context.Context, metadata cevent.Metadata) error {
	record, err := cevent.NewRecord(metadata)
	if err != nil {
		return err
	}

	if rc := common.RequestContextFrom(ctx); rc != nil {
		record.RequestID = rc.RequestID
		record.TraceID = rc.TraceID
		record.SpanID = rc.SpanID
	}

	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		record.TraceID = span.SpanContext().TraceID().String()
		record.SpanID = span.SpanContext().SpanID().String()
	}

	return ep.dispatcher.Dispatch(ctx, record)
}

func NewPublisher(dispatcher cevent.Dispatcher) *Publisher {
	return &Publisher{dispatcher: dispatcher}
}
