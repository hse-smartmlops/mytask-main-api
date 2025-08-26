package response

import (
	"time"
)

type ReportUniversalResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ReportResponse struct {
	ID            string            `json:"id"`
	UserID        string            `json:"user_id"`
	ReportDate    time.Time         `json:"report_date"`
	CompletedWork []WorkItem        `json:"completed_work"`
	PlanTomorrow  []WorkItem        `json:"plan_tomorrow"`
	HelpRequest   HelpRequestItem   `json:"help_request,omitempty"`
	Status        string            `json:"status"`
	TaskId        string            `json:"task_id"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	UserInfo      UserShort         `json:"user_info"`
	Problems      []ProblemResponse `json:"problem"`
}

type ProblemResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`		
	Description []string  `json:"description"`
	CreatorId   string    `json:"creator_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkItem struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

type HelpRequestItem struct {
	ID          string `json:"id"`
	HelperID    string `json:"helper_id"`
	Description string `json:"description"`
}

type ReportListResponse struct {
	Reports    []ReportResponse `json:"reports"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

type ReportListByTaskId struct {
	TaskID  string           `json:"task_id"`
	Reports []ReportResponse `json:"reports"`
}

type ReportListByProjectId struct {
	ProjectID string           `json:"project_id"`
	Reports   []ReportResponse `json:"reports"`
}
