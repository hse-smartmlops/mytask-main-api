package request

import (
	"time"

	"github.com/google/uuid"
)

type ProjectReport struct {
	Description string `json:"description"`
}

type HelpRequest struct {
	HelperID    string `json:"helper_id"`
	Description string `json:"description"`
}

type ReportCreateRequest struct {
	UserId       string           `json:"user_id"`
	TaskId       string           `json:"task_id"`
	ReportDate   *time.Time       `json:"report_date"`
	Status       *string          `json:"status"`
	CompleteWork *[]ProjectReport `json:"complete_work"`
	PlanTomorrow *[]ProjectReport `json:"plan_tomorrow"`
	ProblemsIds  *[]string     `json:"problemsIds"`
	Help         *HelpRequest     `json:"help"`
}

type ReportUpdateRequest struct {
	UserId     *string    `json:"user_id"`
	ReportDate *time.Time `json:"report_date"`
	Status       *string          `json:"status"`
}

type HelpRequestUpdateRequest struct{
	HelperID    *string     `json:"helper_id"`
	Description *string     `json:"description"`
}

type CompletedWorkUpdateRequest struct{
	Description  *string   `json:"description"`
}

type TomorrowPlansUpdateRequest struct{
	Description  *string   `json:"description"`
}

type ReportListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type ReportsByProjectId struct {
	ProjectID string `json:"project_id"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type Problem struct {
	ID          string    `json:"id"`
	Description []string  `json:"description"`
	CreatorId   uuid.UUID `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name"`
}

