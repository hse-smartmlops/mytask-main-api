package request

import "time"

type CreateTask struct {
	ProjectID     string     `json:"project_id" validate:"required,uuid4"`
	BoardID       *string    `json:"board_id" validate:"omitempty,uuid4"`
	Name          string     `json:"name" validate:"required,min=3,max=100"`
	Description   string     `json:"description" validate:"max=500"`
	Status        string     `json:"status" validate:"required,oneof=todo in_progress done"`
	Priority      int        `json:"priority" validate:"required,min=1,max=10"`
	AssignedTo    *string    `json:"assigned_to" validate:"omitempty,uuid4"`
	Deadline      *time.Time `json:"deadline"`
	StartDate     time.Time  `json:"start_date" validate:"required"`
	GitlabIssueID *int       `json:"gitlab_issue_id"`
}

type UpdateTask struct {
	Name          *string    `json:"name" validate:"omitempty,min=3,max=100"`
	Description   *string    `json:"description" validate:"omitempty,max=500"`
	Status        *string    `json:"status" validate:"omitempty,oneof=todo in_progress done"`
	Priority      *int       `json:"priority" validate:"omitempty,min=1,max=10"`
	AssignedTo    *string    `json:"assigned_to" validate:"omitempty,uuid4"`
	Deadline      *time.Time `json:"deadline"`
	StartDate     *time.Time `json:"start_date"`
	GitlabIssueID *int       `json:"gitlab_issue_id"`
	BoardID       *string    `json:"board_id" validate:"omitempty,uuid4"`
}

type TaskFilter struct {
	ProjectID    *string    `json:"project_id" form:"project_id" validate:"omitempty,uuid4"`
	BoardID      *string    `json:"board_id" form:"board_id" validate:"omitempty,uuid4"`
	Status       *string    `json:"status" form:"status" validate:"omitempty,oneof=todo in_progress done"`
	Priority     *int       `json:"priority" form:"priority" validate:"omitempty,min=1,max=10"`
	AssignedTo   *string    `json:"assigned_to" form:"assigned_to" validate:"omitempty,uuid4"`
	CreatedBy    *string    `json:"created_by" form:"created_by" validate:"omitempty,uuid4"`
	DeadlineFrom *time.Time `json:"deadline_from" form:"deadline_from"`
	DeadlineTo   *time.Time `json:"deadline_to" form:"deadline_to"`
	Page         *int       `json:"page" form:"page" validate:"omitempty,min=1"`
	PageSize     *int       `json:"page_size" form:"page_size" validate:"omitempty,min=5,max=100"`
}
