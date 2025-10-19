package request

type CreateProject struct {
	Name            string  `json:"name"`
	Description     *string `json:"description,omitempty"`
	GitlabProjectID *int    `json:"gitlab_project_id,omitempty"`
	GitlabURL       *string `json:"gitlab_url,omitempty"`
	CreatedBy       *string `json:"created_by,omitempty"`
	Status          *string `json:"status,omitempty"`
}

type UpdateProject struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	GitlabProjectID *int    `json:"gitlab_project_id,omitempty"`
	GitlabURL       *string `json:"gitlab_url,omitempty"`
	Status          *string `json:"status,omitempty"`
}
