package models

import (
	"time"

	"github.com/google/uuid"
)

type ReportProblem struct {
	ReportID  uuid.UUID  `gorm:"type:uuid;primaryKey;column:report_id;index"`
	ProblemID uuid.UUID  `gorm:"type:uuid;primaryKey;column:problem_id;index"`
	CreatedAt *time.Time `gorm:"type:timestamp"`
	UpdatedAt *time.Time `gorm:"type:timestamp"`
	Deleted   *bool      `gorm:"type:boolean;default:false"`

	Report  *DailyReport `gorm:"foreignKey:ReportID;references:ID;constraint:OnDelete:CASCADE"`
	Problem *Problem     `gorm:"foreignKey:ProblemID;references:ID;constraint:OnDelete:CASCADE"`
}
