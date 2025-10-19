package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain/models"
)

func ToRoleDTO(role *models.Role) response.Role {
	if role == nil {
		return response.Role{}
	}

	var description *string
	if role.Description != nil && *role.Description != "" {
		description = role.Description
	}

	return response.Role{
		ID:          role.ID.String(),
		Name:        valueOrEmpty(role.Name),
		Description: description,
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}

func MapRolesPage(page *ports.Page[models.Role]) (dto.Pagination, response.RolesPage) {
	if page == nil {
		return dto.Pagination{}, response.RolesPage{}
	}

	items := make([]response.Role, len(page.Items))
	for i := range page.Items {
		items[i] = ToRoleDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.RolesPage{Roles: items}
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
