package request

type AuthLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthRefresh struct {
	RefreshToken string `json:"refresh_token"`
}
