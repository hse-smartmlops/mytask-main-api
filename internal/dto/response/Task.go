package response

import (
	"time"
)

type TaskResponse struct {
	ID            string     `json:"id"`
	ProjectID     string     `json:"project_id"`
	BoardID       *string    `json:"board_id,omitempty"`
	Name          string     `json:"name"`
	Description   string     `json:"description"`
	Status        string     `json:"status"`
	Priority      int        `json:"priority"`
	CreatedBy     UserShort  `json:"created_by"`
	AssignedTo    *UserShort `json:"assigned_to,omitempty"`
	Deadline      *time.Time `json:"deadline,omitempty"`
	TimeSpent     int        `json:"time_spent"` 
	StartDate     time.Time  `json:"start_date"`
	GitlabIssueID *int       `json:"gitlab_issue_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type UserShort struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type TaskListResponse struct {
	Tasks      []TaskResponse `json:"tasks"`
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
}

type TaskCreateResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type TaskUpdateResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type TaskDeleteResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}