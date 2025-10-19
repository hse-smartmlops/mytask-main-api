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

type roleService struct {
	repo app.RoleRepository
}

func NewRoleService(repo app.RoleRepository) app.RoleService {
	return &roleService{repo: repo}
}

func (s *roleService) ListRoles(ctx context.Context, params app.PaginationParams) (*app.Page[models.Role], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListRoles(ctx, params)
}

func (s *roleService) GetRole(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	role, err := s.repo.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, domain.ErrNotFound
	}
	return role, nil
}

func (s *roleService) CreateRole(ctx context.Context, input app.CreateRoleInput) (*models.Role, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	deleted := false

	role := &models.Role{
		ID:          uuid.New(),
		Name:        stringPtr(name),
		Description: cloneStringPtr(input.Description),
		Deleted:     &deleted,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	if err := s.repo.CreateRole(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

func (s *roleService) UpdateRole(ctx context.Context, id uuid.UUID, input app.UpdateRoleInput) (*models.Role, error) {
	updates := make(map[string]interface{})

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return nil, domain.ErrInvalidInput
		}
		updates["name"] = name
	}

	if input.Description != nil {
		desc := strings.TrimSpace(*input.Description)
		if desc == "" {
			updates["description"] = nil
		} else {
			updates["description"] = &desc
		}
	}

	if len(updates) == 0 {
		return s.repo.GetRoleByID(ctx, id)
	}

	return s.repo.UpdateRole(ctx, id, updates)
}

func (s *roleService) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteRole(ctx, id)
}

var _ app.RoleService = (*roleService)(nil)
