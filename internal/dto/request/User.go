package request

import "time"

type UserCreateRequest struct {
	Email          *string    `json:"email"`
	IsActive       *bool      `json:"is_active"`
	CreatedAt      *time.Time `json:"created_at"`
	TgId           *string    `json:"tg_id"`
	TgUserId       *int64     `json:"tg_user_id"`
	Profession     *string    `json:"profession"`
	EmailVerified  *bool      `json:"email_verified"`
	FirstName      *string    `json:"first_name"`
	LastName       *string    `json:"last_name"`
	LastLogin      *time.Time `json:"last_login"`
	AuthProviderId *string    `json:"auth_provider_id"`
}

type UpdateUserRequest struct {
	Email          *string    `json:"email"`
	IsActive       *bool      `json:"is_active"`
	CreatedAt      *time.Time `json:"created_at"`
	TgId           *string    `json:"tg_id"`
	TgUserId       *int64     `json:"tg_user_id"`
	Profession     *string    `json:"profession"`
	EmailVerified  *bool      `json:"email_verified"`
	FirstName      *string    `json:"first_name"`
	LastName       *string    `json:"last_name"`
	LastLogin      *time.Time `json:"last_login"`
	AuthProviderId *string    `json:"auth_provider_id"`
}

type GetAllUsersRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type AddRoleUserRequest struct {
	RoleId string `json:"role_id"`
	UserId string `json:"user_id"`
}

type RemoveRoleUserRequest struct {
	RoleId string `json:"role_id"`
	UserId string `json:"user_id"`
}
