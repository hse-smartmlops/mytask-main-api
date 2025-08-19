package request

type ProblemCreateRequest struct {
	Description *[]string `json:"description"`
	CreatorID   string    `json:"creator_id"`
}

type ProblemUpdateRequest struct {
	Description *[]string `json:"description"`
}

type ProblemsListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}