package abstractrepo

import (
	"context"
	"fmt"
	"math"

	"github.com/aawadallak/go-core-kit/core/repository"
	"gorm.io/gorm"
)

// AbstractPaginatedRepository provides a generic paginated repository implementation
// for querying entities using GORM. T represents the entity type, and E represents
// the query filter type.
type AbstractPaginatedRepository[T, E any] struct {
	db   *gorm.DB
	opts options
}

var _ repository.AbstractPaginatedRepository[any, any] = (*AbstractPaginatedRepository[any, any])(nil)

// Save creates a new entity in the database.
func (r *AbstractPaginatedRepository[T, E]) Save(ctx context.Context, target *T) (*T, error) {
	tx, err := FromContext(ctx)
	if err != nil {
		tx = r.db
	}
	if err := tx.WithContext(ctx).Create(target).Error; err != nil {
		return nil, fmt.Errorf("failed to create entity: %w", err)
	}
	return target, nil
}

// FindAll retrieves a paginated list of entities based on the provided query.
// It supports sorting, pagination, and returns a Pagination struct with results.
func (r *AbstractPaginatedRepository[T, E]) FindAll(
	ctx context.Context, entity *repository.PaginationQuery[E]) (*repository.Pagination[T], error) {
	tx, err := FromContext(ctx)
	if err != nil {
		tx = r.db // Fallback to default DB
	}

	instance := tx.WithContext(ctx).Model(new(T))

	// Get total count
	var count int64
	if err := instance.Count(&count).Error; err != nil {
		return nil, fmt.Errorf("failed to count entities: %w", err)
	}

	// Apply sorting
	if entity.Order.Field != "" && entity.Order.Direction != "" {
		instance = instance.Order(fmt.Sprintf("%s %s", entity.Order.Field, entity.Order.Direction))
	}

	// Calculate pagination
	const defaultPerPage = 10
	const defaultPage = 1
	perPage := entity.PerPage
	page := entity.Page
	if perPage <= 0 {
		perPage = defaultPerPage
	}
	if page <= 0 {
		page = defaultPage
	}

	totalPages := int(math.Ceil(float64(count) / float64(perPage)))
	offset := (page - 1) * perPage
	instance = instance.Limit(perPage).Offset(offset)

	// Fetch data
	var data []T
	if err := instance.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch entities: %w", err)
	}

	return &repository.Pagination[T]{
		Data:  data,
		Page:  page,
		Pages: totalPages,
		Total: int(count),
	}, nil
}
