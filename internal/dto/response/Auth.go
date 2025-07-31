package response

// DTO для ответа при успешной аутентификации
// Например, после login, oauth, totp

type AuthResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// DTO для ответа с информацией о пользователе
// GET /auth/me

type UserInfoResponse struct {
	UserID     string   `json:"user_id"`
	Email      string   `json:"email"`
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name"`
	Profession string   `json:"profession"`
	IsActive   bool     `json:"is_active"`
	Roles      []string `json:"roles"`
}
