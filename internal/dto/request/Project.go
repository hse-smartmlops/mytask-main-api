package request

type CreateProjectRequest struct {
	Name              *string `json:"name" validate:"required,min=3,max=100"`
	Description       *string `json:"description" validate:"max=500"`
	Status            *string `json:"status" validate:"required"`
	Gitlab_project_id *int    `json:"gitlab_project_id" validate:"required"`
	Gitlab_url        *string `json:"gitlab_url" validate:"required"`
}

type UpdateProjectRequest struct {
	Name            *string `json:"name" validate:"omitempty"`
	Description     *string `json:"description" validate:"omitempty"`
	Status          *string `json:"status" validate:"omitempty"`
	GitlabProjectId *int    `json:"gitlab_project_id" validate:"omitempty"`
	GitlabUrl       *string `json:"gitlab_url" validate:"omitempty"`
}

type ProjectListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}
