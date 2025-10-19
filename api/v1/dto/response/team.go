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

type TeamMessage struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type TeamUserMessage struct {
	TeamID  string `json:"team_id"`
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

type TeamProjectMessage struct {
	TeamID    string `json:"team_id"`
	ProjectID string `json:"project_id"`
	Message   string `json:"message"`
}

type UsersAdd struct {
	TeamID  string   `json:"team_id"`
	UserIDs []string `json:"users_id"`
	Message string   `json:"message"`
}
