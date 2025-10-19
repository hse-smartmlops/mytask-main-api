package presenter

import (
	"emplacc-api/api/v1/dto"
	"emplacc-api/api/v1/dto/response"
	"emplacc-api/internal/app"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
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

func MapProjectsPage(page *app.Page[models.Project]) (dto.Pagination, response.ProjectsPage) {
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

func uuidString(value *uuid.UUID) *string {
	if value == nil {
		return nil
	}
	str := value.String()
	return &str
}
