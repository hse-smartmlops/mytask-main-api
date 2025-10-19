package pg

import (
	"context"
	"time"

	"emplacc-api/internal/app"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) ListTasks(ctx context.Context, params app.PaginationParams) (*app.Page[models.Task], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Task{}).Where("tasks.deleted = FALSE OR tasks.deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var tasks []models.Task
	offset := (params.Page - 1) * params.PageSize
	if err := base.Order("tasks.created_at DESC NULLS LAST").
		Preload("Status", "deleted = FALSE OR deleted IS NULL").
		Preload("Status.Board").
		Preload("CreatedByUser").
		Preload("AssignedToUser").
		Limit(params.PageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &app.Page[models.Task]{
		Items:      tasks,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *TaskRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Preload("Status").
		Preload("Status.Board").
		Preload("CreatedByUser").
		Preload("AssignedToUser").
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		First(&task).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (r *TaskRepository) ListTasksByProject(ctx context.Context, projectID uuid.UUID, params app.PaginationParams) (*app.Page[models.Task], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Task{}).
		Joins("JOIN statuses ON statuses.id = tasks.status_id").
		Joins("JOIN boards ON boards.id = statuses.board_id").
		Where("(tasks.deleted = FALSE OR tasks.deleted IS NULL) AND (statuses.deleted = FALSE OR statuses.deleted IS NULL) AND boards.project_id = ?", projectID)
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var tasks []models.Task
	offset := (params.Page - 1) * params.PageSize
	if err := base.Order("tasks.created_at DESC NULLS LAST").
		Preload("Status").
		Preload("Status.Board").
		Preload("CreatedByUser").
		Preload("AssignedToUser").
		Limit(params.PageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &app.Page[models.Task]{
		Items:      tasks,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *TaskRepository) ListTasksByUser(ctx context.Context, userID uuid.UUID, params app.PaginationParams) (*app.Page[models.Task], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Task{}).
		Where("tasks.assigned_to = ? AND (tasks.deleted = FALSE OR tasks.deleted IS NULL)", userID)
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var tasks []models.Task
	offset := (params.Page - 1) * params.PageSize
	if err := base.Order("tasks.created_at DESC NULLS LAST").
		Preload("Status").
		Preload("Status.Board").
		Preload("CreatedByUser").
		Preload("AssignedToUser").
		Limit(params.PageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &app.Page[models.Task]{
		Items:      tasks,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *TaskRepository) CreateTask(ctx context.Context, task *models.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *TaskRepository) UpdateTask(ctx context.Context, id uuid.UUID, updates map[string]interface{}) (*models.Task, error) {
	updates["updated_at"] = timePtr(time.Now().UTC())

	result := r.db.WithContext(ctx).
		Model(&models.Task{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetTaskByID(ctx, id)
}

func (r *TaskRepository) SoftDeleteTask(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.Task{}).
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

var _ app.TaskRepository = (*TaskRepository)(nil)
