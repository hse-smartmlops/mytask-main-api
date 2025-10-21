package repository

import (
	models "emplacc-api/internal/domain"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TaskRepository interface {
	GetAllTasks(limit, offset int) ([]models.Task, int64, error)
	GetTaskByID(taskID uuid.UUID) (*models.Task, error)
	GetTasksByProjectID(projectID uuid.UUID, limit, offset int) ([]models.Task, int64, error)
	CreateTask(task models.Task) error
	UpdateTask(taskID uuid.UUID, updates map[string]interface{}) (bool, error)
	DeleteTask(taskID uuid.UUID) (bool, error)
	GetTasksByUserId(userID uuid.UUID, limit, offset int) ([]models.Task, int64, error) 
	GetStatusByID(statusID uuid.UUID) (*models.Status, error)
	GetStatusesByBoardID(boardID uuid.UUID) ([]models.Status, error)
	UpdateTaskStatus(taskID uuid.UUID, toStatusID uuid.UUID, updatedAt time.Time) error
	UserExists(userID uuid.UUID) (bool, error)
	StatusExists(statusID uuid.UUID) (bool, error)
	GetTasksByUserIDAndProjectID(userID, projectID uuid.UUID, limit, offset int) ([]models.Task, int64, error)
	GetActiveTasksByUserId(userID uuid.UUID, limit, offset int) ([]models.Task, int64, error)
	GetTaskBoardAndProjectIDs(taskID uuid.UUID) (boardId, projectId uuid.UUID, err error)
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{
		db: db,
	}
}

func (r *taskRepository) GetAllTasks(limit, offset int) ([]models.Task, int64, error) {
	var totalCount int64
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.Task{}).
		Where("deleted = ?", false).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var tasks []models.Task
	if err := r.db.Session(&gorm.Session{}).
		Where("deleted = ?", false).
		Limit(limit).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, totalCount, nil
}

func (r *taskRepository) GetTaskByID(taskID uuid.UUID) (*models.Task, error) {
	var task models.Task
	if err := r.db.Session(&gorm.Session{}).
		Where("id = ? AND deleted = ?", taskID, false).
		Preload("Status", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		Preload("AssignedToUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		First(&task).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("task not found")
		}
		return nil, err
	}

	return &task, nil
}

func (r *taskRepository) GetTaskBoardAndProjectIDs(taskID uuid.UUID) (boardID uuid.UUID, projectID uuid.UUID, err error) {
    // 1. Получаем задачу
    var task models.Task
    if err := r.db.Session(&gorm.Session{}).
        Where("id = ? AND deleted = ?", taskID, false).
        First(&task).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return uuid.Nil, uuid.Nil, errors.New("task not found")
        }
        return uuid.Nil, uuid.Nil, err
    }

    // 2. Получаем статус (только board_id)
    var status models.Status
    if err := r.db.Session(&gorm.Session{}).
        Select("board_id").
        Where("id = ? AND deleted = ?", task.StatusID, false).
        First(&status).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return uuid.Nil, uuid.Nil, errors.New("status not found")
        }
        return uuid.Nil, uuid.Nil, err
    }

    // 3. Получаем доску (только project_id)
    var board models.Board
    if err := r.db.Session(&gorm.Session{}).
        Select("project_id").
        Where("id = ? AND deleted = ?", status.BoardID, false).
        First(&board).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            return uuid.Nil, uuid.Nil, errors.New("board not found")
        }
        return uuid.Nil, uuid.Nil, err
    }

    return status.BoardID, board.ProjectID, nil
}

func (r *taskRepository) GetTasksByProjectID(projectID uuid.UUID, limit, offset int) ([]models.Task, int64, error) {
	// Подзапрос: все status_id, принадлежащие доскам этого проекта
	var statusIDs []uuid.UUID
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Select("statuses.id").
		Joins("JOIN boards ON statuses.board_id = boards.id").
		Where("boards.project_id = ? AND boards.deleted = ? AND statuses.deleted = ?", projectID, false, false).
		Scan(&statusIDs).Error; err != nil {
		return nil, 0, err
	}

	var totalCount int64
	if len(statusIDs) == 0 {
		totalCount = 0
	} else {
		if err := r.db.Session(&gorm.Session{}).
			Model(&models.Task{}).
			Where("status_id IN ? AND deleted = ?", statusIDs, false).
			Count(&totalCount).Error; err != nil {
			return nil, 0, err
		}
	}

	var tasks []models.Task
	if len(statusIDs) > 0 {
		if err := r.db.Session(&gorm.Session{}).
			Where("status_id IN ? AND deleted = ?", statusIDs, false).
			Limit(limit).
			Offset(offset).
			Find(&tasks).Error; err != nil {
			return nil, 0, err
		}
	}

	return tasks, totalCount, nil
}

func (r *taskRepository) CreateTask(task models.Task) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: false}).
		Omit(clause.Associations).
		Create(&task).Error
}

func (r *taskRepository) UpdateTask(taskID uuid.UUID, updates map[string]interface{}) (bool, error) {
	var affected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.Task{}).Where("id = ? AND deleted = FALSE", taskID).Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		affected = res.RowsAffected
		return nil
	})

	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (r *taskRepository) DeleteTask(taskID uuid.UUID) (bool, error) {
	updateData := map[string]interface{}{"deleted": true}

	var affected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.Task{}).Where("id = ?", taskID).Updates(updateData)
		if res.Error != nil {
			return res.Error
		}
		affected = res.RowsAffected
		return nil
	})

	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

