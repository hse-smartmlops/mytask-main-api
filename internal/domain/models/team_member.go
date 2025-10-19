package models

import (
	"time"

	"github.com/google/uuid"
)

type TeamMember struct {
	UserID         uuid.UUID  `gorm:"type:uuid;primaryKey;index"`
	TeamID         uuid.UUID  `gorm:"type:uuid;primaryKey;index"`
	Specialization *string    `gorm:"size:50"`
	Deleted        *bool      `gorm:"type:boolean"`
	CreatedAt      *time.Time `gorm:"type:timestamp"`
	UpdatedAt      *time.Time `gorm:"type:timestamp"`

	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	Team *Team `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE"`
}
