package models

import (
	"time"

	"github.com/google/uuid"
)

type HelpRequest struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	HelperID    *uuid.UUID `gorm:"type:uuid"`
	Description *string    `gorm:"type:varchar(255)"`
	Deleted     *bool      `gorm:"type:boolean;default:false"`
	ReportID    *uuid.UUID `gorm:"type:uuid"`
	Status      *string    `gorm:"type:varchar(30)"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`

	Helper *User        `gorm:"foreignKey:HelperID;references:ID"`
	Report *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
}
