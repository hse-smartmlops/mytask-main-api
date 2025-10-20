package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email         string    `gorm:"size:100"`
	IsActive      bool
	CreatedAt     time.Time `gorm:"type:timestamp"`
	TgID          string    `gorm:"size:50"`
	TgUserID      int64
	Profession    string `gorm:"size:50"`
	EmailVerified bool
	FirstName     string `gorm:"size:50"`
	LastName      string `gorm:"size:50"`
	LastLogin     time.Time
	Deleted       bool      `gorm:"type:boolean"`
	UpdatedAt     time.Time `gorm:"type:timestamp"`
	AvatarPath    string    `gorm:"size:255"`

	UserRoles []UserRole `gorm:"foreignKey:UserID;references:ID"`
}
