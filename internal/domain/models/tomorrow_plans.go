package models

import (
	"time"

	"github.com/google/uuid"
)

type TomorrowPlans struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey"`
	TaskID      *uuid.UUID `gorm:"type:uuid"`
	Description *string    `gorm:"type:text"`
	Deleted     *bool      `gorm:"type:boolean;default:false"`
	ReportID    *uuid.UUID `gorm:"type:uuid"`
	UpdatedAt   *time.Time `gorm:"type:timestamp"`
	CreatedAt   *time.Time `gorm:"type:timestamp"`

	Report *DailyReport `gorm:"foreignKey:ReportID;references:ID"`
	Task   *Task        `gorm:"foreignKey:TaskID;references:ID"`
}
