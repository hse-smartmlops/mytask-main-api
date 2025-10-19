package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain/models"
)

func ToUserDTO(user *models.User) response.User {
	if user == nil {
		return response.User{}
	}

	return response.User{
		ID:            user.ID.String(),
		Email:         user.Email,
		IsActive:      user.IsActive,
		TgID:          user.TgID,
		TgUserID:      user.TgUserID,
		Profession:    user.Profession,
		EmailVerified: user.EmailVerified,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		LastLogin:     user.LastLogin,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

func MapUsersPage(page *app.Page[models.User]) (dto.Pagination, response.UsersPage) {
	if page == nil {
		return dto.Pagination{}, response.UsersPage{}
	}

	items := make([]response.User, len(page.Items))
	for i := range page.Items {
		items[i] = ToUserDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.UsersPage{Users: items}
}
