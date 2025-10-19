package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain/models"
)

func ToProblemDTO(problem *models.Problem) response.Problem {
	if problem == nil {
		return response.Problem{}
	}

	var creatorID *string
	if problem.CreatorID != nil {
		id := problem.CreatorID.String()
		creatorID = &id
	}

	return response.Problem{
		ID:          problem.ID.String(),
		Description: problem.Description,
		CreatorID:   creatorID,
		Name:        problem.Name,
		CreatedAt:   problem.CreatedAt,
		UpdatedAt:   problem.UpdatedAt,
		Creator:     toUserSummary(problem.User),
	}
}

func MapProblemsPage(page *ports.Page[models.Problem]) (dto.Pagination, response.ProblemsPage) {
	if page == nil {
		return dto.Pagination{}, response.ProblemsPage{}
	}

	items := make([]response.Problem, len(page.Items))
	for i := range page.Items {
		items[i] = ToProblemDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.ProblemsPage{Problems: items}
}
