package response

import "time"

type BoardUniversalResponse struct {
	Id      string `json:"id"`
	Message string `json:"message"`
}

type BoardResponse struct {
	Id          string    `json:"id"`
	ProjectId   string    `json:"project_id"`
	Name        string    `json:"name"`
	Description string    `json:"descroption"`
	Filter      string    `json:"filter"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BoardListResponse struct {
	Boards     []BoardResponse `json:"boards"`
	TotalCount int             `json:"total_count"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

type BoardForProjectResponse struct {
	Boards    []BoardResponse `json:"board"`
	ProjectId string          `json:"project_id"`
}
