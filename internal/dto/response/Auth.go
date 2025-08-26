package response

import "time"

// DTO для ответа при успешной аутентификации
// Например, после login, oauth, totp

type AuthResponse struct {
	UserID       string `json:"userId"`
	Email        string `json:"email,omitempty"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`    // время жизни access token (секунды)
	RefreshExp   int    `json:"refresh_exp"`   // время жизни refresh token (секунды)
	TokenType    string `json:"token_type"`    // обычно "Bearer"
	ExpiresAt    string `json:"expires_at"`
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

// UserInfo godoc
// @Description Структура с информацией о пользователе из Keycloak
type UserInfo struct {
    // ID пользователя
    Sub string `json:"sub"`
    // Полное имя пользователя
    Name string `json:"name,omitempty"`
    // Имя пользователя (логин)
    PreferredUsername string `json:"preferred_username,omitempty"`
    // Имя
    GivenName string `json:"given_name,omitempty"`
    // Фамилия
    FamilyName string `json:"family_name,omitempty"`
    // Email пользователя
    Email string `json:"email,omitempty"`
    // Подтверждён ли email
    EmailVerified bool `json:"email_verified,omitempty"`
}

type RefreshResponse struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
    RefreshExp   int    `json:"refresh_exp"`
    TokenType    string `json:"token_type"`
    ExpiresAt    time.Time `json:"expires_at"`
}

type TokenValidationResponse struct {
    Message string `json:"message"` // "Token is valid" или описание ошибки
}
