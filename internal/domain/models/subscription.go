package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID         uuid.UUID  `gorm:"type:uuid;index"`
	SubscriptionID *uuid.UUID `gorm:"type:uuid"`
	TypeID         *int8      `gorm:"type:int8"`
	CreatedAt      *time.Time `gorm:"type:timestamp"`
	UpdatedAt      *time.Time `gorm:"type:timestamp"`
	Deleted        *bool      `gorm:"type:boolean;default:false"`

	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
}
