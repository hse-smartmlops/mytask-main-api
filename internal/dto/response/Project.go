package response

import (
	"time"
)

type ProjectResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"descroption"`
	GitlabProjectId int    	  `json:"gitlab_project_id"`
	GitlabUrl       string    `json:"gitlab_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProjectListResponse struct {
	Projects   []ProjectResponse `json:"projects"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

type ProjectUniversalResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}
