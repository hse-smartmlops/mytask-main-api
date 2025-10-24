package service

import (
	"context"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type roleService struct {
	repo ports.RoleRepository
	config *config.PaginationConfig
}

func NewRoleService(repo ports.RoleRepository, config *config.PaginationConfig) ports.RoleService {
	return &roleService{repo: repo, config: config}
}

func (s *roleService) ListRoles(ctx context.Context, p ports.PaginationParams) (*ports.Page[models.Role], error) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return s.repo.ListRoles(ctx, p)
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

func (s *roleService) CreateRole(ctx context.Context, input ports.CreateRoleInput) (*models.Role, error) {
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

func (s *roleService) UpdateRole(ctx context.Context, id uuid.UUID, input ports.UpdateRoleInput) (*models.Role, error) {
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

var _ ports.RoleService = (*roleService)(nil)
