package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain/models"
)

func ToStatusDTO(status *models.Status) response.Status {
	if status == nil {
		return response.Status{}
	}

	return response.Status{
		ID:        status.ID.String(),
		BoardID:   status.BoardID.String(),
		Name:      valueOrEmpty(status.Name),
		Key:       status.Key,
		Color:     status.Color,
		IsDefault: status.IsDefault,
		IsActive:  status.IsActive,
		IsOpen:    status.IsOpen,
		SortOrder: status.SortOrder,
		CreatedAt: status.CreatedAt,
		UpdatedAt: status.UpdatedAt,
	}
}

func MapStatusesPage(page *app.Page[models.Status]) (dto.Pagination, response.StatusesPage) {
	if page == nil {
		return dto.Pagination{}, response.StatusesPage{}
	}

	items := make([]response.Status, len(page.Items))
	for i := range page.Items {
		items[i] = ToStatusDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.StatusesPage{Statuses: items}
}
