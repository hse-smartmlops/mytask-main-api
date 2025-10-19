package request

type CreateForumMessage struct {
	ProblemID   string   `json:"problem_id"`
	Description []string `json:"description"`
	CreatorID   *string  `json:"creator_id,omitempty"`
}

type UpdateForumMessage struct {
	Description *[]string `json:"description,omitempty"`
}
