package request

type CreateProject struct {
	Name              string `json:"name" validate:"required,min=3,max=100"`
	Description       string `json:"description" validate:"max=500"`
	Priority          int    `json:"priority" validate:"required"`
	Status            string `json:"status" validate:"required"`
	Gitlab_project_id string `json:"gitlab_project_id" validate:"required"`
	Gitlab_url        string `json:"gitlab_url" validate:"required"`
}

type UpdateProject struct {
	Name              *string `json:"name" validate:"omitempty"`
	Description       *string `json:"description" validate:"omitempty"`
	Priority          *int    `json:"priority" validate:"omitempty"`
	Status            *string `json:"status" validate:"omitempty"`
	Gitlab_project_id *string `json:"gitlab_project_id" validate:"omitempty"`
	Gitlab_url        *string `json:"gitlab_url" validate:"omitempty"`
}
