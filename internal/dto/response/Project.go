package response

import (
	"time"
)

type ProjectResponse struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Description       *string   `json:"descroption"`
	Status            string    `json:"status"`
	Gitlab_project_id string    `json:"gitlab_project_id"`
	Gitlab_url        string    `json:"gitlab_url"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type ProjectListResponse struct {
	Prijects   []ProjectResponse `json:"projects"`
	TotalCount int               `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

type ProjectDeleteResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
