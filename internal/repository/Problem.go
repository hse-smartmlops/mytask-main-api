package repository

import (
	models "emplacc-api/internal/domain"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProblemRepository interface {
	GetAllProblems(limit, offset int) ([]models.Problem, int64, error)
	GetProblemsByUserId(creatorUUID uuid.UUID, limit, offset int) ([]models.Problem, int64, error)
	GetProblemByID(problemId uuid.UUID) (*models.Problem, error)
	CreateProblem(p models.Problem) error
	UpdateProblem(problemId uuid.UUID, updateData map[string]interface{}) (bool, error)
	DeleteProblem(problemId uuid.UUID) (bool, error)
}

type problemRepository struct {
	db *gorm.DB
}

func NewProblemRepository(db *gorm.DB) ProblemRepository {
	return &problemRepository{
		db: db,
	}
}

func (r *problemRepository) GetAllProblems(limit, offset int) ([]models.Problem, int64, error) {
	var totalCount int64
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.Problem{}).
		Where("deleted = FALSE").
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var problems []models.Problem
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.Problem{}).
		Where("deleted = FALSE").
		Limit(limit).
		Offset(offset).
		Find(&problems).Error; err != nil {
		return nil, 0, err
	}

	return problems, totalCount, nil
}

func (r *problemRepository) GetProblemsByUserId(creatorUUID uuid.UUID, limit, offset int) ([]models.Problem, int64, error) {
	var totalCount int64
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.Problem{}).
		Where("creator_id = ? AND deleted = FALSE", creatorUUID).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var problems []models.Problem
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.Problem{}).
		Where("creator_id = ? AND deleted = FALSE", creatorUUID).
		Limit(limit).
		Offset(offset).
		Find(&problems).Error; err != nil {
		return nil, 0, err
	}

	return problems, totalCount, nil
}

func (r *problemRepository) GetProblemByID(problemId uuid.UUID) (*models.Problem, error) {
	var p models.Problem
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.Problem{}).
		Where("id = ? AND deleted = FALSE", problemId).
		First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("problem not found")
		}
		return nil, err
	}
	return &p, nil
}

func (r *problemRepository) CreateProblem(p models.Problem) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.Problem{}).
			Create(&p); res.Error != nil {
			return res.Error
		}
		return nil
	})
}

func (r *problemRepository) UpdateProblem(problemId uuid.UUID, updateData map[string]interface{}) (bool, error) {
	var affected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.Problem{}).
			Where("id = ? AND deleted = FALSE", problemId).
			Updates(updateData)
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

func (r *problemRepository) DeleteProblem(problemId uuid.UUID) (bool, error) {
	delTrue := true
	now := time.Now()
	update := map[string]interface{}{
		"deleted":    &delTrue,
		"updated_at": &now,
	}

	var affected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// помечаем проблему удалённой
		res := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.Problem{}).
			Where("id = ?", problemId).
			Updates(update)
		if res.Error != nil {
			return res.Error
		}
		affected = res.RowsAffected
		if affected == 0 {
			return errors.New("problem not found")
		}

		// каскадно помечаем связанные сущности
		if res := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.ForumMessage{}).
			Where("problem_id = ?", problemId).
			Updates(update); res.Error != nil {
			return res.Error
		}

		if res := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.ReportProblem{}).
			Where("problem_id = ?", problemId).
			Updates(update); res.Error != nil {
			return res.Error
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return affected > 0, nil
}