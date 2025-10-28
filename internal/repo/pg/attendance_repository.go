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

type AttendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(
	db *gorm.DB,
) *AttendanceRepository {
	return &AttendanceRepository{
		db: db,
	}
}

func (r *AttendanceRepository) ListAttendances(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.Attendance], error) {
	var totalCount int64

	base := r.db.WithContext(ctx).Model(&models.Attendance{}).
		Where("attendances.deleted = FALSE OR attendances.deleted IS NULL")
	if err := base.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	var attendances []models.Attendance
	offset := (p.Page - 1) * p.PageSize
	if err := base.Order("attendances.date DESC NULLS LAST, attendances.created_at DESC NULLS LAST").
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Limit(p.PageSize).
		Offset(offset).
		Find(&attendances).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.Attendance]{
		Items:      attendances,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: totalCount,
	}, nil
}

func (r *AttendanceRepository) ListAttendancesByUser(
	ctx context.Context,
	userID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.Attendance], error) {
	var attendances []models.Attendance
	err := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Where("attendances.user_id = ? AND (attendances.deleted = FALSE OR attendances.deleted IS NULL)", userID).
		Order("attendances.date DESC NULLS LAST, attendances.created_at DESC NULLS LAST").
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Find(&attendances).Error
	if err != nil {
		return nil, err
	}
	return &ports.Page[models.Attendance]{ // TODO Add pagination on db layer
		Items:      attendances,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: 0,
	}, nil
}

func (r *AttendanceRepository) GetAttendanceByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Attendance, error) {
	var attendance models.Attendance
	err := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Where("attendances.id = ? AND (attendances.deleted = FALSE OR attendances.deleted IS NULL)", id).
		First(&attendance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &attendance, nil
}

func (r *AttendanceRepository) CreateAttendance(
	ctx context.Context,
	attendance *models.Attendance,
) error {
	return r.db.WithContext(ctx).Create(attendance).Error
}

func (r *AttendanceRepository) UpdateAttendance(
	ctx context.Context,
	id uuid.UUID,
	updates map[string]interface{},
) (*models.Attendance, error) {
	updates["updated_at"] = timePtr(time.Now().UTC())

	result := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.GetAttendanceByID(ctx, id)
}

func (r *AttendanceRepository) SoftDeleteAttendance(
	ctx context.Context,
	id uuid.UUID,
) error {
	now := time.Now().UTC()
	deleted := true
	result := r.db.WithContext(ctx).
		Model(&models.Attendance{}).
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

var _ ports.AttendanceRepository = (*AttendanceRepository)(nil)
