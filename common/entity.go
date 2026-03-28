package common

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Entity struct {
	ID         uint       `json:"-" gorm:"primaryKey"`
	ExternalID string     `json:"id" gorm:"uniqueIndex;not null"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at,omitzero"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
}

func (e Entity) GetID() uint { return e.ID } //nolint:gocritic // value receiver needed for interface satisfaction

func (e *Entity) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ExternalID == "" {
		e.ExternalID = uuid.New().String()
	}
	return nil
}
