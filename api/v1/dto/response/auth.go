package response

import "time"

type Auth struct {
	UserID       string    `json:"user_id"`
	Email        string    `json:"email"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshExp   int       `json:"refresh_expires_in"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Refresh struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshExp   int       `json:"refresh_expires_in"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type UserInfo struct {
	Sub               string  `json:"sub"`
	Name              *string `json:"name,omitempty"`
	PreferredUsername *string `json:"preferred_username,omitempty"`
	GivenName         *string `json:"given_name,omitempty"`
	FamilyName        *string `json:"family_name,omitempty"`
	Email             *string `json:"email,omitempty"`
	EmailVerified     bool    `json:"email_verified"`
}

type TokenValidation struct {
	Message string `json:"message"`
}
