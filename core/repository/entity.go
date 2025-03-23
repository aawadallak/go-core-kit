package repository

import (
	"time"
)

// Entity represents a base model structure that provides common fields for database entities.
// It implements soft deletion and includes standard timestamp fields for tracking creation,
// updates, and deletion times.
type Entity struct {
	ID         int32      `json:"-"`                    // Internal database ID
	CreatedAt  time.Time  `json:"created_at"`           // Timestamp when the entity was created
	UpdatedAt  time.Time  `json:"updated_at,omitempty"` // Timestamp when the entity was last updated
	DeletedAt  *time.Time `json:"deleted_at,omitempty"` // Soft deletion timestamp
	ExternalID string     `json:"id"`                   // Public-facing ID used in API responses
}

// GetID returns the entity's internal ID as an unsigned integer.
// This method is used to satisfy interfaces that require a uint ID.
func (e Entity) GetID() uint { return uint(e.ID) }
