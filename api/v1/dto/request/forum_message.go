package request

import "github.com/google/uuid"

type CreateForumMessage struct {
	ProblemID   string   `json:"problem_id"`
	Description []string `json:"description"`
	CreatorID   *string  `json:"creator_id,omitempty"`

	ProblemUUID uuid.UUID  `json:"-"`
	CreatorUUID *uuid.UUID `json:"-"`
	CleanDesc   []string   `json:"-"`
}

type UpdateForumMessage struct {
	Description *[]string `json:"description,omitempty"`
	CleanDesc   *[]string `json:"-"`
}
