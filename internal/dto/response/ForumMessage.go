package response

import (
	"time"
)

type ForumMessageResponse struct {
	ID          string `json:"id"`
	ProblemID   string `json:"problem_id"`
	Description []string `json:"desctription"`
	CreatorID   string `json:"creator_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"` 
}