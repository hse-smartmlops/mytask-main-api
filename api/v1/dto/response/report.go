package response

import "time"

type Report struct {
	ID            string              `json:"id"`
	UserID        string              `json:"user_id"`
	ReportDate    *time.Time          `json:"report_date,omitempty"`
	Checked       *int8               `json:"checked,omitempty"`
	CreatedAt     *time.Time          `json:"created_at,omitempty"`
	UpdatedAt     *time.Time          `json:"updated_at,omitempty"`
	CompletedWork []CompletedWorkItem `json:"completed_work,omitempty"`
	TomorrowPlans []TomorrowPlanItem  `json:"plan_tomorrow,omitempty"`
	HelpRequests  []HelpRequestItem   `json:"help_requests,omitempty"`
	Problems      []ProblemItem       `json:"problem,omitempty"`
	UserInfo      *UserSummary        `json:"user_info,omitempty"`
}

type CompletedWorkItem struct {
	ID          string  `json:"id"`
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`
}

type TomorrowPlanItem struct {
	ID          string  `json:"id"`
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`
}

type HelpRequestItem struct {
	ID          string  `json:"id"`
	Description *string `json:"description,omitempty"`
	HelperID    *string `json:"helper_id,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type HelpRequestWithAssigner struct {
	HelpRequest   HelpRequestItem `json:"help_request"`
	UserFirstName *string         `json:"user_first_name,omitempty"`
	UserLastName  *string         `json:"user_last_name,omitempty"`
}

type HelpRequestsForUser struct {
	HelpRequests []HelpRequestWithAssigner `json:"help_requests"`
}

type ProblemItem struct {
	ID          string   `json:"id"`
	Description []string `json:"description,omitempty"`
}

type ReportsPage struct {
	Reports []Report `json:"reports"`
}

type ReportUniversal struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
