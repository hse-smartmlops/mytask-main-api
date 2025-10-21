package response

import (
	"time"
)

type ForumMessageResponse struct {
	ID          string `json:"id"`
	ProblemID   string `json:"problem_id"`
	Description []string `json:"desctription"`
	CreatorID   string `json:"creator_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"` 
}

type ForumMessageListResponse struct {
	Messages   []ForumMessageResponse `json:"messages"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
}

type ForumMessageUniversalResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ForumMessageListByProblemIdResponse struct {
	ProblemId  string                 `json:"problem_id"`
	Messages   []ForumMessageResponse `json:"messages"`
	TotalCount int64                   `json:"total_count"`
	Page       int                     `json:"page"`
	PageSize   int                     `json:"page_size"`
}