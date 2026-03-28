// Package audit provides audit functionality.
package audit

import (
	"context"
	"time"
)

type Log struct {
	Timestamp time.Time      `json:"timestamp"`
	TraceID   string         `json:"trace_id,omitempty"`
	UserID    string         `json:"user_id,omitempty"`
	Action    string         `json:"action"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// NewHTTPLog creates a Log pre-populated with HTTP-specific metadata.
func NewHTTPLog(traceID, userID, method, endpoint string, statusCode int, ip string) Log {
	return Log{
		Timestamp: time.Now(),
		TraceID:   traceID,
		UserID:    userID,
		Action:    method + " " + endpoint,
		Metadata: map[string]any{
			"method":      method,
			"endpoint":    endpoint,
			"status_code": statusCode,
			"ip":          ip,
		},
	}
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
