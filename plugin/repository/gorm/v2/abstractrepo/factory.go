package abstractrepo

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// defaultEntityIndexes returns the default indexes for the core Entity fields
func defaultEntityIndexes() []indexConfig {
	return []indexConfig{
		{
			Name:   "idx_entity_external_id",
			Fields: []string{"external_id"},
			Type:   "unique",
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

// indexConfig represents the configuration for a single index
type indexConfig struct {
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
	// Migrate indicates whether to migrate the database
	Migrate bool
	// ApplyDefaultEntityIndexes indicates whether to apply default indexes for core Entity fields
	ApplyDefaultEntityIndexes bool
	// SoftDelete indicates whether to use soft delete
	SoftDelete bool
}

func WithMigrate() Option {
	return func(o *options) {
		o.Migrate = true
	}
}

// WithDefaultEntityIndexesOmitted configures the repository to skip applying default indexes for core Entity fields.
func WithDefaultEntityIndexesOmitted() Option {
	return func(o *options) {
		o.ApplyDefaultEntityIndexes = false
	}
}

// WithHardDelete configures the repository to use hard delete instead of soft delete.
func WithHardDelete() Option {
	return func(o *options) {
		o.SoftDelete = false
	}
}

// applyIndexes applies the specified indexes to the database if they don't already exist
func applyIndexes[T any](db *gorm.DB) error {
	indexes := defaultEntityIndexes()

	// Get the table name for type T
	var model T
	tableName := db.NamingStrategy.TableName(reflect.TypeOf(model).Name())

	for _, idx := range indexes {
		fields := strings.Join(idx.Fields, ", ")
		indexType := "INDEX"
		if idx.Type == "unique" {
			indexType = "UNIQUE INDEX"
		}

		// Construct the raw SQL query for index creation
		query := fmt.Sprintf("CREATE %s %s ON %s (%s)",
			indexType,
			idx.Name,
			tableName,
			fields,
		)

		// Attempt to create the index
		if err := db.Exec(query).Error; err != nil {
			// Check if the error is because the index already exists
			if isIndexExistsError(err) {
				continue
			}
			return fmt.Errorf("failed to create index %s: %w", idx.Name, err)
		}
	}

	return nil
}

// isIndexExistsError checks if the error indicates the index already exists
func isIndexExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	// Common error messages for "index already exists" across databases
	return strings.Contains(errStr, "already exists") ||
		strings.Contains(errStr, "duplicate key") ||
		strings.Contains(errStr, "duplicate index") ||
		strings.Contains(errStr, "exists") ||
		errors.Is(err, gorm.ErrDuplicatedKey)
}

// setupUUIDCallback sets up GORM hooks to automatically set UUID for external IDs
func setupUUIDCallback[T any](db *gorm.DB) {
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

	return instance, nil
}

func NewAbstractRepository[T any](
	db *gorm.DB, opts ...Option) (*AbstractRepository[T], error) {
	options := &options{
		Migrate:                   false,
		SoftDelete:                true,
		ApplyDefaultEntityIndexes: true,
	}

	for _, opt := range opts {
		opt(options)
	}

	if val, err := strconv.ParseBool(os.Getenv("DB_MIGRATE")); err == nil {
		options.Migrate = val
	}

	instance := &AbstractRepository[T]{
		db:   db,
		opts: *options,
	}

	if options.Migrate {
		if err := db.AutoMigrate(new(T)); err != nil {
			return nil, fmt.Errorf("failed to migrate: %w", err)
		}
	}

	if options.ApplyDefaultEntityIndexes {
		if err := applyIndexes[T](db); err != nil {
			return nil, fmt.Errorf("failed to apply indexes: %w", err)
		}
	}

	setupUUIDCallback[T](db)

	return instance, nil
}

func notDeletedScope(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}
