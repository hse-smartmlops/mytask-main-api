package models

import (
	"time"

	"github.com/google/uuid"
)

type ProjectTeam struct {
	ProjectID uuid.UUID  `gorm:"type:uuid;primaryKey;index"`
	TeamID    uuid.UUID  `gorm:"type:uuid;primaryKey;index"`
	Deleted   *bool      `gorm:"type:boolean"`
	CreatedAt *time.Time `gorm:"type:timestamp"`
	UpdatedAt *time.Time `gorm:"type:timestamp"`

	Project *Project `gorm:"foreignKey:ProjectID;references:ID;constraint:OnDelete:CASCADE"`
	Team    *Team    `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE"`
}
