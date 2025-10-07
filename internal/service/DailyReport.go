package service

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/repository"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ReportService interface {
	GetAllReports(page, pageSize int) ([]models.DailyReport, int64, error)
	GetAllReportsByUserId(userID uuid.UUID, page, pageSize int) ([]models.DailyReport, int64, error)
	GetReport(reportID uuid.UUID) (*models.DailyReport, error)
	GetReportsByTaskId(taskID uuid.UUID) ([]models.DailyReport, error)
	GetReportsByProjectId(projectID uuid.UUID) ([]models.DailyReport, error)
	CreateReport(req request.ReportCreateRequest) (uuid.UUID, error)
	UpdateReport(reportID uuid.UUID, req request.ReportReplaceRequest) error
	DeleteReport(reportID uuid.UUID) error
	UpdateHelpRequest(helpID uuid.UUID, req request.HelpRequestUpdateRequest) error
	UpdateCompletedWork(cwID uuid.UUID, req request.CompletedWorkUpdateRequest) error
	UpdateTomorrowPlans(tpID uuid.UUID, req request.TomorrowPlansUpdateRequest) error
	GetHelpRequestsForUser(userID uuid.UUID) ([]models.HelpRequest, error)
	DeleteHelpRequest(requestID uuid.UUID) error
}

type reportService struct {
	repo repository.ReportRepository
}

func NewReportService(repo repository.ReportRepository) ReportService {
	return &reportService{
		repo: repo,
	}
}

func (s *reportService) GetAllReports(page, pageSize int) ([]models.DailyReport, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetAllReports(pageSize, offset)
}

func (s *reportService) GetAllReportsByUserId(userID uuid.UUID, page, pageSize int) ([]models.DailyReport, int64, error) {
	offset := (page - 1) * pageSize
	return s.repo.GetAllReportsByUserId(userID, pageSize, offset)
}

func (s *reportService) GetReport(reportID uuid.UUID) (*models.DailyReport, error) {
	return s.repo.GetReport(reportID)
}

func (s *reportService) GetReportsByTaskId(taskID uuid.UUID) ([]models.DailyReport, error) {
	return s.repo.GetReportsByTaskId(taskID)
}

func (s *reportService) GetReportsByProjectId(projectID uuid.UUID) ([]models.DailyReport, error) {
	return s.repo.GetReportsByProjectId(projectID)
}

func (s *reportService) CreateReport(req request.ReportCreateRequest) (uuid.UUID, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return uuid.Nil, errors.New("invalid user id")
	}

	now := time.Now()
	delFalse := false
	zero := int8(0)
	rep := models.DailyReport{
		ID:         uuid.New(),
		UserID:     userID,
		ReportDate: req.ReportDate,
		CreatedAt:  &now,
		Deleted:    &delFalse,
		Checked:    &zero,
	}

	err = s.repo.CreateReportWithRelations(rep, req)
	if err != nil {
		return uuid.Nil, err
	}

	return rep.ID, nil
}

func (s *reportService) UpdateReport(reportID uuid.UUID, req request.ReportReplaceRequest) error {
	report, err := s.repo.GetReport(reportID)
	if err != nil {
		if err.Error() == "report not found" {
			return errors.New("report not found")
		}
		return err
	}

	now := time.Now()
	if err := s.repo.Transaction(func(tx repository.ReportRepository) error {
		updateData := map[string]interface{}{"updated_at": &now}
		if req.UserId != "" {
			if uid, err := uuid.Parse(req.UserId); err == nil {
				updateData["user_id"] = uid
			} else {
				return errors.New("invalid user_id")
			}
		}
		if req.ReportDate != nil {
			updateData["report_date"] = req.ReportDate
		}
		if req.Checked != nil {
			updateData["checked"] = req.Checked
		}
		// Передаем разыменованный report
		if err := tx.UpdateReport(*report, updateData); err != nil {
			return err
		}

		// Обновляем связанные сущности
		return tx.UpdateReportRelations(reportID, req, now)
	}); err != nil {
		return err
	}

	return nil
}

func (s *reportService) DeleteReport(reportID uuid.UUID) error {
	deleted, err := s.repo.DeleteReport(reportID)
	if err != nil {
		return err
	}

	if !deleted {
		return errors.New("report not found")
	}

	return nil
}

func (s *reportService) UpdateHelpRequest(helpID uuid.UUID, req request.HelpRequestUpdateRequest) error {
	updateData := make(map[string]interface{})
	if req.Description != nil {
		updateData["description"] = *req.Description
	}
	if req.HelperID != nil {
		updateData["helper_id"] = *req.HelperID
	}
	if req.Status != nil {
		updateData["status"] = *req.Status
	}

	if len(updateData) == 0 {
		return errors.New("no fields to update")
	}

	now := time.Now()
	updateData["updated_at"] = &now

	updated, err := s.repo.UpdateHelpRequest(helpID, updateData)
	if err != nil {
		return err
	}

	if !updated {
		return errors.New("help request not found")
	}

	return nil
}

func (s *reportService) UpdateCompletedWork(cwID uuid.UUID, req request.CompletedWorkUpdateRequest) error {
	updateData := make(map[string]interface{})
	if req.Description != nil {
		updateData["description"] = *req.Description
	}

	if len(updateData) == 0 {
		return errors.New("no fields to update")
	}

	now := time.Now()
	updateData["updated_at"] = &now

	updated, err := s.repo.UpdateCompletedWork(cwID, updateData)
	if err != nil {
		return err
	}

	if !updated {
		return errors.New("completed work not found")
	}

	return nil
}

func (s *reportService) UpdateTomorrowPlans(tpID uuid.UUID, req request.TomorrowPlansUpdateRequest) error {
	updateData := make(map[string]interface{})
	if req.Description != nil {
		updateData["description"] = *req.Description
	}

	if len(updateData) == 0 {
		return errors.New("no fields to update")
	}

	now := time.Now()
	updateData["updated_at"] = &now

	updated, err := s.repo.UpdateTomorrowPlans(tpID, updateData)
	if err != nil {
		return err
	}

	if !updated {
		return errors.New("tomorrow plans not found")
	}

	return nil
}

func (s *reportService) GetHelpRequestsForUser(userID uuid.UUID) ([]models.HelpRequest, error) {
	return s.repo.GetHelpRequestsForUser(userID)
}

func (s *reportService) DeleteHelpRequest(requestID uuid.UUID) error {
	deleted, err := s.repo.DeleteHelpRequest(requestID)
	if err != nil {
		return err
	}

	if !deleted {
		return errors.New("help request not found")
	}

	return nil
}