package request

type CreateUser struct {
	Email         string  `json:"email"`
	IsActive      *bool   `json:"is_active,omitempty"`
	TgID          *string `json:"tg_id,omitempty"`
	TgUserID      *int64  `json:"tg_user_id,omitempty"`
	Profession    *string `json:"profession,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
	FirstName     *string `json:"first_name,omitempty"`
	LastName      *string `json:"last_name,omitempty"`
}

type UpdateUser struct {
	Email      string  `json:"email,omitempty"`
	IsActive   *bool   `json:"is_active,omitempty"`
	TgID       *string `json:"tg_id,omitempty"`
	Profession *string `json:"profession,omitempty"`
	FirstName  *string `json:"first_name,omitempty"`
	LastName   *string `json:"last_name,omitempty"`
}

type UserRoleAssignment struct {
}
