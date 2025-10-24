package service

import (
	"context"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type attendanceService struct {
	repo ports.AttendanceRepository
	config *config.PaginationConfig
}

func NewAttendanceService(repo ports.AttendanceRepository, config *config.PaginationConfig) ports.AttendanceService {
	return &attendanceService{repo: repo,config: config}
}

func (s *attendanceService) ListAttendances(ctx context.Context, params ports.PaginationParams) (*ports.Page[models.Attendance], error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListAttendances(ctx, params)
}

func (s *attendanceService) ListAttendancesByUser(ctx context.Context, userID uuid.UUID, params ports.PaginationParams) ([]models.Attendance, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}
	return s.repo.ListAttendancesByUser(ctx, userID, params)
}

func (s *attendanceService) GetAttendance(ctx context.Context, id uuid.UUID) (*models.Attendance, error) {
	attendance, err := s.repo.GetAttendanceByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if attendance == nil {
		return nil, domain.ErrNotFound
	}
	return attendance, nil
}

func (s *attendanceService) CreateAttendance(ctx context.Context, input ports.CreateAttendanceInput) (*models.Attendance, error) {
	if input.UserID == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}

	now := time.Now().UTC()
	deleted := false

	attendance := &models.Attendance{
		ID:            uuid.New(),
		UserID:        input.UserID,
		Date:          input.Date,
		WorkdayHours:  input.WorkdayHours,
		PlannedStart:  input.PlannedStart,
		ActualStart:   input.ActualStart,
		Commits:       input.Commits,
		MergeRequests: input.MergeRequests,
		CodeReviews:   input.CodeReviews,
		EndWork:       input.EndWork,
		Status:        input.Status,
		Deleted:       &deleted,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}

	if err := s.repo.CreateAttendance(ctx, attendance); err != nil {
		return nil, err
	}

	return s.repo.GetAttendanceByID(ctx, attendance.ID)
}

func (s *attendanceService) UpdateAttendance(ctx context.Context, id uuid.UUID, input ports.UpdateAttendanceInput) (*models.Attendance, error) {
	updates := make(map[string]interface{})

	if input.Date != nil {
		updates["date"] = input.Date
	}
	if input.WorkdayHours != nil {
		updates["workday_hours"] = input.WorkdayHours
	}
	if input.PlannedStart != nil {
		updates["planned_start"] = input.PlannedStart
	}
	if input.ActualStart != nil {
		updates["actual_start"] = input.ActualStart
	}
	if input.Commits != nil {
		updates["commits"] = input.Commits
	}
	if input.MergeRequests != nil {
		updates["merge_requests"] = input.MergeRequests
	}
	if input.CodeReviews != nil {
		updates["code_reviews"] = input.CodeReviews
	}
	if input.EndWork != nil {
		updates["end_work"] = input.EndWork
	}
	if input.Status != nil {
		updates["status"] = input.Status
	}

	if len(updates) == 0 {
		return s.repo.GetAttendanceByID(ctx, id)
	}

	return s.repo.UpdateAttendance(ctx, id, updates)
}

func (s *attendanceService) DeleteAttendance(ctx context.Context, id uuid.UUID) error {
	return s.repo.SoftDeleteAttendance(ctx, id)
}

var _ ports.AttendanceService = (*attendanceService)(nil)
