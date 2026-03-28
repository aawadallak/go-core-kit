package gorm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aawadallak/go-core-kit/core/idem"
	gormpkg "gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type model struct {
	IdempotencyKey string     `gorm:"column:idempotency_key;primaryKey"`
	Status         string     `gorm:"column:status;not null"`
	Outcome        []byte     `gorm:"column:outcome;type:jsonb"`
	Owner          string     `gorm:"column:owner;not null"`
	LeaseUntil     *time.Time `gorm:"column:lease_until"`
	CreatedAt      time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;not null"`
	CompletedAt    *time.Time `gorm:"column:completed_at"`
}

type Store struct {
	db    *gormpkg.DB
	table string
}

func NewStore(db *gormpkg.DB, table string) (*Store, error) {
	if db == nil {
		return nil, errors.New("gorm store requires a non-nil db")
	}
	if table == "" {
		table = "idem_records"
	}
	return &Store{db: db, table: table}, nil
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	query := fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s (
	idempotency_key TEXT PRIMARY KEY,
	status TEXT NOT NULL,
	outcome JSONB,
	owner TEXT NOT NULL,
	lease_until TIMESTAMPTZ NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	completed_at TIMESTAMPTZ NULL
);
`, s.table)
	return s.db.WithContext(ctx).Exec(query).Error
}

func (s *Store) Claim(ctx context.Context, key string, opts idem.ClaimOptions) (idem.ClaimResult, error) {
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return idem.ClaimResult{}, tx.Error
	}
	defer func() {
		_ = tx.Rollback().Error // best-effort rollback; benign error if tx already committed
	}()

	now := time.Now()
	lease := now.Add(opts.TTL)

	var current model
	res := tx.Table(s.table).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("idempotency_key = ?", key).
		Limit(1).
		Find(&current)

	if res.Error != nil {
		return idem.ClaimResult{}, res.Error
	}

	if res.RowsAffected == 0 {
		current = model{
			IdempotencyKey: key,
			Status:         string(idem.StatusProcessing),
			Owner:          opts.Owner,
			LeaseUntil:     &lease,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := tx.Table(s.table).Create(&current).Error; err != nil {
			return idem.ClaimResult{}, err
		}
		if err := tx.Commit().Error; err != nil {
			return idem.ClaimResult{}, err
		}
		return idem.ClaimResult{Acquired: true, Record: toRecord(&current)}, nil
	}

	switch idem.Status(current.Status) {
	case idem.StatusCompleted, idem.StatusFailed, idem.StatusDropped:
		if err := tx.Commit().Error; err != nil {
			return idem.ClaimResult{}, err
		}
		return idem.ClaimResult{Acquired: false, Record: toRecord(&current)}, nil
	case idem.StatusProcessing:
		if current.LeaseUntil != nil && current.LeaseUntil.After(now) {
			if err := tx.Commit().Error; err != nil {
				return idem.ClaimResult{}, err
			}
			return idem.ClaimResult{Acquired: false, Record: toRecord(&current)}, nil
		}
	}

	updates := map[string]any{
		"status":       string(idem.StatusProcessing),
		"owner":        opts.Owner,
		"lease_until":  lease,
		"updated_at":   now,
		"completed_at": nil,
	}
	if err := tx.Table(s.table).
		Where("idempotency_key = ?", key).
		Updates(updates).Error; err != nil {
		return idem.ClaimResult{}, err
	}

	if err := tx.Table(s.table).
		Where("idempotency_key = ?", key).
		Take(&current).Error; err != nil {
		return idem.ClaimResult{}, err
	}

	if err := tx.Commit().Error; err != nil {
		return idem.ClaimResult{}, err
	}

	return idem.ClaimResult{Acquired: true, Record: toRecord(&current)}, nil
}

func (s *Store) Complete(ctx context.Context, key string, outcome []byte) (idem.Record, error) {
	return s.updateTerminal(ctx, key, idem.StatusCompleted, outcome)
}

func (s *Store) Fail(ctx context.Context, key string, outcome []byte, status idem.Status) (idem.Record, error) {
	return s.updateTerminal(ctx, key, status, outcome)
}

func (s *Store) Get(ctx context.Context, key string) (idem.Record, bool, error) {
	var m model
	err := s.db.WithContext(ctx).Table(s.table).Where("idempotency_key = ?", key).Take(&m).Error
	if errors.Is(err, gormpkg.ErrRecordNotFound) {
		return idem.Record{}, false, nil
	}
	if err != nil {
		return idem.Record{}, false, err
	}
	return toRecord(&m), true, nil
}

func (s *Store) updateTerminal(ctx context.Context, key string, status idem.Status, outcome []byte) (idem.Record, error) {
	now := time.Now()
	updates := map[string]any{
		"status":       string(status),
		"outcome":      gormpkg.Expr("?::jsonb", string(outcome)),
		"lease_until":  nil,
		"updated_at":   now,
		"completed_at": now,
	}

	res := s.db.WithContext(ctx).Table(s.table).Where("idempotency_key = ?", key).Updates(updates)
	if res.Error != nil {
		return idem.Record{}, res.Error
	}
	if res.RowsAffected == 0 {
		return idem.Record{}, gormpkg.ErrRecordNotFound
	}

	var m model
	if err := s.db.WithContext(ctx).Table(s.table).Where("idempotency_key = ?", key).Take(&m).Error; err != nil {
		return idem.Record{}, err
	}
	return toRecord(&m), nil
}

func toRecord(m *model) idem.Record {
	return idem.Record{
		Key:         m.IdempotencyKey,
		Status:      idem.Status(m.Status),
		Outcome:     m.Outcome,
		Owner:       m.Owner,
		LeaseUntil:  m.LeaseUntil,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		CompletedAt: m.CompletedAt,
	}
}
