// Package job provides job functionality.
package job

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aawadallak/go-core-kit/common"
)

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type Job struct {
	common.Entity
	CorrelationID string
	Status        JobStatus
	Type          string
	Error         string
	Metadata      json.RawMessage
	TriggerAt     *time.Time
}

type Repository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	AppendToQueue(ctx context.Context, jobs ...Job) error
	GetNextJob(ctx context.Context) (*Job, error)
	UpdateJob(ctx context.Context, job *Job) error
}

type Handler interface {
	Process(ctx context.Context, job *Job) error
	GetType() string
}

type JobParams struct {
	JobType       string
	CorrelationID string
	Metadata      json.RawMessage
	TriggerAt     *time.Time
}

func NewJob(params JobParams) Job {
	return Job{
		Type:          params.JobType,
		CorrelationID: params.CorrelationID,
		Metadata:      params.Metadata,
		Status:        JobStatusPending,
		TriggerAt:     params.TriggerAt,
	}
}
