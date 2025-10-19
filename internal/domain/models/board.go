package models

import (
	"time"

	"github.com/google/uuid"
)

type Board struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	ProjectID   uuid.UUID  `gorm:"type:uuid;index;not null"`
	Name        *string    `gorm:"size:100"`
	Description *string    `gorm:"type:text"`
	Deleted     *bool      `gorm:"type:boolean"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`

	Project  *Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
	Statuses []Status `gorm:"foreignKey:BoardID;references:ID;constraint:OnDelete:CASCADE"`
}
