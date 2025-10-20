package request

type BoardCreateRequest struct {
	ProjectID   *string `json:"project_id"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type BoardUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

type BoardListRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}
