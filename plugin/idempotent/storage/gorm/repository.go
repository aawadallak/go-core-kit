package idemgorm

import (
	"context"
	"errors"

	"github.com/aawadallak/go-core-kit/core/idempotent"
	"github.com/aawadallak/go-core-kit/plugin/repository/gorm/v2/abstractrepo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrDatabaseConnectionRequired = errors.New("database connection is required")
	ErrKeyRequired                = errors.New("key is required")
	ErrPayloadRequired            = errors.New("payload is required")
)

const (
	DefaultTableName = "idempotent_events"
)

// Entity represents the database model for idempotent events
type Entity struct {
	gorm.Model
	Key     string `gorm:"column:idem_key;uniqueIndex;not null;size:255"`
	Payload []byte `gorm:"not null"`
}

// Repository implements the idempotent.Repository interface using GORM
type Repository struct {
	db        *gorm.DB
	tableName string
}

// Option defines the function signature for repository options
type Option func(*Repository)

// WithTableName sets a custom table name for the repository
func WithTableName(name string) Option {
	return func(r *Repository) {
		r.tableName = name
	}
}

// TableName returns the name of the table for the Entity
func (e *Entity) TableName() string {
	return "" // This will be overridden by GORM's tabler interface
}

// NewRepository creates a new Repository instance with the given GORM DB connection and options
func NewRepository(db *gorm.DB, opts ...Option) (*Repository, error) {
	if db == nil {
		return nil, ErrDatabaseConnectionRequired
	}

	repo := &Repository{
		db:        db,
		tableName: DefaultTableName, // Set default table name
	}

	// Apply all provided options
	for _, opt := range opts {
		opt(repo)
	}

	if err := db.
		Table(repo.tableName).
		AutoMigrate(&Entity{}); err != nil {
		return nil, err
	}

	return repo, nil
}

// Save stores an idempotent event in the database
func (i *Repository) Save(ctx context.Context, item idempotent.EventItem) error {
	if item.Key == "" {
		return ErrKeyRequired
	}
	if item.Payload == nil {
		return ErrPayloadRequired
	}
	data := &Entity{
		Key:     item.Key,
		Payload: item.Payload,
	}
	if err := i.db.
		Table(i.tableName).
		WithContext(ctx).
		Save(data).Error; err != nil {
		return err
	}
	return nil
}

// Find retrieves an idempotent event from the database by its key
func (i *Repository) Find(ctx context.Context, key string) (*idempotent.EventItem, error) {
	if key == "" {
		return nil, ErrKeyRequired
	}

	tx, err := abstractrepo.FromContext(ctx)
	if err != nil {
		tx = i.db
	}

	var out Entity

	if err := tx.
		Table(i.tableName).
		WithContext(ctx).
		Where("idem_key = ?", key). // GORM handles placeholder differences
		// NOTE: Row-level locking (FOR UPDATE) only works when a transaction (`tx`) is provided.
		// If `tx` is nil or not a real transaction, the row will not be locked.
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&out).
		Error; err != nil {
		return nil, err
	}

	return &idempotent.EventItem{
		Key:     key,
		Payload: out.Payload,
	}, nil
}
