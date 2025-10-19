package response

import "time"

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	IsActive      bool      `json:"is_active"`
	TgID          string    `json:"tg_id"`
	TgUserID      int64     `json:"tg_user_id"`
	Profession    string    `json:"profession"`
	EmailVerified bool      `json:"email_verified"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	LastLogin     time.Time `json:"last_login"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type UsersPage struct {
	Users []User `json:"users"`
}

type UserSummary struct {
	ID        string  `json:"id"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     string  `json:"email"`
}
