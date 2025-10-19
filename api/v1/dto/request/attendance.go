package request

type CreateAttendance struct {
	UserID        string  `json:"user_id"`
	Date          *string `json:"date,omitempty"`
	WorkdayHours  *int    `json:"workday_hours,omitempty"`
	PlannedStart  *string `json:"planned_start,omitempty"`
	ActualStart   *string `json:"actual_start,omitempty"`
	Commits       *int    `json:"commits,omitempty"`
	MergeRequests *int    `json:"merge_requests,omitempty"`
	CodeReviews   *int    `json:"code_reviews,omitempty"`
	EndWork       *string `json:"end_work,omitempty"`
	Status        *string `json:"status,omitempty"`
}

type UpdateAttendance struct {
	Date          *string `json:"date,omitempty"`
	WorkdayHours  *int    `json:"workday_hours,omitempty"`
	PlannedStart  *string `json:"planned_start,omitempty"`
	ActualStart   *string `json:"actual_start,omitempty"`
	Commits       *int    `json:"commits,omitempty"`
	MergeRequests *int    `json:"merge_requests,omitempty"`
	CodeReviews   *int    `json:"code_reviews,omitempty"`
	EndWork       *string `json:"end_work,omitempty"`
	Status        *string `json:"status,omitempty"`
}
