package abstractrepo

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// AbstractRepository provides a generic CRUD repository for entities that implement Identifiable.
type AbstractRepository[T any] struct {
	db   *gorm.DB
	opts options
}

// Save creates a new entity in the database.
func (r *AbstractRepository[T]) Save(ctx context.Context, target *T) (*T, error) {
	tx, err := FromContext(ctx)
	if err != nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Create(target).Error; err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}
	return target, nil
}

// Update modifies an existing entity by its ID.
func (r *AbstractRepository[T]) Update(ctx context.Context, target *T) (*T, error) {
	v, ok := any(target).(interface{ GetID() uint })
	if !ok {
		return nil, ErrInvalidType
	}
	tx, err := FromContext(ctx)
	if err != nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).
		Where("id = ?", v.GetID()).
		Updates(target).Error; err != nil {
		return nil, fmt.Errorf("failed to update entity with ID %d: %w", v.GetID(), err)
	}
	return target, nil
}

// Delete removes an entity by its ID.
func (r *AbstractRepository[T]) Delete(ctx context.Context, target *T) error {
	v, ok := any(target).(interface{ GetID() uint })
	if !ok {
		return ErrInvalidType
	}
	tx, err := FromContext(ctx)
	if err != nil {
		tx = r.db
	}
	query := tx.WithContext(ctx).
		Model(new(T)).
		Where("id = ?", v.GetID())
	if r.opts.SoftDelete {
		err = query.Update("deleted_at", time.Now()).Error
	} else {
		err = query.Delete(new(T)).Error
	}
	return err
}

// FindOne retrieves a single entity matching the provided filter.
func (r *AbstractRepository[T]) FindOne(ctx context.Context, filter any) (*T, error) {
	var out T
	tx, err := FromContext(ctx)
	if err != nil {
		tx = r.db
	}
	if r.opts.SoftDelete {
		tx = tx.Scopes(notDeletedScope)
	}
	if err := tx.
		WithContext(ctx).
		Where(filter).
		Take(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}

// Find retrieves all entities matching the provided filter.
func (r *AbstractRepository[T]) Find(ctx context.Context, filter any) ([]*T, error) {
	var out []*T
	tx, err := FromContext(ctx)
	if err != nil {
		tx = r.db
	}
	if r.opts.SoftDelete {
		tx = tx.Scopes(notDeletedScope)
	}
	if err := tx.WithContext(ctx).Where(filter).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Tx executes a function within a database transaction.
func (r *AbstractRepository[T]) Tx(ctx context.Context, fn func(context.Context) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(WithTx(ctx, tx))
	})
}
