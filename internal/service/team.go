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

type teamService struct {
	repo app.TeamRepository
}

func NewTeamService(repo app.TeamRepository) app.TeamService {
	return &teamService{repo: repo}
}

func (s *teamService) ListTeams(ctx context.Context, params app.PaginationParams) (*app.Page[models.Team], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListTeams(ctx, params)
}

func (s *teamService) GetTeam(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	team, err := s.repo.GetTeamByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if team == nil {
		return nil, domain.ErrNotFound
	}
	return team, nil
}

func (s *teamService) ListTeamsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Team, error) {
	return s.repo.ListTeamsByProject(ctx, projectID)
}

func (s *teamService) CreateTeam(ctx context.Context, input app.CreateTeamInput) (*models.Team, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	deleted := false

	team := &models.Team{
		ID:          uuid.New(),
		Name:        stringPtr(name),
		Description: cloneStringPtr(input.Description),
		CreatedAt:   &now,
		UpdatedAt:   &now,
		Deleted:     &deleted,
	}

	if err := s.repo.CreateTeam(ctx, team); err != nil {
		return nil, err
	}

	return s.repo.GetTeamByID(ctx, team.ID)
}

func (s *teamService) UpdateTeam(ctx context.Context, id uuid.UUID, input app.UpdateTeamInput) (*models.Team, error) {
	updates := make(map[string]interface{})

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.ErrInvalidInput
		}
		updates["name"] = name
	}

	if input.Description != nil {
		updates["description"] = cloneStringPtr(input.Description)
	}

	if len(updates) == 0 {
		return s.repo.GetTeamByID(ctx, id)
	}

	return s.repo.UpdateTeam(ctx, id, updates)
}

func (s *teamService) DeleteTeam(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteTeam(ctx, id)
}

func (s *teamService) AddUserToTeam(ctx context.Context, input app.TeamMemberInput) error {
	if input.TeamID == uuid.Nil || input.UserID == uuid.Nil {
		return domain.ErrInvalidInput
	}
	if input.Specialization != nil {
		specialization := strings.TrimSpace(*input.Specialization)
		if specialization == "" {
			input.Specialization = nil
		} else {
			input.Specialization = &specialization
		}
	}
	return s.repo.AddUserToTeam(ctx, input)
}

func (s *teamService) RemoveUserFromTeam(ctx context.Context, teamID, userID uuid.UUID) error {
	if teamID == uuid.Nil || userID == uuid.Nil {
		return domain.ErrInvalidInput
	}
	return s.repo.RemoveUserFromTeam(ctx, teamID, userID)
}

func (s *teamService) AddTeamToProject(ctx context.Context, input app.TeamProjectInput) error {
	if input.TeamID == uuid.Nil || input.ProjectID == uuid.Nil {
		return domain.ErrInvalidInput
	}
	return s.repo.AddTeamToProject(ctx, input)
}

func (s *teamService) RemoveTeamFromProject(ctx context.Context, teamID, projectID uuid.UUID) error {
	if teamID == uuid.Nil || projectID == uuid.Nil {
		return domain.ErrInvalidInput
	}
	return s.repo.RemoveTeamFromProject(ctx, teamID, projectID)
}

var _ app.TeamService = (*teamService)(nil)