func (r *taskRepository) GetStatusByID(statusID uuid.UUID) (*models.Status, error) {
	var status models.Status
	if err := r.db.Session(&gorm.Session{}).Model(&models.Status{}).
		Select("board_id").
		Where("id = ? AND deleted = ?", statusID, false).
		First(&status).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("status not found")
		}
		return nil, err
	}

	return &status, nil
}

func (r *taskRepository) GetStatusesByBoardID(boardID uuid.UUID) ([]models.Status, error) {
	var statuses []models.Status
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Where("board_id = ? AND deleted = ?", boardID, false).
		Order("sort_order ASC").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		Find(&statuses).Error; err != nil {
		return nil, err
	}

	return statuses, nil
}

func (r *taskRepository) UpdateTaskStatus(taskID uuid.UUID, toStatusID uuid.UUID, updatedAt time.Time) error {
	return r.db.Session(&gorm.Session{}).Model(&models.Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status_id":   toStatusID,
			"updated_at":  updatedAt,
		}).Error
}

func (r *taskRepository) UserExists(userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.Session(&gorm.Session{}).Model(&models.User{}).
		Where("id = ? AND deleted = ?", userID, false).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *taskRepository) StatusExists(statusID uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Session(&gorm.Session{}).Model(&models.Status{}).
		Joins("INNER JOIN boards ON statuses.board_id = boards.id").
		Where("statuses.id = ? AND statuses.deleted = ? AND boards.deleted = ?", statusID, false, false).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *taskRepository) GetTasksByUserIDAndProjectID(userID, projectID uuid.UUID, limit, offset int) ([]models.Task, int64, error) {
	var totalCount int64

	if err := r.db.Session(&gorm.Session{}).
		Model(&models.Task{}).
		Joins("JOIN statuses ON tasks.status_id = statuses.id").
		Joins("JOIN boards ON statuses.board_id = boards.id").
		Where("tasks.assigned_to = ? AND tasks.deleted = ?", userID, false).
		Where("boards.project_id = ? AND boards.deleted = ?", projectID, false).
		Where("statuses.deleted = ?", false).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []models.Task{}, 0, nil
	}

	var tasks []models.Task
	if err := r.db.Session(&gorm.Session{}).
		Where("tasks.assigned_to = ? AND tasks.deleted = ?", userID, false).
		Joins("JOIN statuses ON tasks.status_id = statuses.id").
		Joins("JOIN boards ON statuses.board_id = boards.id").
		Where("boards.project_id = ? AND boards.deleted = ?", projectID, false).
		Where("statuses.deleted = ?", false).
		Limit(limit).
		Offset(offset).
		Preload("Status", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).
				Where("deleted = ?", false). // ← ИСПРАВЛЕНО
				Preload("Board", func(d *gorm.DB) *gorm.DB {
					return d.Session(&gorm.Session{}).
						Select("id, name, project_id").
						Where("deleted = ?", false)
				})
		}).
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).
				Select("id, first_name, last_name, email").
				Where("deleted = ?", false)
		}).
		Preload("AssignedToUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).
				Select("id, first_name, last_name, email").
				Where("deleted = ?", false)
		}).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, totalCount, nil
}

func (r *taskRepository) GetActiveTasksByUserId(userID uuid.UUID, limit, offset int) ([]models.Task, int64, error) {
	var totalCount int64
	
	// Для подсчета используем подзапрос
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.Task{}).
		Where("assigned_to = ? AND deleted = ?", userID, false).
		Where("status_id NOT IN (SELECT id FROM statuses WHERE name = ? AND deleted = ?)", "Done", false).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []models.Task{}, 0, nil
	}

	var tasks []models.Task
	// В основном запросе используем Preload с условием
	if err := r.db.Session(&gorm.Session{}).
		Where("assigned_to = ? AND deleted = ?", userID, false).
		Where("status_id NOT IN (SELECT id FROM statuses WHERE name = ? AND deleted = ?)", "Done", false).
		Limit(limit).
		Offset(offset).
		Preload("Status", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).
				Where("deleted = ? AND name != ?", false, "Done")
		}).
		Preload("Status.Board", func(d *gorm.DB) *gorm.DB {
			return d.Session(&gorm.Session{}).
				Select("id, name, project_id").
				Where("deleted = ?", false)
		}).
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).
				Select("id, first_name, last_name, email").
				Where("deleted = ?", false)
		}).
		Preload("AssignedToUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).
				Select("id, first_name, last_name, email").
				Where("deleted = ?", false)
		}).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, totalCount, nil
}

func (r *taskRepository) GetTasksByUserId(userID uuid.UUID, limit, offset int) ([]models.Task, int64, error) {
	var totalCount int64
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.Task{}).
		Where("assigned_to = ? AND deleted = ?", userID, false).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if totalCount == 0 {
		return []models.Task{}, 0, nil
	}

	var tasks []models.Task
	if err := r.db.Session(&gorm.Session{}).
		Where("assigned_to = ? AND deleted = ?", userID, false).
		Limit(limit).
		Offset(offset).
		Preload("Status", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).
				Where("deleted = ?", false).
				Preload("Board", func(d *gorm.DB) *gorm.DB {
					return d.Session(&gorm.Session{}).
						Select("id, name, project_id").
						Where("deleted = ?", false)
				})
		}).
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).
				Select("id, first_name, last_name, email").
				Where("deleted = ?", false)
		}).
		Preload("AssignedToUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).
				Select("id, first_name, last_name, email").
				Where("deleted = ?", false)
		}).
		Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, totalCount, nil
}