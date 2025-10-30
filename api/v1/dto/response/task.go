package response

import "time"

type Task struct {
	ID             string       `json:"id"`
	ProjectID      *string      `json:"project_id,omitempty"`
	StatusID       string       `json:"status_id"`
	Priority       *int16       `json:"priority,omitempty"`
	Name           string       `json:"name"`
	Description    *string      `json:"description,omitempty"`
	CreatedBy      *string      `json:"created_by,omitempty"`
	AssignedTo     *string      `json:"assigned_to,omitempty"`
	Deadline       *time.Time   `json:"deadline,omitempty"`
	StartDate      *time.Time   `json:"start_date,omitempty"`
	GitlabIssueID  *int         `json:"gitlab_issue_id,omitempty"`
	Category       *int8        `json:"category,omitempty"`
	CreatedAt      *time.Time   `json:"created_at,omitempty"`
	UpdatedAt      *time.Time   `json:"updated_at,omitempty"`
	Status         *TaskStatus  `json:"status,omitempty"`
	CreatedByUser  *UserSummary `json:"created_by_user,omitempty"`
	AssignedToUser *UserSummary `json:"assigned_to_user,omitempty"`
}

type TaskStatus struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Key       *string `json:"key,omitempty"`
	Color     *string `json:"color,omitempty"`
	IsDefault *bool   `json:"is_default,omitempty"`
	SortOrder *int    `json:"sort_order,omitempty"`
	IsActive  *bool   `json:"is_active,omitempty"`
	IsOpen    *bool   `json:"is_open,omitempty"`
}

type TasksPage struct {
	Tasks []Task `json:"tasks"`
}
