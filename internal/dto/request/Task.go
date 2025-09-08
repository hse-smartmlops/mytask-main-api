package request

import "time"

type TaskCreateRequest struct {
	ProjectID     string     `json:"project_id" validate:"required,uuid4"`
	Name          *string     `json:"name" validate:"required,min=3,max=100"`
	Description   *string     `json:"description" validate:"max=500"`
	Priority      *int16     `json:"priority" validate:"required,min=1,max=10"`
	CreatorID     *string     `json:"creator_id" validate:"required,uuid4"`
	AssignedTo    *string    `json:"assigned_to" validate:"omitempty,uuid4"`
	Deadline      *time.Time `json:"deadline"`
	StartDate     *time.Time `json:"start_date" validate:"required"`
	GitlabIssueID *int       `json:"gitlab_issue_id"`
	Category      *int8      `json:"category"`
}

type TaskUpdateRequest struct {
	Name          *string    `json:"name" validate:"omitempty,min=3,max=100"`
	Description   *string    `json:"description" validate:"omitempty,max=500"`
	Priority      *int16     `json:"priority" validate:"omitempty,min=1,max=10"`
	AssignedTo    *string    `json:"assigned_to" validate:"omitempty,uuid4"`
	Deadline      *time.Time `json:"deadline"`
	StartDate     *time.Time `json:"start_date"`
	GitlabIssueID *int       `json:"gitlab_issue_id"`
	Category     *bool      `json:"category"`
}

type TaskListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}
