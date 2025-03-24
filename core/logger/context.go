package logger

import "context"

// loggerContextKey is a private struct used as a unique key for context values
type loggerContextKey struct{}

// loggerContext is the key used to store/retrieve logger instances from context
var loggerContext = &loggerContextKey{}

// Of retrieves a Logger from the context, falling back to Global if none found
func Of(ctx context.Context) Logger {
	if l := fromContext(ctx); l != nil {
		return l
	}
	return Global()
}

// fromContext attempts to extract a Logger from the context
func fromContext(ctx context.Context) Logger {
	if l, ok := ctx.Value(loggerContext).(Logger); ok {
		return l
	}
	return nil
}

// WithContext appends a single option to the existing logger in the context
// If no logger exists, creates a new one with the provided option
func WithContext(ctx context.Context, opt Option) context.Context {
	current := fromContext(ctx)
	if current == nil {
		return context.WithValue(ctx, loggerContext, New(opt))
	}
	newLogger := current.With(opt)
	return context.WithValue(ctx, loggerContext, newLogger)
}

// WithLogger creates a new logger instance based on an existing one
// If parent is nil, creates a new logger with the provided options
func WithLogger(parent Logger, opts ...Option) Logger {
	if parent == nil {
		return New(opts...)
	}
	return parent.With(opts...)
}
