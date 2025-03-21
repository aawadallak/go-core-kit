package idgorm

import (
	"context"
	"errors"

	"github.com/aawadallak/go-core-kit/core/idempotent"
	"gorm.io/gorm"
)

var (
	ErrDatabaseConnectionRequired = errors.New("database connection is required")
	ErrKeyRequired                = errors.New("key is required")
	ErrPayloadRequired            = errors.New("payload is required")
)

const (
	// TableName is the name of the table used for storing idempotent events
	TableName = "tb_tx_outbox"
)

// Entity represents the database model for idempotent events
type Entity struct {
	gorm.Model
	Key     string `gorm:"uniqueIndex;not null;size:255"`
	Payload []byte `gorm:"not null"`
}

// TableName returns the name of the table for the Entity
func (e *Entity) TableName() string {
	return TableName
}

// Repository implements the idempotent.Repository interface using GORM
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new Repository instance with the given GORM DB connection
func NewRepository(db *gorm.DB) (*Repository, error) {
	if db == nil {
		return nil, ErrDatabaseConnectionRequired
	}

	if err := db.AutoMigrate(&Entity{}); err != nil {
		return nil, err
	}

	return &Repository{
		db: db,
	}, nil
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

	out := &Entity{}
	if err := i.db.
		WithContext(ctx).
		Where("key = ?", key).
		First(out).
		Error; err != nil {
		return nil, err
	}

	return &idempotent.EventItem{
		Key:     key,
		Payload: out.Payload,
	}, nil
}
