package request

type CreateProblem struct {
	Description []string `json:"description"`
	CreatorID   *string  `json:"creator_id,omitempty"`
	Name        *string  `json:"name,omitempty"`
}

type UpdateProblem struct {
	Description *[]string `json:"description,omitempty"`
	Name        *string   `json:"name,omitempty"`
}
