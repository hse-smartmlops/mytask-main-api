package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"emplacc-api/internal/app/ports"
	"emplacc-api/internal/config"
	"emplacc-api/internal/domain"
	"emplacc-api/internal/domain/models"

	"github.com/google/uuid"
)

type dailyReportService struct {
	repo    ports.DailyReportRepository
	storage ports.ObjectStorage
	cfg     config.MinioConfig
	config  *config.PaginationConfig
}

func NewDailyReportService(
	repo ports.DailyReportRepository,
	storage ports.ObjectStorage,
	cfg config.MinioConfig, config *config.PaginationConfig,
) ports.DailyReportService {
	return &dailyReportService{
		repo:    repo,
		storage: storage,
		cfg:     cfg,
		config:  config,
	}
}

func (s *dailyReportService) ListReports(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	checkPagination(&p, s.config)
	return s.repo.ListReports(ctx, p)
}

func (s *dailyReportService) ListReportsByUser(
	ctx context.Context,
	userID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	checkPagination(&p, s.config)
	return s.repo.ListReportsByUser(ctx, userID, p)
}

func (s *dailyReportService) ListReportsByProject(
	ctx context.Context,
	projectID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	checkPagination(&p, s.config)
	return s.repo.ListReportsByProject(ctx, projectID, p)
}

func (s *dailyReportService) ListReportsByTask(
	ctx context.Context,
	taskID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	return s.repo.ListReportsByTask(ctx, taskID, p)
}

func (s *dailyReportService) ListReportsByDateRange(
	ctx context.Context,
	startDate,
	endDate time.Time,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	if endDate.Before(startDate) {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.ListReportsByDateRange(ctx, startDate, endDate, p)
}

func (s *dailyReportService) ListHelpRequestsByHelper(
	ctx context.Context,
	helperID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.HelpRequest], error) {
	return s.repo.ListHelpRequestsByHelper(ctx, helperID, p)
}

func (s *dailyReportService) GetReport(
	ctx context.Context,
	id uuid.UUID,
) (*models.DailyReport, error) {
	report, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if report == nil {
		return nil, domain.ErrNotFound
	}
	return report, nil
}

func (s *dailyReportService) CreateReport(
	ctx context.Context,
	input ports.CreateDailyReportInput,
) (*models.DailyReport, error) {
	if input.UserID == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}

	sanitizeReportInputs(&input)
	report, err := s.repo.CreateReport(ctx, input)
	if err != nil {
		return nil, err
	}

	if err := s.syncReportFile(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *dailyReportService) UpdateReport(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateDailyReportInput,
) (*models.DailyReport, error) {
	sanitizeUpdateReportInputs(&input)
	report, err := s.repo.UpdateReport(ctx, id, input)
	if err != nil {
		return nil, err
	}

	if err := s.syncReportFile(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}

func (s *dailyReportService) DeleteReport(
	ctx context.Context,
	id uuid.UUID,
) error {
	report, err := s.repo.GetReportByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.SoftDeleteReport(ctx, id); err != nil {
		return err
	}

	return s.removeReportFile(ctx, report)
}

func (s *dailyReportService) UpdateCompletedWork(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateCompletedWorkInput,
) (*models.CompletedWork, error) {
	input.Description = cloneStringPtr(input.Description)
	return s.repo.UpdateCompletedWork(ctx, id, input)
}

func (s *dailyReportService) UpdateHelpRequest(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateHelpRequestInput,
) (*models.HelpRequest, error) {
	input.Description = cloneStringPtr(input.Description)
	input.Status = cloneStringPtr(input.Status)
	return s.repo.UpdateHelpRequest(ctx, id, input)
}

func (s *dailyReportService) DeleteHelpRequest(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.repo.SoftDeleteHelpRequest(ctx, id)
}

func (s *dailyReportService) UpdateTomorrowPlan(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateTomorrowPlanInput,
) (*models.TomorrowPlans, error) {
	input.Description = cloneStringPtr(input.Description)
	return s.repo.UpdateTomorrowPlan(ctx, id, input)
}

func (s *dailyReportService) ExportReportsToXLSX(
	ctx context.Context,
	input ports.ReportsByDateInput,
	fileName string,
) ([]byte, error) {
	if input.EndDate.Before(input.StartDate) {
		return nil, domain.ErrInvalidInput
	}

	p := ports.PaginationParams{Page: 1, PageSize: 0} // TODO get all reports
	reports, err := s.ListReportsByDateRange(ctx, input.StartDate, input.EndDate, p)
	if err != nil {
		return nil, err
	}

	data, err := buildDailyReportsWorkbook(reports.Items, input.StartDate, input.EndDate)
	if err != nil {
		return nil, err
	}

	if s.storage != nil {
		object := reportExportObjectKey(fileName)
		metadata := map[string]string{
			"start_date": input.StartDate.Format(time.RFC3339),
			"end_date":   input.EndDate.Format(time.RFC3339),
		}
		if err := s.storage.Upload(ctx, s.reportBucket(), object, data, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", metadata); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func (s *dailyReportService) syncReportFile(
	ctx context.Context,
	report *models.DailyReport,
) error {
	if s.storage == nil || report == nil {
		return nil
	}

	data, err := serializeReport(report)
	if err != nil {
		return err
	}

	bucket := s.reportBucket()
	object := reportObjectKey(report.ID)
	path := fmt.Sprintf("%s/%s", bucket, object)

	if err := s.storage.Upload(ctx, bucket, object, data, "application/json", map[string]string{"report_id": report.ID.String()}); err != nil {
		return err
	}

	if err := s.repo.UpdateReportStorage(ctx, report.ID, path); err != nil {
		return err
	}

	report.StorageObject = path
	return nil
}

func (s *dailyReportService) removeReportFile(
	ctx context.Context,
	report *models.DailyReport,
) error {
	if s.storage == nil || report == nil {
		return nil
	}

	bucket, object := splitStoragePath(report.StorageObject)
	if bucket == "" || object == "" {
		return nil
	}

	return s.storage.Delete(ctx, bucket, object)
}

func (s *dailyReportService) reportBucket() string {
	bucket := strings.TrimSpace(s.cfg.ReportBucket)
	if bucket == "" {
		return "reports"
	}
	return bucket
}

var _ ports.DailyReportService = (*dailyReportService)(nil)
