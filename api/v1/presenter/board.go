package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain/models"
)

func ToBoardDTO(board *models.Board) response.Board {
	if board == nil {
		return response.Board{}
	}

	statuses := make([]response.BoardStatus, 0, len(board.Statuses))
	for _, status := range board.Statuses {
		statuses = append(statuses, response.BoardStatus{
			ID:        status.ID.String(),
			Key:       status.Key,
			Name:      valueOrEmpty(status.Name),
			Color:     status.Color,
			IsDefault: status.IsDefault,
			SortOrder: status.SortOrder,
			IsActive:  status.IsActive,
			IsOpen:    status.IsOpen,
			CreatedAt: status.CreatedAt,
			UpdatedAt: status.UpdatedAt,
		})
	}

	return response.Board{
		ID:          board.ID.String(),
		ProjectID:   board.ProjectID.String(),
		Name:        valueOrEmpty(board.Name),
		Description: board.Description,
		CreatedAt:   board.CreatedAt,
		UpdatedAt:   board.UpdatedAt,
		Statuses:    statuses,
	}
}

func MapBoardsPage(page *ports.Page[models.Board]) (dto.Pagination, response.BoardsPage) {
	if page == nil {
		return dto.Pagination{}, response.BoardsPage{}
	}

	items := make([]response.Board, len(page.Items))
	for i := range page.Items {
		items[i] = ToBoardDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.BoardsPage{Boards: items}
}
