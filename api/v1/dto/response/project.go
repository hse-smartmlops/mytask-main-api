package response

import "time"

type Project struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Description     *string    `json:"description,omitempty"`
	GitlabProjectID *int       `json:"gitlab_project_id,omitempty"`
	GitlabURL       *string    `json:"gitlab_url,omitempty"`
	CreatedBy       *string    `json:"created_by,omitempty"`
	Status          *string    `json:"status,omitempty"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
}

type ProjectsPage struct {
	Projects []Project `json:"projects"`
}
