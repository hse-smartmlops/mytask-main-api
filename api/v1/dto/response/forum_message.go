package response

import "time"

type ForumMessage struct {
	ID          string     `json:"id"`
	ProblemID   string     `json:"problem_id"`
	Description []string   `json:"description"`
	CreatorID   *string    `json:"creator_id,omitempty"`
	CreatedAt   *time.Time `json:"created_at,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type ForumMessagesPage struct {
	Messages []ForumMessage `json:"messages"`
}
