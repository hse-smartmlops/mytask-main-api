package request

import (
	"time"

	"github.com/google/uuid"
)

type ReportCompletedWork struct {
	ID          *string `json:"id,omitempty"`
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`

	IDUUID   *uuid.UUID `json:"-"`
	TaskUUID *uuid.UUID `json:"-"`
}

type ReportHelpRequest struct {
	ID          *string `json:"id,omitempty"`
	Description *string `json:"description,omitempty"`
	HelperID    *string `json:"helper_id,omitempty"`
	Status      *string `json:"status,omitempty"`

	IDUUID     *uuid.UUID `json:"-"`
	HelperUUID *uuid.UUID `json:"-"`
}

type ReportTomorrowPlan struct {
	ID          *string `json:"id,omitempty"`
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`

	IDUUID   *uuid.UUID `json:"-"`
	TaskUUID *uuid.UUID `json:"-"`
}

type ReportCreate struct {
	UserID        string                `json:"user_id"`
	ReportDate    *string               `json:"report_date,omitempty"`
	CompletedWork []ReportCompletedWork `json:"complete_work,omitempty"`
	HelpRequests  []ReportHelpRequest   `json:"help,omitempty"`
	TomorrowPlans []ReportTomorrowPlan  `json:"plan_tomorrow,omitempty"`
	Problems      []string              `json:"problems,omitempty"`

	UserUUID        uuid.UUID   `json:"-"`
	ReportDateValue *time.Time  `json:"-"`
	ProblemUUIDs    []uuid.UUID `json:"-"`
}

type ReportUpdate struct {
	Checked       *int                  `json:"checked,omitempty"`
	ReportDate    *string               `json:"report_date,omitempty"`
	CompletedWork []ReportCompletedWork `json:"complete_work,omitempty"`
	HelpRequests  []ReportHelpRequest   `json:"help,omitempty"`
	TomorrowPlans []ReportTomorrowPlan  `json:"plan_tomorrow,omitempty"`
	Problems      []string              `json:"problems,omitempty"`

	CheckedValue    *int8       `json:"-"`
	ReportDateValue *time.Time  `json:"-"`
	ProblemUUIDs    []uuid.UUID `json:"-"`
}

type CompletedWorkUpdate struct {
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`

	TaskUUID *uuid.UUID `json:"-"`
}

type HelpRequestUpdate struct {
	Description *string `json:"description,omitempty"`
	HelperID    *string `json:"helper_id,omitempty"`
	Status      *string `json:"status,omitempty"`

	HelperUUID *uuid.UUID `json:"-"`
}

type TomorrowPlanUpdate struct {
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`

	TaskUUID *uuid.UUID `json:"-"`
}

type ReportsByDate struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`

	StartValue time.Time `json:"-"`
	EndValue   time.Time `json:"-"`
}
