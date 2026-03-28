package outbox

import (
	"context"
	"time"

	"github.com/aawadallak/go-core-kit/plugin/abstractrepo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Insert(ctx context.Context, entry *Entry) error {
	tx, err := abstractrepo.FromContext(ctx)
	if err != nil {
		tx = r.db
	}

	if entry.CorrelationID != "" {
		return tx.WithContext(ctx).
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(entry).Error
	}

	return tx.WithContext(ctx).Create(entry).Error
}

func (r *Repository) FetchPending(ctx context.Context, batchSize int, eventNames []string) ([]Entry, error) {
	var entries []Entry
	q := r.db.WithContext(ctx).
		Where("status IN ?", []EntryStatus{EntryStatusPending, EntryStatusFailed}).
		Where("retry_count < ?", 10)

	if len(eventNames) > 0 {
		q = q.Where("event_name IN ?", eventNames)
	}

	err := q.
		Order("created_at ASC").
		Limit(batchSize).
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Find(&entries).Error
	return entries, err
}

func (r *Repository) MarkSent(ctx context.Context, id uint) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&Entry{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       EntryStatusSent,
			"processed_at": &now,
		}).Error
}

func (r *Repository) MarkFailed(ctx context.Context, id uint, errMsg string) error {
	return r.db.WithContext(ctx).
		Model(&Entry{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      EntryStatusFailed,
			"retry_count": gorm.Expr("retry_count + 1"),
			"last_error":  errMsg,
		}).Error
}

func (r *Repository) MarkExhausted(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).
		Model(&Entry{}).
		Where("id = ?", id).
		Update("status", EntryStatusExhausted).Error
}

func (r *Repository) CleanupSent(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := r.db.WithContext(ctx).
		Where("status = ?", EntryStatusSent).
		Where("processed_at < ?", cutoff).
		Delete(&Entry{})
	return result.RowsAffected, result.Error
}
