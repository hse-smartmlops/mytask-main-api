package request

type CreateTask struct {
	StatusID      string  `json:"status_id"`
	Priority      *int16  `json:"priority,omitempty"`
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	CreatedBy     *string `json:"created_by,omitempty"`
	AssignedTo    *string `json:"assigned_to,omitempty"`
	Deadline      *string `json:"deadline,omitempty"`
	StartDate     *string `json:"start_date,omitempty"`
	GitlabIssueID *int    `json:"gitlab_issue_id,omitempty"`
	Category      *int8   `json:"category,omitempty"`
}

type UpdateTask struct {
	StatusID      *string `json:"status_id,omitempty"`
	Priority      *int16  `json:"priority,omitempty"`
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
	CreatedBy     *string `json:"created_by,omitempty"`
	AssignedTo    *string `json:"assigned_to,omitempty"`
	Deadline      *string `json:"deadline,omitempty"`
	StartDate     *string `json:"start_date,omitempty"`
	GitlabIssueID *int    `json:"gitlab_issue_id,omitempty"`
	Category      *int8   `json:"category,omitempty"`
}
