// Package sealgorm provides sealgorm functionality.
package sealgorm

import (
	"context"

	"github.com/aawadallak/go-core-kit/core/repository"
	"github.com/aawadallak/go-core-kit/plugin/abstractrepo"
	"github.com/aawadallak/go-core-kit/plugin/seal"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
	repository.AbstractRepository[seal.SealedMessage]
}

var _ seal.SealedMessageRepository = (*Repository)(nil)

func (r *Repository) FindOne(ctx context.Context, query *seal.SealedMessage) (*seal.SealedMessage, error) {
	var out seal.SealedMessage

	tx, err := abstractrepo.FromContext(ctx)
	if err != nil {
		tx = r.db
	}

	tx = tx.Where("deleted_at IS NULL")

	if err := tx.
		Model(new(seal.SealedMessage)).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(query).
		Take(&out).Error; err != nil {
		return nil, err
	}

	return &out, nil
}

func NewRepository(db *gorm.DB) (seal.SealedMessageRepository, error) {
	ab, err := abstractrepo.NewAbstractRepository[seal.SealedMessage](db)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:                 db,
		AbstractRepository: ab,
	}, nil
}
