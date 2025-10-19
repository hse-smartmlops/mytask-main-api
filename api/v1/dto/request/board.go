package request

type CreateBoard struct {
	ProjectID   string  `json:"project_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateBoard struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
