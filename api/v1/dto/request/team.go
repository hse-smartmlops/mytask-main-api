package request

import "github.com/google/uuid"

type CreateTeam struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type UpdateTeam struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

type AddTeamMember struct {
	UserID         string  `json:"user_id"`
	Specialization *string `json:"specialization,omitempty"`

	UserUUID uuid.UUID `json:"-"`
}

type AddTeamProject struct {
	ProjectID string `json:"project_id"`

	ProjectUUID uuid.UUID `json:"-"`
}

type LegacyCreateTeam struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	UserIDs     []string `json:"user_ids,omitempty"`
}

type TeamAddUsers struct {
	TeamID  string   `json:"team_id"`
	UserIDs []string `json:"user_ids"`
}

type TeamRemoveUser struct {
	TeamID string `json:"team_id"`
	UserID string `json:"user_id"`
}

type TeamProjectRequest struct {
	TeamID    string `json:"team_id"`
	ProjectID string `json:"project_id"`
}
