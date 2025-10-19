package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Problem struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Description pq.StringArray `gorm:"type:varchar(255)[]"`
	CreatorID   *uuid.UUID     `gorm:"type:uuid;index"`
	Name        *string        `gorm:"type:varchar(255)"`
	CreatedAt   *time.Time     `gorm:"type:timestamp"`
	UpdatedAt   *time.Time     `gorm:"type:timestamp"`
	Deleted     *bool          `gorm:"type:boolean"`

	User  *User          `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:CASCADE"`
	Forum []ForumMessage `gorm:"foreignKey:ProblemID;references:ID"`
}
