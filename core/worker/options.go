package worker

import "time"

const (
	NoInterval = time.Nanosecond
)

// Option configures a Worker
type Option func(*AbstractWorker)

// WithAbortOnError makes the AbstractWorker stop on any error
func WithAbortOnError() Option {
	return func(w *AbstractWorker) {
		w.abortOnError = true
	}
}

// WithSynchronousExecution runs the AbstractWorker in the calling goroutine
func WithSynchronousExecution() Option {
	return func(w *AbstractWorker) {
		w.synchronous = true
	}
}
