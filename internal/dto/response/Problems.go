package response

type ProblemListResponse struct {
	Problems   []ProblemResponse `json:"problems"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}

type ProblemUniversalResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type ProblemsByUserId struct {
	UserID     string            `json:"user_id"`
	Problems   []ProblemResponse `json:"problems"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
}
