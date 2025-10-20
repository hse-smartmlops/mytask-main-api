package models

import (
	"time"

	"github.com/google/uuid"
)

type DailyReport struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID  `gorm:"type:uuid;index"`
	Checked       *int8      `gorm:"type:int8"`
	ReportDate    *time.Time `gorm:"type:date"`
	CreatedAt     *time.Time `gorm:"type:date"`
	UpdatedAt     *time.Time `gorm:"type:date"`
	Deleted       *bool      `gorm:"type:boolean"`
	StorageObject string     `gorm:"size:255"`

	User           *User           `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE"`
	HelpRequests   []HelpRequest   `gorm:"foreignKey:ReportID;references:ID"`
	CompletedWork  []CompletedWork `gorm:"foreignKey:ReportID;references:ID"`
	TomorrowPlans  []TomorrowPlans `gorm:"foreignKey:ReportID;references:ID"`
	ReportProblems []ReportProblem `gorm:"foreignKey:ReportID;references:ID"`
}
