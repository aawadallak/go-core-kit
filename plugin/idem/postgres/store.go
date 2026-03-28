package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aawadallak/go-core-kit/core/idem"
)

type Store struct {
	db    *sql.DB
	table string
}

func NewStore(db *sql.DB, table string) (*Store, error) {
	if db == nil {
		return nil, errors.New("postgres store requires a non-nil db")
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

	_, err := s.db.ExecContext(ctx, query)
	return err
}

func (s *Store) Claim(ctx context.Context, key string, opts idem.ClaimOptions) (idem.ClaimResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return idem.ClaimResult{}, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback idem claim tx: %v", err)
		}
	}()

	rec, found, err := s.getForUpdate(ctx, tx, key)
	if err != nil {
		return idem.ClaimResult{}, err
	}

	now := time.Now()
	lease := now.Add(opts.TTL)

	if !found {
		//nolint:gosec // G201: table name is from trusted config, not user input
		insertQuery := fmt.Sprintf(`
INSERT INTO %s (idempotency_key, status, owner, lease_until, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING idempotency_key, status, outcome, owner, lease_until, created_at, updated_at, completed_at
`, s.table)

		if err := tx.QueryRowContext(ctx, insertQuery, key, idem.StatusProcessing, opts.Owner, lease).Scan(
			&rec.Key,
			&rec.Status,
			&rec.Outcome,
			&rec.Owner,
			&rec.LeaseUntil,
			&rec.CreatedAt,
			&rec.UpdatedAt,
			&rec.CompletedAt,
		); err != nil {
			return idem.ClaimResult{}, err
		}

		if err := tx.Commit(); err != nil {
			return idem.ClaimResult{}, err
		}
		return idem.ClaimResult{Acquired: true, Record: rec}, nil
	}

	switch rec.Status {
	case idem.StatusCompleted, idem.StatusFailed, idem.StatusDropped:
		if err := tx.Commit(); err != nil {
			return idem.ClaimResult{}, err
		}
		return idem.ClaimResult{Acquired: false, Record: rec}, nil
	case idem.StatusProcessing:
		if rec.LeaseUntil != nil && rec.LeaseUntil.After(now) {
			if err := tx.Commit(); err != nil {
				return idem.ClaimResult{}, err
			}
			return idem.ClaimResult{Acquired: false, Record: rec}, nil
		}
	}

	//nolint:gosec // G201: table name is from trusted config, not user input
	updateQuery := fmt.Sprintf(`
UPDATE %s
SET status = $2, owner = $3, lease_until = $4, updated_at = NOW(), completed_at = NULL
WHERE idempotency_key = $1
RETURNING idempotency_key, status, outcome, owner, lease_until, created_at, updated_at, completed_at
`, s.table)

	if err := tx.QueryRowContext(ctx, updateQuery, key, idem.StatusProcessing, opts.Owner, lease).Scan(
		&rec.Key,
		&rec.Status,
		&rec.Outcome,
		&rec.Owner,
		&rec.LeaseUntil,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&rec.CompletedAt,
	); err != nil {
		return idem.ClaimResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return idem.ClaimResult{}, err
	}

	return idem.ClaimResult{Acquired: true, Record: rec}, nil
}

func (s *Store) Complete(ctx context.Context, key string, outcome []byte) (idem.Record, error) {
	query := fmt.Sprintf(`
UPDATE %s
SET status = $2, outcome = $3::jsonb, updated_at = NOW(), completed_at = NOW(), lease_until = NULL
WHERE idempotency_key = $1
RETURNING idempotency_key, status, outcome, owner, lease_until, created_at, updated_at, completed_at
`, s.table)

	return s.scanOne(ctx, query, key, idem.StatusCompleted, string(outcome))
}

func (s *Store) Fail(ctx context.Context, key string, outcome []byte, status idem.Status) (idem.Record, error) {
	query := fmt.Sprintf(`
UPDATE %s
SET status = $2, outcome = $3::jsonb, updated_at = NOW(), completed_at = NOW(), lease_until = NULL
WHERE idempotency_key = $1
RETURNING idempotency_key, status, outcome, owner, lease_until, created_at, updated_at, completed_at
`, s.table)

	return s.scanOne(ctx, query, key, status, string(outcome))
}

func (s *Store) Get(ctx context.Context, key string) (idem.Record, bool, error) {
	//nolint:gosec // G201: table name is from trusted config, not user input
	query := fmt.Sprintf(`
SELECT idempotency_key, status, outcome, owner, lease_until, created_at, updated_at, completed_at
FROM %s
WHERE idempotency_key = $1
`, s.table)

	row := s.db.QueryRowContext(ctx, query, key)
	rec, err := scanRecord(row)
	if errors.Is(err, sql.ErrNoRows) {
		return idem.Record{}, false, nil
	}
	if err != nil {
		return idem.Record{}, false, err
	}

	return rec, true, nil
}

func (s *Store) getForUpdate(ctx context.Context, tx *sql.Tx, key string) (idem.Record, bool, error) {
	//nolint:gosec // G201: table name is from trusted config, not user input
	query := fmt.Sprintf(`
SELECT idempotency_key, status, outcome, owner, lease_until, created_at, updated_at, completed_at
FROM %s
WHERE idempotency_key = $1
FOR UPDATE
`, s.table)

	row := tx.QueryRowContext(ctx, query, key)
	rec, err := scanRecord(row)
	if errors.Is(err, sql.ErrNoRows) {
		return idem.Record{}, false, nil
	}
	if err != nil {
		return idem.Record{}, false, err
	}

	return rec, true, nil
}

func (s *Store) scanOne(ctx context.Context, query string, args ...any) (idem.Record, error) {
	row := s.db.QueryRowContext(ctx, query, args...)
	return scanRecord(row)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanRecord(row scanner) (idem.Record, error) {
	var rec idem.Record
	if err := row.Scan(
		&rec.Key,
		&rec.Status,
		&rec.Outcome,
		&rec.Owner,
		&rec.LeaseUntil,
		&rec.CreatedAt,
		&rec.UpdatedAt,
		&rec.CompletedAt,
	); err != nil {
		return idem.Record{}, err
	}
	return rec, nil
}
