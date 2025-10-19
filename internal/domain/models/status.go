package models

import (
	"time"

	"github.com/google/uuid"
)

type Status struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	BoardID   uuid.UUID  `gorm:"type:uuid;index;not null"`
	SortOrder *int       `gorm:"type:int"`
	Key       *string    `gorm:"type:varchar(8);uniqueIndex"`
	Name      *string    `gorm:"type:varchar(50)"`
	Color     *string    `gorm:"type:varchar(16)"`
	IsDefault *bool      `gorm:"type:boolean;default:false"`
	IsActive  *bool      `gorm:"type:boolean;default:true"`
	IsOpen    *bool      `gorm:"type:boolean;default:true"`
	CreatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	Deleted   *bool      `gorm:"type:boolean;default:false"`

	Board *Board `gorm:"foreignKey:BoardID;references:ID;constraint:OnDelete:CASCADE"`
	Tasks []Task `gorm:"foreignKey:StatusID;references:ID"`
}
