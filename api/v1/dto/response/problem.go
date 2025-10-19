package response

import "time"

type Problem struct {
	ID          string       `json:"id"`
	Description []string     `json:"description"`
	CreatorID   *string      `json:"creator_id,omitempty"`
	Name        *string      `json:"name,omitempty"`
	CreatedAt   *time.Time   `json:"created_at,omitempty"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"`
	Creator     *UserSummary `json:"creator,omitempty"`
}

type ProblemsPage struct {
	Problems []Problem `json:"problems"`
}
