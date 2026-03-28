// Package common provides common functionality.
package common

import "context"

// RequestContext holds transport-agnostic observability fields.
type RequestContext struct {
	RequestID string `json:"request_id"`
	TraceID   string `json:"trace_id"`
	SpanID    string `json:"span_id"`
}

type requestContextKey struct{}

// WithRequestContext stores a RequestContext in the context.
func WithRequestContext(ctx context.Context, rc *RequestContext) context.Context {
	return context.WithValue(ctx, requestContextKey{}, rc)
}

// RequestContextFrom retrieves the RequestContext from the context.
// Returns nil if no RequestContext is present.
func RequestContextFrom(ctx context.Context) *RequestContext {
	rc, ok := ctx.Value(requestContextKey{}).(*RequestContext)
	if !ok {
		return nil
	}

	return rc
}
