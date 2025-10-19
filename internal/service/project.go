package service

import (
	"context"
	"strings"
	"time"

	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type projectService struct {
	repo app.ProjectRepository
}

func NewProjectService(repo app.ProjectRepository) app.ProjectService {
	return &projectService{repo: repo}
}

func (s *projectService) ListProjects(ctx context.Context, params app.PaginationParams) (*app.Page[models.Project], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListProjects(ctx, params)
}

func (s *projectService) GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	project, err := s.repo.GetProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, domain.ErrNotFound
	}
	return project, nil
}

func (s *projectService) CreateProject(ctx context.Context, input app.CreateProjectInput) (*models.Project, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	deleted := false

	project := &models.Project{
		ID:              uuid.New(),
		Name:            stringPtr(name),
		Description:     cloneStringPtr(input.Description),
		GitlabProjectID: input.GitlabProjectID,
		GitlabURL:       cloneStringPtr(input.GitlabURL),
		CreatedBy:       input.CreatedBy,
		Status:          cloneStringPtr(input.Status),
		CreatedAt:       &now,
		UpdatedAt:       &now,
		Deleted:         &deleted,
	}

	if err := s.repo.CreateProject(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) UpdateProject(ctx context.Context, id uuid.UUID, input app.UpdateProjectInput) (*models.Project, error) {
	updates := make(map[string]interface{})

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.ErrInvalidInput
		}
		updates["name"] = &name
	}

	if input.Description != nil {
		updates["description"] = cloneStringPtr(input.Description)
	}

	if input.GitlabProjectID != nil {
		updates["gitlab_project_id"] = input.GitlabProjectID
	}

	if input.GitlabURL != nil {
		updates["gitlab_url"] = cloneStringPtr(input.GitlabURL)
	}

	if input.Status != nil {
		updates["status"] = cloneStringPtr(input.Status)
	}

	if len(updates) == 0 {
		return s.repo.GetProjectByID(ctx, id)
	}

	return s.repo.UpdateProject(ctx, id, updates)
}

func (s *projectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteProject(ctx, id)
}

var _ app.ProjectService = (*projectService)(nil)
