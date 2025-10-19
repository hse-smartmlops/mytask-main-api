package request

type ReportCompletedWork struct {
	ID          *string `json:"id,omitempty"`
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`
}

type ReportHelpRequest struct {
	ID          *string `json:"id,omitempty"`
	Description *string `json:"description,omitempty"`
	HelperID    *string `json:"helper_id,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type ReportTomorrowPlan struct {
	ID          *string `json:"id,omitempty"`
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`
}

type ReportCreate struct {
	UserID        string                `json:"user_id"`
	ReportDate    *string               `json:"report_date,omitempty"`
	CompletedWork []ReportCompletedWork `json:"complete_work,omitempty"`
	HelpRequests  []ReportHelpRequest   `json:"help,omitempty"`
	TomorrowPlans []ReportTomorrowPlan  `json:"plan_tomorrow,omitempty"`
	Problems      []string              `json:"problems,omitempty"`
}

type ReportUpdate struct {
	Checked       *int                  `json:"checked,omitempty"`
	ReportDate    *string               `json:"report_date,omitempty"`
	CompletedWork []ReportCompletedWork `json:"complete_work,omitempty"`
	HelpRequests  []ReportHelpRequest   `json:"help,omitempty"`
	TomorrowPlans []ReportTomorrowPlan  `json:"plan_tomorrow,omitempty"`
	Problems      []string              `json:"problems,omitempty"`
}

type CompletedWorkUpdate struct {
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`
}

type HelpRequestUpdate struct {
	Description *string `json:"description,omitempty"`
	HelperID    *string `json:"helper_id,omitempty"`
	Status      *string `json:"status,omitempty"`
}

type TomorrowPlanUpdate struct {
	Description *string `json:"description,omitempty"`
	TaskID      *string `json:"task_id,omitempty"`
}

type ReportsByDate struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}
