package abstractrepo

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DefaultEntityIndexes returns the default indexes for the core Entity fields
func DefaultEntityIndexes() []IndexConfig {
	return []IndexConfig{
		{
			Name:   "idx_entity_external_id",
			Fields: []string{"external_id"},
			Type:   "index",
		},
		{
			Name:   "idx_entity_created_at",
			Fields: []string{"created_at"},
			Type:   "index",
		},
		{
			Name:   "idx_entity_updated_at",
			Fields: []string{"updated_at"},
			Type:   "index",
		},
		{
			Name:   "idx_entity_deleted_at",
			Fields: []string{"deleted_at"},
			Type:   "index",
		},
	}
}

// IndexConfig represents the configuration for a single index
type IndexConfig struct {
	// Name is the name of the index
	Name string
	// Fields are the fields that make up the index
	Fields []string
	// Type is the type of index (e.g., "index", "unique")
	Type string
}

// Option is a function type that modifies the options struct
type Option func(*options)

// options defines configuration options for repository creation
type options struct {
	// UseUUID indicates whether to use UUID for external IDs
	UseUUID bool
	// ApplyDefaultEntityIndexes indicates whether to apply default indexes for core Entity fields
	ApplyDefaultEntityIndexes bool
}

// WithDefaultEntityIndexes returns an Option that will apply default indexes for core Entity fields
func WithDefaultEntityIndexes() Option {
	return func(o *options) {
		o.ApplyDefaultEntityIndexes = true
	}
}

// WithUUID returns an Option that enables UUID
func WithUUID() Option {
	return func(o *options) {
		o.UseUUID = true
	}
}

// applyIndexes applies the specified indexes to the database
func applyIndexes[T any](db *gorm.DB, opts options) error {
	// Initialize indexes as an empty slice to avoid nil panic when appending
	indexes := make([]IndexConfig, 0)

	// Add default entity indexes if requested
	if opts.ApplyDefaultEntityIndexes {
		indexes = append(indexes, DefaultEntityIndexes()...)
	}

	// Create the indexes
	for _, idx := range indexes {
		// Create the index using GORM's migrator
		if err := db.Migrator().CreateIndex(new(T), idx.Name); err != nil {
			return fmt.Errorf("failed to create index %s: %w", idx.Name, err)
		}
	}

	return nil
}

// setupUUIDHooks sets up GORM hooks to automatically set UUID for external IDs
func setupUUIDHooks[T any](db *gorm.DB) {
	db.Callback().Create().Before("gorm:create").Register("set_uuid", func(db *gorm.DB) {
		dest := db.Statement.Dest
		if dest == nil {
			return
		}

		// Get the value of the destination
		value := reflect.ValueOf(dest)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}

		// Look for ExternalID field
		if field := value.FieldByName("ExternalID"); field.IsValid() && field.CanSet() {
			field.SetString(uuid.New().String())
		}
	})
}

func NewAbstractPaginatedRepository[T, E any](
	db *gorm.DB, opts ...Option) (*AbstractPaginatedRepository[T, E], error) {
	options := &options{}

	for _, opt := range opts {
		opt(options)
	}

	instance := &AbstractPaginatedRepository[T, E]{
		db:   db,
		opts: *options,
	}

	if err := applyIndexes[T](db, *options); err != nil {
		return nil, fmt.Errorf("failed to apply indexes: %w", err)
	}

	if options.UseUUID {
		setupUUIDHooks[T](db)
	}

	return instance, nil
}

func NewAbstractRepository[T any](
	db *gorm.DB, opts ...Option) (*AbstractRepository[T], error) {
	options := &options{}

	for _, opt := range opts {
		opt(options)
	}

	instance := &AbstractRepository[T]{
		db:   db,
		opts: *options,
	}

	if err := applyIndexes[T](db, *options); err != nil {
		return nil, fmt.Errorf("failed to apply indexes: %w", err)
	}

	if options.UseUUID {
		setupUUIDHooks[T](db)
	}

	return instance, nil
}
