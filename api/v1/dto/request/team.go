package request

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
}

type AddTeamProject struct {
	ProjectID string `json:"project_id"`
}
