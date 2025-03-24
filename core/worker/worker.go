package worker

import (
	"context"
	"time"
)

// Worker defines the interface for a worker that can be started and closed.
type Worker interface {
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}

// TaskProvider defines the required interface for worker tasks
type TaskProvider interface {
	Interval() time.Duration
	Execute(context.Context) error
}

// Optional lifecycle hooks
type OnStarter interface {
	OnStart(context.Context) error
}

type OnPreExecutor interface {
	OnPreExecute(context.Context) error
}

type OnPostExecutor interface {
	OnPostExecute(context.Context) error
}

type OnShutdowner interface {
	OnShutdown(context.Context) error
}
