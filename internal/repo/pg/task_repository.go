package pg

import (
	"context"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(
	db *gorm.DB,
) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (r *TaskRepository) ListActiveTasksByUser(
    ctx context.Context,
    userID uuid.UUID,
    p ports.PaginationParams,
) (*ports.Page[models.Task], error) {
    var totalCount int64

    base := r.db.WithContext(ctx).
        Model(&models.Task{}).
        Preload("Status", "deleted = FALSE OR deleted IS NULL").
        Preload("Status.Board").
        Preload("CreatedByUser").
        Preload("AssignedToUser").
        Where(`
            tasks.assigned_to = ? 
            AND (tasks.deleted = FALSE OR tasks.deleted IS NULL)
            AND tasks.status_id NOT IN (
                SELECT id FROM statuses 
                WHERE name = 'Done' 
                AND (deleted = FALSE OR deleted IS NULL)
            )
        `, userID)

    if err := base.Count(&totalCount).Error; err != nil {
        return nil, err
    }

    var tasks []models.Task
    offset := (p.Page - 1) * p.PageSize
    if err := base.Order("tasks.created_at DESC NULLS LAST").
        Limit(p.PageSize).
        Offset(offset).
        Find(&tasks).Error; err != nil {
        return nil, err
    }

    return &ports.Page[models.Task]{
        Items:      tasks,
        Page:       p.Page,
        PageSize:   p.PageSize,
        TotalCount: totalCount,
    }, nil
}

func (r *TaskRepository) ListTasksByBoard(
    ctx context.Context,
    boardID uuid.UUID,
    p ports.PaginationParams,
) (*ports.Page[models.Task], error) {
    var totalCount int64

    base := r.db.WithContext(ctx).
        Model(&models.Task{}).
        Joins("JOIN statuses ON statuses.id = tasks.status_id").
        Preload("Status").
        Preload("Status.Board").
        Preload("CreatedByUser").
        Preload("AssignedToUser").
        Where(`
            (tasks.deleted = FALSE OR tasks.deleted IS NULL) 
            AND (statuses.deleted = FALSE OR statuses.deleted IS NULL) 
            AND statuses.board_id = ?
        `, boardID)

    if err := base.Count(&totalCount).Error; err != nil {
        return nil, err
    }

    var tasks []models.Task
    offset := (p.Page - 1) * p.PageSize
    if err := base.Order("tasks.created_at DESC NULLS LAST").
        Limit(p.PageSize).
        Offset(offset).
        Find(&tasks).Error; err != nil {
        return nil, err
    }

    return &ports.Page[models.Task]{
        Items:      tasks,
        Page:       p.Page,
        PageSize:   p.PageSize,
        TotalCount: totalCount,
    }, nil
}

func (r *TaskRepository) ListTasksByUserAndProject(
    ctx context.Context,
    userID, projectID uuid.UUID,
    p ports.PaginationParams,
) (*ports.Page[models.Task], error) {
    var totalCount int64

    base := r.db.WithContext(ctx).
        Model(&models.Task{}).
        Joins("JOIN statuses ON statuses.id = tasks.status_id").
        Joins("JOIN boards ON boards.id = statuses.board_id").
        Preload("Status").
        Preload("Status.Board").
        Preload("CreatedByUser").
        Preload("AssignedToUser").
        Where(`
            tasks.assigned_to = ? 
            AND (tasks.deleted = FALSE OR tasks.deleted IS NULL)
            AND (statuses.deleted = FALSE OR statuses.deleted IS NULL)
            AND (boards.deleted = FALSE OR boards.deleted IS NULL)
            AND boards.project_id = ?
        `, userID, projectID)

    if err := base.Count(&totalCount).Error; err != nil {
        return nil, err
    }

    var tasks []models.Task
    offset := (p.Page - 1) * p.PageSize
    if err := base.Order("tasks.created_at DESC NULLS LAST").
        Limit(p.PageSize).
        Offset(offset).
        Find(&tasks).Error; err != nil {
        return nil, err
    }

    return &ports.Page[models.Task]{
        Items:      tasks,
        Page:       p.Page,
        PageSize:   p.PageSize,
        TotalCount: totalCount,
    }, nil
}

func (r *TaskRepository) MoveTaskBetweenStatuses(
    ctx context.Context,
    taskID uuid.UUID,
    newStatusID uuid.UUID,
) (*models.Task, error) {
    updates := map[string]interface{}{
        "status_id":  newStatusID,
        "updated_at": timePtr(time.Now().UTC()),
    }

    result := r.db.WithContext(ctx).
        Model(&models.Task{}).
        Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", taskID).
        Updates(updates)
    if result.Error != nil {
        return nil, result.Error
    }
    if result.RowsAffected == 0 {
        return nil, domain.ErrNotFound
    }

    return r.GetTaskByID(ctx, taskID)
}

func (r *TaskRepository) ListTasks(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.Task], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Task{}).
		Preload("Status", "deleted = FALSE OR deleted IS NULL").
		Preload("Status.Board").
		Preload("CreatedByUser").
		Preload("AssignedToUser").
		Where("tasks.deleted = FALSE OR tasks.deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var tasks []models.Task
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("tasks.created_at DESC NULLS LAST").
		Limit(p.PageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Task]{
		Items:      tasks,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *TaskRepository) GetTaskByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Task, error) {
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

func (r *TaskRepository) ListTasksByProject(
	ctx context.Context,
	projectID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.Task], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Task{}).
		Joins("JOIN statuses ON statuses.id = tasks.status_id").
		Joins("JOIN boards ON boards.id = statuses.board_id").
		Preload("Status").
		Preload("Status.Board").
		Preload("CreatedByUser").
		Preload("AssignedToUser").
		Where("(tasks.deleted = FALSE OR tasks.deleted IS NULL) AND (statuses.deleted = FALSE OR statuses.deleted IS NULL) AND boards.project_id = ?", projectID)
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var tasks []models.Task
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("tasks.created_at DESC NULLS LAST").
		Limit(p.PageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Task]{
		Items:      tasks,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *TaskRepository) ListTasksByUser(
	ctx context.Context,
	userID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.Task], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Task{}).
		Preload("Status").
		Preload("Status.Board").
		Preload("CreatedByUser").
		Preload("AssignedToUser").
		Where("tasks.assigned_to = ? AND (tasks.deleted = FALSE OR tasks.deleted IS NULL)", userID)
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var tasks []models.Task
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("tasks.created_at DESC NULLS LAST").
		Limit(p.PageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Task]{
		Items:      tasks,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *TaskRepository) CreateTask(
	ctx context.Context,
	task *models.Task,
) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *TaskRepository) UpdateTask(
	ctx context.Context,
	id uuid.UUID,
	updates map[string]interface{},
) (*models.Task, error) {
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

func (r *TaskRepository) SoftDeleteTask(
	ctx context.Context,
	id uuid.UUID,
) error {
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

func (r *TaskRepository) GetTaskBoardAndProjectIDs(
    ctx context.Context,
    taskID uuid.UUID,
) (boardID uuid.UUID, projectID uuid.UUID, err error) {
    var task models.Task
    err = r.db.WithContext(ctx).
        Model(&models.Task{}).
        Select("tasks.status_id").
        Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", taskID).
        First(&task).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return uuid.Nil, uuid.Nil, domain.ErrNotFound
        }
        return uuid.Nil, uuid.Nil, err
    }

    var status models.Status
    err = r.db.WithContext(ctx).
        Model(&models.Status{}).
        Select("board_id").
        Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", task.StatusID).
        First(&status).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return uuid.Nil, uuid.Nil, domain.ErrNotFound
        }
        return uuid.Nil, uuid.Nil, err
    }

    var board models.Board
    err = r.db.WithContext(ctx).
        Model(&models.Board{}).
        Select("project_id").
        Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", status.BoardID).
        First(&board).Error
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            return uuid.Nil, uuid.Nil, domain.ErrNotFound
        }
        return uuid.Nil, uuid.Nil, err
    }

    return status.BoardID, board.ProjectID, nil
}

func (r *TaskRepository) StatusExists(
    ctx context.Context,
    statusID uuid.UUID,
) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&models.Status{}).
        Joins("JOIN boards ON statuses.board_id = boards.id").
        Where(`
            statuses.id = ? 
            AND (statuses.deleted = FALSE OR statuses.deleted IS NULL)
            AND (boards.deleted = FALSE OR boards.deleted IS NULL)
        `, statusID).
        Count(&count).Error
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

func (r *TaskRepository) UserExists(
    ctx context.Context,
    userID uuid.UUID,
) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).
        Model(&models.User{}).
        Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", userID).
        Count(&count).Error
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

var _ ports.TaskRepository = (*TaskRepository)(nil)
