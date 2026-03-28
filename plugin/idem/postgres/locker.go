// Package postgres provides postgres functionality.
package postgres

import (
	"context"
	"database/sql"
	"errors"
)

type Locker struct {
	db *sql.DB
}

func NewLocker(db *sql.DB) (*Locker, error) {
	if db == nil {
		return nil, errors.New("postgres locker requires a non-nil db")
	}

	return &Locker{db: db}, nil
}

func (l *Locker) TryLock(ctx context.Context, key string) (acquired bool, release func(context.Context) error, err error) {
	conn, err := l.db.Conn(ctx)
	if err != nil {
		return false, nil, err
	}

	var locked bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(hashtext($1))", key).Scan(&locked); err != nil {
		_ = conn.Close() // best-effort cleanup; query error takes precedence
		return false, nil, err
	}

	if !locked {
		_ = conn.Close() // best-effort cleanup; lock not acquired
		return false, nil, nil
	}

	unlock := func(unlockCtx context.Context) error {
		var unlocked bool
		err := conn.QueryRowContext(unlockCtx, "SELECT pg_advisory_unlock(hashtext($1))", key).Scan(&unlocked)
		closeErr := conn.Close()
		if err != nil {
			return err
		}
		return closeErr
	}

	return true, unlock, nil
}
