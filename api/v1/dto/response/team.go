package response

import "time"

type Team struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description *string      `json:"description,omitempty"`
	CreatedAt   *time.Time   `json:"created_at,omitempty"`
	UpdatedAt   *time.Time   `json:"updated_at,omitempty"`
	Members     []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID         string  `json:"user_id"`
	FirstName      *string `json:"first_name,omitempty"`
	LastName       *string `json:"last_name,omitempty"`
	Email          string  `json:"email"`
	Specialization *string `json:"specialization,omitempty"`
}

type TeamsPage struct {
	Teams []Team `json:"teams"`
}
