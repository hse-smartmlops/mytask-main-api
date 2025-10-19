package models

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name        *string    `gorm:"size:50"`
	Description *string    `gorm:"size:200"`
	Deleted     *bool      `gorm:"type:boolean"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`

	UserRoles []UserRole `gorm:"foreignKey:RoleID;references:ID"`
}
