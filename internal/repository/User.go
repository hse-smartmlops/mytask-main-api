package repository

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/utils"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository interface {
	GetAllUsers(limit, offset int) ([]models.User, int64, error)
	GetUserById(userId uuid.UUID) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	CreateUser(user models.User) error
	UpdateUser(userId uuid.UUID, updateData map[string]interface{}) (bool, error)
	DeleteUser(userId uuid.UUID) (bool, error)
	BanUser(userId uuid.UUID) (bool, error)
	RestoreUser(req request.RestoreUserRequest) (uuid.UUID, error)
	GetUser(userID string) (*models.User, error)
	GetRole(roleID string) (*models.Role, error)
	CreateUserRole(userRole models.UserRole) error
	RemoveUserRole(userID uuid.UUID, roleID uuid.UUID) (bool, error)
	CreateUserWithID(req request.UserCreateRequest, userID uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) GetAllUsers(limit, offset int) ([]models.User, int64, error) {
	var totalCount int64
	result := r.db.Table("users").Where("users.deleted = FALSE").Count(&totalCount)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	var users []models.User
	if err := r.db.Table("users").
		Where("users.deleted = FALSE").
		Limit(limit).
		Offset(offset).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, totalCount, nil
}

func (r *userRepository) GetUserById(userId uuid.UUID) (*models.User, error) {
	var user models.User
	result := r.db.Session(&gorm.Session{}).Model(models.User{}).
		Where("id = ? AND deleted = FALSE", userId).
		First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

func (r *userRepository) CreateUser(user models.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		log.Printf("CreateUser: db transaction started for id=%s", user.ID)
		
		if err := tx.Omit("UserRoles").Create(&user).Error; err != nil {
			return err
		}

		user.UserRoles = []models.UserRole{}
		log.Printf("CreateUser: user created id=%s", user.ID)
		return nil
	})
}

func (r *userRepository) UpdateUser(userId uuid.UUID, updateData map[string]interface{}) (bool, error) {
	var affected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.User{}).
			Where("id = ? AND deleted = FALSE", userId).
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

func (r *userRepository) DeleteUser(userId uuid.UUID) (bool, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Удаляем пользователя
		result := tx.Session(&gorm.Session{}).Model(&models.User{}).Where("id = ?", userId).Delete(&models.User{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("user not found")
		}

		return nil
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *userRepository) BanUser(userId uuid.UUID) (bool, error) {
	updateData := map[string]interface{}{
		"deleted":    true,
		"updated_at": time.Now(),
	}

	var affected int64
	err := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.User{}).
			Where("id = ? AND deleted = FALSE", userId).
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

func (r *userRepository) RestoreUser(req request.RestoreUserRequest) (uuid.UUID, error) {
	updateData := map[string]interface{}{
		"deleted":    false,
		"updated_at": time.Now(),
	}

	var userId uuid.UUID
	var user models.User
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{}).Model(&models.User{}).
			Where("email = ? AND deleted = TRUE", req.Email).
			First(&user).Error; err != nil {
			return err
		}

		res := tx.Session(&gorm.Session{}).Model(&models.User{}).
			Where("email = ? AND deleted = TRUE", req.Email).
			Updates(updateData)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("user not found")
		}
		userId = user.ID
		return nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	return userId, nil
}

func (r *userRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ? AND deleted = FALSE", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetUser(userID string) (*models.User, error) {
	var user models.User
	if err := r.db.Session(&gorm.Session{}).Model(models.User{}).
		Select("id, profession").
		Where("id = ? AND deleted = FALSE", userID).
		First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetRole(roleID string) (*models.Role, error) {
	var role models.Role
	if err := r.db.Session(&gorm.Session{}).Model(models.Role{}).
		Select("id").
		Where("id = ? AND deleted = FALSE", roleID).
		First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *userRepository) CreateUserRole(userRole models.UserRole) error {
	return r.db.Session(&gorm.Session{}).Omit(clause.Associations).Create(&userRole).Error
}

func (r *userRepository) RemoveUserRole(userID uuid.UUID, roleID uuid.UUID) (bool, error) {
	updateData := map[string]interface{}{
		"deleted":    true,
		"updated_at": time.Now(),
	}

	var affected int64
	result := r.db.Session(&gorm.Session{}).Model(models.UserRole{}).
		Where("user_id = ? AND role_id = ? AND deleted = FALSE", userID, roleID).
		Updates(updateData)
	if result.Error != nil {
		return false, result.Error
	}
	affected = result.RowsAffected

	return affected > 0, nil
}

func (r *userRepository) CreateUserWithID(req request.UserCreateRequest, userID uuid.UUID) error {
	now := time.Now()
	del := false

	user := models.User{
		ID:            userID,
		Email:         utils.GetString(req.Email),
		IsActive:      utils.GetBool(req.IsActive),
		CreatedAt:     now,
		UpdatedAt:     now,
		EmailVerified: utils.GetBool(req.EmailVerified),
		FirstName:     utils.GetString(req.FirstName),
		LastName:      utils.GetString(req.LastName),
		LastLogin:     now,
		Deleted:       del,
		TgID:          "", // Установите значения по умолчанию или из req, если они там есть
		TgUserID:      0,
		Profession:    "",
		UserRoles:     nil,
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).Model(&models.User{}).Omit(clause.Associations).Create(&user); res.Error != nil {
			log.Printf("DB error (create user): %v", res.Error)
			return res.Error
		}
		user.UserRoles = []models.UserRole{} // Инициализируем пустой срез для соответствия ожиданиям
		log.Printf("CreateUserWithId: user created id=%s", user.ID)
		return nil
	})
}