package request

type CreateRole struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateRole struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
