// Package jorm provides jorm functionality.
package jorm

import (
	"context"
	"errors"
	"time"

	"github.com/aawadallak/go-core-kit/common"

	j "github.com/aawadallak/go-core-kit/core/job"
	"github.com/aawadallak/go-core-kit/core/ptr"
	"github.com/aawadallak/go-core-kit/plugin/abstractrepo"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JobRepository struct {
	delegate *gorm.DB
}

var _ j.Repository = (*JobRepository)(nil)

func NewRepository(db *gorm.DB) (*JobRepository, error) {
	return &JobRepository{delegate: db}, nil
}

// AppendToQueue implements job.PipelineJobRepository.
func (p *JobRepository) AppendToQueue(ctx context.Context, jobs ...j.Job) error {
	return p.delegate.CreateInBatches(jobs, 100).Error
}

// GetNextJob implements job.PipelineJobRepository.
func (p *JobRepository) GetNextJob(ctx context.Context) (*j.Job, error) {
	tx, err := abstractrepo.FromContext(ctx)
	if err != nil {
		tx = p.delegate
	}

	var job j.Job

	if err := tx.
		WithContext(ctx).
		Model(&j.Job{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("status = ?", "pending").
		Where("trigger_at <= ? OR trigger_at IS NULL", ptr.Now()).
		Order("created_at ASC").
		First(&job).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.NewErrResourceNotFound(err)
		}

		return nil, err
	}

	return &job, nil
}

// UpdateJob implements job.PipelineJobRepository.
func (p *JobRepository) UpdateJob(ctx context.Context, job *j.Job) error {
	tx, err := abstractrepo.FromContext(ctx)
	if err != nil {
		tx = p.delegate
	}

	return tx.WithContext(ctx).
		Model(&j.Job{}).
		Where("id = ?", job.ID).
		UpdateColumns(map[string]any{
			"updated_at": time.Now(),
			"status":     job.Status,
			"error":      job.Error,
		}).Error
}

func (p *JobRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return p.delegate.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(abstractrepo.WithTx(ctx, tx))
	})
}
