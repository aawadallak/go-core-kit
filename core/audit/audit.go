// Package audit provides audit functionality.
package audit

import (
	"context"
	"time"
)

type Log struct {
	Timestamp  time.Time `json:"timestamp"`
	TraceID    string    `json:"trace_id"`
	UserID     string    `json:"user_id,omitempty"`
	Method     string    `json:"method"`
	Endpoint   string    `json:"endpoint"`
	StatusCode int       `json:"status_code"`
	IP         string    `json:"ip"`
	Signature  string    `json:"signature,omitempty"`
}

type Provider interface {
	Close(ctx context.Context) error
	Flush(ctx context.Context, entries ...Log) error
}

type Service interface {
	Dispatch(ctx context.Context, log Log) error
	Flush(ctx context.Context) error
	Close(ctx context.Context) error
}
