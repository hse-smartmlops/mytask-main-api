package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	RoleID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey"`
	AssignedAt *time.Time
	AssignedBy *uuid.UUID `gorm:"type:uuid"`
	Deleted    *bool      `gorm:"type:boolean"`
	CreatedAt  *time.Time `gorm:"type:timestamp"`
	UpdatedAt  *time.Time `gorm:"type:timestamp"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
