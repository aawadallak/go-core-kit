package abstractrepo

import (
	"gorm.io/gorm"
)

// Option is a function type that modifies the options struct
type Option func(*options)

// options defines configuration options for repository creation
type options struct {
	// Migrate indicates whether to migrate the database
	Migrate bool
	// SoftDelete indicates whether to use soft delete
	SoftDelete bool
	NoReturn   bool

	Preloads []string
}

func WithMigrate() Option {
	return func(o *options) {
		o.Migrate = true
	}
}

// WithHardDelete configures the repository to use hard delete instead of soft delete.
func WithHardDelete() Option {
	return func(o *options) {
		o.SoftDelete = false
	}
}

func WithPreloads(preloads ...string) Option {
	return func(o *options) {
		o.Preloads = preloads
	}
}

func WithNoReturn() Option {
	return func(o *options) {
		o.NoReturn = true
	}
}

func NewAbstractPaginatedRepository[T, E any](
	db *gorm.DB, opts ...Option) (*AbstractPaginatedRepository[T, E], error) {
	options := &options{
		SoftDelete: true,
	}

	for _, opt := range opts {
		opt(options)
	}

	instance := &AbstractPaginatedRepository[T, E]{
		db:   db,
		opts: *options,
	}

	return instance, nil
}

func NewAbstractRepository[T any](
	db *gorm.DB, opts ...Option) (*AbstractRepository[T], error) {
	options := &options{
		SoftDelete: true,
	}

	for _, opt := range opts {
		opt(options)
	}

	instance := &AbstractRepository[T]{
		db:   db,
		opts: *options,
	}

	// AutoMigrate removed - all migrations are now handled via SQL migrations

	return instance, nil
}

func notDeletedScope(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}
