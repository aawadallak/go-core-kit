// Package event provides event functionality.
package event

import (
	"context"
	"encoding/json"
	"time" // For timestamps

	"github.com/google/uuid" // For generating unique IDs
)

// Metadata is an interface that all specific event metadata structs should implement.
// This allows for type-safe handling of event specific data.
type Metadata interface {
	EventType() string // Returns the unique name for the event type (e.g., "USER_SIGNED_UP")
	EventVersion() int // Returns the version of this specific event's schema (e.g., 1)
}

// Record represents a standardized event in your system.
type Record struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Timestamp      time.Time       `json:"timestamp"`
	CorrelationID  string          `json:"correlation_id,omitempty"`
	OrganizationID string          `json:"organization_id,omitempty"`
	RequestID      string          `json:"request_id,omitempty"`
	TraceID        string          `json:"trace_id,omitempty"`
	SpanID         string          `json:"span_id,omitempty"`
	Version        int             `json:"version"`
	Metadata       json.RawMessage `json:"metadata"`
}

// Option is a function that can modify an EventRecord.
type Option func(*Record)

// WithCorrelationID sets the correlation ID for the event.
func WithCorrelationID(id string) Option {
	return func(r *Record) {
		r.CorrelationID = id
	}
}

func WithOrganizationID(id string) Option {
	return func(r *Record) {
		r.OrganizationID = id
	}
}

func NewRecord(metadata Metadata, opts ...Option) (*Record, error) {
	bs, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}

	r := &Record{
		ID:        uuid.New().String(),
		Name:      metadata.EventType(),
		Timestamp: time.Now().UTC(),
		Version:   metadata.EventVersion(),
		Metadata:  bs,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func ExtractMetadata[T Metadata](r *Record) (T, error) {
	var metadata T
	if err := json.Unmarshal(r.Metadata, &metadata); err != nil {
		return metadata, err
	}
	return metadata, nil
}

type Dispatcher interface {
	Dispatch(ctx context.Context, event *Record) error
}

type Publisher interface {
	Publish(ctx context.Context, metadata Metadata) error
}
