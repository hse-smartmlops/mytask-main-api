package service

import (
	"context"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type projectService struct {
	projectRepo ports.ProjectRepository
	teamRepo    ports.TeamRepository
}

func NewProjectService(projectRepo ports.ProjectRepository, teamRepo ports.TeamRepository) ports.ProjectService {
	return &projectService{
		projectRepo: projectRepo,
		teamRepo:    teamRepo,
	}
}

func (s *projectService) GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	project, err := s.projectRepo.GetProjectByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, domain.ErrNotFound
	}
	return project, nil
}

func (s *projectService) CreateProject(ctx context.Context, input ports.CreateProjectInput) (*models.Project, error) {
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

	if err := s.projectRepo.CreateProject(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectService) UpdateProject(ctx context.Context, id uuid.UUID, input ports.UpdateProjectInput) (*models.Project, error) {
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
		return s.projectRepo.GetProjectByID(ctx, id)
	}

	return s.projectRepo.UpdateProject(ctx, id, updates)
}

func (s *projectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	return s.projectRepo.SoftDeleteProject(ctx, id)
}

func (s *projectService) ListProjects(ctx context.Context, p ports.PaginationParams) (*ports.Page[models.Project], error) {
	return s.projectRepo.ListProjects(ctx, p)
}

func (s *projectService) ListProjectsByUser(ctx context.Context, userID uuid.UUID, p ports.PaginationParams) (*ports.Page[models.Project], error) {
	return s.projectRepo.ListProjectsByUser(ctx, userID, p)
}

func (s *projectService) ListTeamsByProject(ctx context.Context, projectID uuid.UUID, p ports.PaginationParams) (*ports.Page[models.Team], error) {
	return s.teamRepo.ListTeamsByProject(ctx, projectID, p)
}

func (s *projectService) ListProjectsByTeam(ctx context.Context, teamID uuid.UUID, p ports.PaginationParams) (*ports.Page[models.Project], error) {
	return s.projectRepo.ListProjectsByTeam(ctx, teamID, p)
}

var _ ports.ProjectService = (*projectService)(nil)
