package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain/models"
)

func ToProjectDTO(project *models.Project) response.Project {
	if project == nil {
		return response.Project{}
	}

	return response.Project{
		ID:              project.ID.String(),
		Name:            valueOrEmpty(project.Name),
		Description:     project.Description,
		GitlabProjectID: project.GitlabProjectID,
		GitlabURL:       project.GitlabURL,
		CreatedBy:       uuidString(project.CreatedBy),
		Status:          project.Status,
		CreatedAt:       project.CreatedAt,
		UpdatedAt:       project.UpdatedAt,
	}
}

func MapProjectsPage(page *ports.Page[models.Project]) (dto.Pagination, response.ProjectsPage) {
	if page == nil {
		return dto.Pagination{}, response.ProjectsPage{}
	}

	items := make([]response.Project, len(page.Items))
	for i := range page.Items {
		items[i] = ToProjectDTO(&page.Items[i])
	}

	pagination := dto.Pagination{
		Page:       page.Page,
		PageSize:   page.PageSize,
		TotalCount: page.TotalCount,
		TotalPages: page.TotalPages(),
	}

	return pagination, response.ProjectsPage{Projects: items}
}
