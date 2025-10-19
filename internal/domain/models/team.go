package models

import (
	"time"

	"github.com/google/uuid"
)

type Team struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Name        *string    `gorm:"size:100"`
	Description *string    `gorm:"type:text"`
	Deleted     *bool      `gorm:"type:boolean"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`

	TeamMembers  []TeamMember  `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE"`
	ProjectTeams []ProjectTeam `gorm:"foreignKey:TeamID;references:ID;constraint:OnDelete:CASCADE"`
}
