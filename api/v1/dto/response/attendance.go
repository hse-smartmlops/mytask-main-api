package response

import "time"

type Attendance struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	Date          *time.Time `json:"date,omitempty"`
	WorkdayHours  *int16     `json:"workday_hours,omitempty"`
	PlannedStart  *time.Time `json:"planned_start,omitempty"`
	ActualStart   *time.Time `json:"actual_start,omitempty"`
	Commits       *int16     `json:"commits,omitempty"`
	MergeRequests *int16     `json:"merge_requests,omitempty"`
	CodeReviews   *int16     `json:"code_reviews,omitempty"`
	EndWork       *time.Time `json:"end_work,omitempty"`
	Status        *string    `json:"status,omitempty"`
	CreatedAt     *time.Time `json:"created_at,omitempty"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

type AttendancesPage struct {
	Attendances []Attendance `json:"attendances"`
}
