package idem

import (
	"context"
	"time"
)

type Status string

const (
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusDropped    Status = "dropped"
)

type Record struct {
	Key         string
	Status      Status
	Outcome     []byte
	Owner       string
	LeaseUntil  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CompletedAt *time.Time
}

type ClaimOptions struct {
	Owner string
	TTL   time.Duration
}

type ClaimResult struct {
	Acquired bool
	Record   Record
}

type HandleRequest[T any] struct {
	Key   string
	Owner string
	TTL   time.Duration
	Run   func(context.Context) (T, error)
}

type HandleResult[T any] struct {
	Key      string
	Status   Status
	Executed bool
	Reused   bool
	Value    *T
}
