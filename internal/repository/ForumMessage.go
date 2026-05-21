package repository

import (
	models "emplacc-api/internal/domain"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ForumMessageRepository interface {
	GetAllForumMessages(limit, offset int) ([]models.ForumMessage, int64, error)
	GetForumMessagesByProblemId(problemID uuid.UUID, limit, offset int) ([]models.ForumMessage, int64, error)
	GetForumMessageById(messageID uuid.UUID) (*models.ForumMessage, error)
	CreateForumMessage(fm models.ForumMessage) error
	UpdateForumMessage(messageID uuid.UUID, updateData map[string]interface{}) (bool, error)
	DeleteForumMessage(messageID uuid.UUID) (bool, error)
}

type forumMessageRepository struct {
	db *gorm.DB
}

func NewForumMessageRepository(db *gorm.DB) ForumMessageRepository {
	return &forumMessageRepository{
		db: db,
	}
}

func (r *forumMessageRepository) GetAllForumMessages(limit, offset int) ([]models.ForumMessage, int64, error) {
	var totalCount int64
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.ForumMessage{}).
		Where("deleted = FALSE").
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var forumMessages []models.ForumMessage
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.ForumMessage{}).
		Where("deleted = FALSE").
		Limit(limit).
		Offset(offset).
		Find(&forumMessages).Error; err != nil {
		return nil, 0, err
	}

	return forumMessages, totalCount, nil
}

func (r *forumMessageRepository) GetForumMessagesByProblemId(problemID uuid.UUID, limit, offset int) ([]models.ForumMessage, int64, error) {
	var totalCount int64
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.ForumMessage{}).
		Where("deleted = FALSE AND problem_id = ?", problemID).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var forumMessages []models.ForumMessage
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Where("forum_messages.deleted = FALSE AND forum_messages.problem_id = ?", problemID).
		Limit(limit).
		Offset(offset).
		Order("forum_messages.created_at ASC").
		Preload("User").
		Preload("ReplyTo", func(db *gorm.DB) *gorm.DB {
			return db.Where("deleted = ?", false)
		}).
		Preload("ReplyTo.User").
		Find(&forumMessages).Error; err != nil {
		return nil, 0, err
	}

	return forumMessages, totalCount, nil
}

func (r *forumMessageRepository) GetForumMessageById(messageID uuid.UUID) (*models.ForumMessage, error) {
	var m models.ForumMessage
	if err := r.db.Session(&gorm.Session{NewDB: true}).
		Model(&models.ForumMessage{}).
		Where("id = ? AND deleted = FALSE", messageID).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("forum message not found")
		}
		return nil, err
	}
	return &m, nil
}

func (r *forumMessageRepository) CreateForumMessage(fm models.ForumMessage) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.ForumMessage{}).
			Create(&fm); res.Error != nil {
			return res.Error
		}
		return nil
	})
}

func (r *forumMessageRepository) UpdateForumMessage(messageID uuid.UUID, updateData map[string]interface{}) (bool, error) {
	var affected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.ForumMessage{}).
			Where("id = ? AND deleted = FALSE", messageID).
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

func (r *forumMessageRepository) DeleteForumMessage(messageID uuid.UUID) (bool, error) {
	delTrue := true
	now := time.Now()
	update := map[string]interface{}{
		"deleted":    &delTrue,
		"updated_at": &now,
	}

	var affected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{NewDB: true}).
			Model(&models.ForumMessage{}).
			Where("id = ?", messageID).
			Updates(update)
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