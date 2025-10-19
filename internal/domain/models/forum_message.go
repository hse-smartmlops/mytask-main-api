package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type ForumMessage struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ProblemID   uuid.UUID      `gorm:"type:uuid;index"`
	Description pq.StringArray `gorm:"type:varchar(255)[]"`
	CreatorID   *uuid.UUID     `gorm:"type:uuid;index"`
	CreatedAt   *time.Time     `gorm:"type:timestamp"`
	UpdatedAt   *time.Time     `gorm:"type:timestamp"`
	Deleted     *bool          `gorm:"type:boolean"`

	Problem *Problem `gorm:"foreignKey:ProblemID;references:ID;constraint:OnDelete:CASCADE"`
	User    *User    `gorm:"foreignKey:CreatorID;references:ID;constraint:OnDelete:SET NULL"`
}
