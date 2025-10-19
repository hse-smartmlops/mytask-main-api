package pg

import (
	"context"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TeamRepository struct {
	db *gorm.DB
}

func NewTeamRepository(db *gorm.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) ListAllTeams(ctx context.Context) ([]models.Team, error) {
	var teams []models.Team
	err := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Where("teams.deleted = FALSE OR teams.deleted IS NULL").
		Preload("TeamMembers", "team_members.deleted = FALSE OR team_members.deleted IS NULL").
		Preload("TeamMembers.User", "users.deleted = FALSE OR users.deleted IS NULL").
		Order("teams.created_at DESC NULLS LAST").
		Find(&teams).Error
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *TeamRepository) ListTeams(ctx context.Context, params ports.PaginationParams) (*ports.Page[models.Team], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Team{}).Where("teams.deleted = FALSE OR teams.deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var teams []models.Team
	offset := (params.Page - 1) * params.PageSize
	if err := base.Order("teams.created_at DESC NULLS LAST").
		Preload("TeamMembers", "team_members.deleted = FALSE OR team_members.deleted IS NULL").
		Preload("TeamMembers.User", "users.deleted = FALSE OR users.deleted IS NULL").
		Limit(params.PageSize).
		Offset(offset).
		Find(&teams).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Team]{
		Items:      teams,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, id uuid.UUID) (*models.Team, error) {
	var team models.Team
	err := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Preload("TeamMembers", "team_members.deleted = FALSE OR team_members.deleted IS NULL").
		Preload("TeamMembers.User", "users.deleted = FALSE OR users.deleted IS NULL").
		Where("teams.id = ? AND (teams.deleted = FALSE OR teams.deleted IS NULL)", id).
		First(&team).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &team, nil
}

func (r *TeamRepository) ListTeamsByProject(ctx context.Context, projectID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team
	err := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Joins("JOIN project_teams ON project_teams.team_id = teams.id").
		Where("(teams.deleted = FALSE OR teams.deleted IS NULL) AND (project_teams.deleted = FALSE OR project_teams.deleted IS NULL) AND project_teams.project_id = ?", projectID).
		Preload("TeamMembers", "team_members.deleted = FALSE OR team_members.deleted IS NULL").
		Preload("TeamMembers.User", "users.deleted = FALSE OR users.deleted IS NULL").
		Order("teams.created_at DESC NULLS LAST").
		Find(&teams).Error
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *TeamRepository) ListTeamsByUser(ctx context.Context, userID uuid.UUID) ([]models.Team, error) {
	var teams []models.Team
	err := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Joins("JOIN team_members ON team_members.team_id = teams.id").
		Where("(teams.deleted = FALSE OR teams.deleted IS NULL) AND (team_members.deleted = FALSE OR team_members.deleted IS NULL) AND team_members.user_id = ?", userID).
		Preload("TeamMembers", "team_members.deleted = FALSE OR team_members.deleted IS NULL").
		Preload("TeamMembers.User", "users.deleted = FALSE OR users.deleted IS NULL").
		Order("teams.created_at DESC NULLS LAST").
		Find(&teams).Error
	if err != nil {
		return nil, err
	}
	return teams, nil
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *models.Team) error {
	return r.db.WithContext(ctx).Create(team).Error
}

func (r *TeamRepository) UpdateTeam(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Team, error) {
	updates["updated_at"] = timePtr(time.Now().UTC())

	result := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetTeamByID(ctx, id)
}

func (r *TeamRepository) SoftDeleteTeam(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.Team{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(map[string]interface{}{
			"deleted":    &deleted,
			"updated_at": &now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TeamRepository) AddUserToTeam(ctx context.Context, input ports.TeamMemberInput) error {
	now := time.Now().UTC()
	deleted := false

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "team_id"}, {Name: "user_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"specialization": input.Specialization, "deleted": &deleted, "updated_at": now}),
	}).Create(&models.TeamMember{
		TeamID:         input.TeamID,
		UserID:         input.UserID,
		Specialization: input.Specialization,
		Deleted:        &deleted,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}).Error
}

func (r *TeamRepository) RemoveUserFromTeam(ctx context.Context, teamID, userID uuid.UUID) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.TeamMember{}).
		Where("team_id = ? AND user_id = ? AND (deleted = FALSE OR deleted IS NULL)", teamID, userID).
		Updates(map[string]interface{}{
			"deleted":    &deleted,
			"updated_at": &now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *TeamRepository) AddTeamToProject(ctx context.Context, input ports.TeamProjectInput) error {
	now := time.Now().UTC()
	deleted := false

	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "project_id"}, {Name: "team_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{"deleted": &deleted, "updated_at": now}),
	}).Create(&models.ProjectTeam{
		ProjectID: input.ProjectID,
		TeamID:    input.TeamID,
		Deleted:   &deleted,
		CreatedAt: &now,
		UpdatedAt: &now,
	}).Error
}

func (r *TeamRepository) RemoveTeamFromProject(ctx context.Context, teamID, projectID uuid.UUID) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.ProjectTeam{}).
		Where("team_id = ? AND project_id = ? AND (deleted = FALSE OR deleted IS NULL)", teamID, projectID).
		Updates(map[string]interface{}{
			"deleted":    &deleted,
			"updated_at": &now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var _ ports.TeamRepository = (*TeamRepository)(nil)
