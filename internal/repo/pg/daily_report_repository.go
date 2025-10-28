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

type DailyReportRepository struct {
	db *gorm.DB
}

func NewDailyReportRepository(
	db *gorm.DB,
) *DailyReportRepository {
	return &DailyReportRepository{
		db: db,
	}
}

func (r *DailyReportRepository) ListReports(
	ctx context.Context,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	var total int64
	base := r.db.WithContext(ctx).Model(&models.DailyReport{}).
		Where("daily_reports.deleted = FALSE OR daily_reports.deleted IS NULL")
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}

	var reports []models.DailyReport
	offset := (p.Page - 1) * p.PageSize
	if err := preloadReportRelations(
		base.Order("daily_reports.report_date DESC NULLS LAST, daily_reports.created_at DESC NULLS LAST"),
	).
		Limit(p.PageSize).
		Offset(offset).
		Find(&reports).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.DailyReport]{
		Items:      reports,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: total,
	}, nil
}

func (r *DailyReportRepository) ListReportsByUser(
	ctx context.Context,
	userID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	var total int64
	base := r.db.WithContext(ctx).
		Model(&models.DailyReport{}).
		Where(
			"daily_reports.user_id = ? AND (daily_reports.deleted = FALSE OR daily_reports.deleted IS NULL)",
			userID,
		)
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}

	var reports []models.DailyReport
	offset := (p.Page - 1) * p.PageSize
	if err := preloadReportRelations(
		base.Order("daily_reports.report_date DESC NULLS LAST, daily_reports.created_at DESC NULLS LAST"),
	).
		Limit(p.PageSize).
		Offset(offset).
		Find(&reports).Error; err != nil {
		return nil, err
	}

	return &ports.Page[models.DailyReport]{
		Items:      reports,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: total,
	}, nil
}

func (r *DailyReportRepository) ListReportsByProject(
	ctx context.Context,
	projectID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	var reports []models.DailyReport
	err := preloadReportRelations(r.db.WithContext(ctx).
		Model(&models.DailyReport{}).
		Select("DISTINCT daily_reports.*").
		Joins("LEFT JOIN completed_works ON completed_works.report_id = daily_reports.id AND (completed_works.deleted = FALSE OR completed_works.deleted IS NULL)").
		Joins("LEFT JOIN tasks ON tasks.id = completed_works.task_id AND (tasks.deleted = FALSE OR tasks.deleted IS NULL)").
		Joins("LEFT JOIN statuses ON statuses.id = tasks.status_id").
		Joins("LEFT JOIN boards ON boards.id = statuses.board_id").
		Where("daily_reports.deleted = FALSE OR daily_reports.deleted IS NULL").
		Where("boards.project_id = ?", projectID).
		Order("daily_reports.report_date DESC NULLS LAST, daily_reports.created_at DESC NULLS LAST")).
		Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return &ports.Page[models.DailyReport]{ // TODO Add pagination on db layer
		Items:      reports,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: 0,
	}, nil
}

func (r *DailyReportRepository) ListReportsByTask(
	ctx context.Context,
	taskID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	var reports []models.DailyReport
	err := preloadReportRelations(r.db.WithContext(ctx).
		Model(&models.DailyReport{}).
		Select("DISTINCT daily_reports.*").
		Joins("JOIN completed_works ON completed_works.report_id = daily_reports.id AND (completed_works.deleted = FALSE OR completed_works.deleted IS NULL)").
		Where("daily_reports.deleted = FALSE OR daily_reports.deleted IS NULL").
		Where("completed_works.task_id = ?", taskID).
		Order("daily_reports.report_date DESC NULLS LAST, daily_reports.created_at DESC NULLS LAST")).
		Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return &ports.Page[models.DailyReport]{ // TODO Add pagination on db layer
		Items:      reports,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: 0,
	}, nil
}

func (r *DailyReportRepository) ListReportsByDateRange(
	ctx context.Context,
	startDate,
	endDate time.Time,
	p ports.PaginationParams,
) (*ports.Page[models.DailyReport], error) {
	var reports []models.DailyReport
	err := preloadReportRelations(r.db.WithContext(ctx).
		Model(&models.DailyReport{}).
		Where("daily_reports.report_date BETWEEN ? AND ? AND (daily_reports.deleted = FALSE OR daily_reports.deleted IS NULL)", startDate, endDate).
		Order("daily_reports.report_date ASC, daily_reports.created_at ASC")).
		Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return &ports.Page[models.DailyReport]{ // TODO Add pagination on db layer
		Items:      reports,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: 0,
	}, nil
}

func (r *DailyReportRepository) ListHelpRequestsByHelper(
	ctx context.Context,
	helperID uuid.UUID,
	p ports.PaginationParams,
) (*ports.Page[models.HelpRequest], error) {
	var requests []models.HelpRequest
	err := r.db.WithContext(ctx).
		Model(&models.HelpRequest{}).
		Where("help_requests.helper_id = ? AND (help_requests.deleted = FALSE OR help_requests.deleted IS NULL)", helperID).
		Preload("Report", "reports.deleted = FALSE OR reports.deleted IS NULL").
		Preload("Report.User", "users.deleted = FALSE OR users.deleted IS NULL").
		Order("help_requests.created_at DESC NULLS LAST").
		Find(&requests).Error
	if err != nil {
		return nil, err
	}
	return &ports.Page[models.HelpRequest]{ // TODO Add pagination on db layer
		Items:      requests,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalCount: 0,
	}, nil
}

func (r *DailyReportRepository) GetReportByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.DailyReport, error) {
	var report models.DailyReport
	err := preloadReportRelations(r.db.WithContext(ctx).
		Model(&models.DailyReport{}).
		Where("daily_reports.id = ? AND (daily_reports.deleted = FALSE OR daily_reports.deleted IS NULL)", id)).
		First(&report).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &report, nil
}

func (r *DailyReportRepository) CreateReport(
	ctx context.Context,
	input ports.CreateDailyReportInput,
) (*models.DailyReport, error) {
	var reportID uuid.UUID
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		deleted := false

		reportID = uuid.New()
		report := &models.DailyReport{
			ID:         reportID,
			UserID:     input.UserID,
			ReportDate: input.ReportDate,
			CreatedAt:  &now,
			UpdatedAt:  &now,
			Deleted:    &deleted,
		}
		if err := tx.Create(report).Error; err != nil {
			return err
		}

		if err := r.persistCompletedWork(tx, reportID, input.CompletedWork); err != nil {
			return err
		}
		if err := r.persistHelpRequests(tx, reportID, input.HelpRequests); err != nil {
			return err
		}
		if err := r.persistTomorrowPlans(tx, reportID, input.TomorrowPlans); err != nil {
			return err
		}
		if err := r.persistReportProblems(tx, reportID, input.Problems); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetReportByID(ctx, reportID)
}

func (r *DailyReportRepository) UpdateReport(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateDailyReportInput,
) (*models.DailyReport, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{}
		if input.ReportDate != nil {
			updates["report_date"] = input.ReportDate
		}
		if input.Checked != nil {
			updates["checked"] = input.Checked
		}
		if len(updates) > 0 {
			updates["updated_at"] = timePtr(time.Now().UTC())
			res := tx.Model(&models.DailyReport{}).
				Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
				Updates(updates)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return domain.ErrNotFound
			}
		}

		if err := r.persistCompletedWork(tx, id, input.CompletedWork); err != nil {
			return err
		}
		if err := r.persistHelpRequests(tx, id, input.HelpRequests); err != nil {
			return err
		}
		if err := r.persistTomorrowPlans(tx, id, input.TomorrowPlans); err != nil {
			return err
		}
		if input.Problems != nil {
			if err := r.persistReportProblems(tx, id, input.Problems); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetReportByID(ctx, id)
}

func (r *DailyReportRepository) SoftDeleteReport(
	ctx context.Context,
	id uuid.UUID,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		truePtr := boolPtr(true)
		updates := map[string]interface{}{
			"deleted":    truePtr,
			"updated_at": &now,
		}
		res := tx.Model(&models.DailyReport{}).
			Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domain.ErrNotFound
		}

		tables := []interface{}{&models.CompletedWork{}, &models.HelpRequest{}, &models.TomorrowPlans{}, &models.ReportProblem{}}
		for _, table := range tables {
			if err := tx.Model(table).
				Where("report_id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
				Updates(map[string]interface{}{"deleted": truePtr, "updated_at": &now}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *DailyReportRepository) UpdateReportStorage(
	ctx context.Context,
	id uuid.UUID,
	storageObject string,
) error {
	updates := map[string]interface{}{
		"storage_object": storageObject,
		"updated_at":     timePtr(time.Now().UTC()),
	}
	res := r.db.WithContext(ctx).Model(&models.DailyReport{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DailyReportRepository) UpdateCompletedWork(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateCompletedWorkInput,
) (*models.CompletedWork, error) {
	updates := map[string]interface{}{"updated_at": timePtr(time.Now().UTC())}
	if input.Description != nil {
		updates["description"] = input.Description
	}
	if input.TaskID != nil {
		updates["task_id"] = input.TaskID
	}

	res := r.db.WithContext(ctx).Model(&models.CompletedWork{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.getCompletedWork(ctx, id)
}

func (r *DailyReportRepository) UpdateHelpRequest(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateHelpRequestInput,
) (*models.HelpRequest, error) {
	updates := map[string]interface{}{"updated_at": timePtr(time.Now().UTC())}
	if input.Description != nil {
		updates["description"] = input.Description
	}
	if input.HelperID != nil {
		updates["helper_id"] = input.HelperID
	}
	if input.Status != nil {
		updates["status"] = input.Status
	}

	res := r.db.WithContext(ctx).Model(&models.HelpRequest{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.getHelpRequest(ctx, id)
}

func (r *DailyReportRepository) SoftDeleteHelpRequest(
	ctx context.Context,
	id uuid.UUID,
) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).
		Model(&models.HelpRequest{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(map[string]interface{}{"deleted": boolPtr(true), "updated_at": &now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *DailyReportRepository) UpdateTomorrowPlan(
	ctx context.Context,
	id uuid.UUID,
	input ports.UpdateTomorrowPlanInput,
) (*models.TomorrowPlans, error) {
	updates := map[string]interface{}{"updated_at": timePtr(time.Now().UTC())}
	if input.Description != nil {
		updates["description"] = input.Description
	}
	if input.TaskID != nil {
		updates["task_id"] = input.TaskID
	}

	res := r.db.WithContext(ctx).Model(&models.TomorrowPlans{}).
		Where("id = ? AND (deleted = FALSE OR deleted IS NULL)", id).
		Updates(updates)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}

	return r.getTomorrowPlan(ctx, id)
}

func (r *DailyReportRepository) persistCompletedWork(
	tx *gorm.DB,
	reportID uuid.UUID,
	items []ports.CompletedWorkInput,
) error {
	if len(items) == 0 {
		return nil
	}
	for _, item := range items {
		if item.ID != nil {
			updates := map[string]interface{}{"updated_at": timePtr(time.Now().UTC())}
			if item.Description != nil {
				updates["description"] = item.Description
			}
			if item.TaskID != nil {
				updates["task_id"] = item.TaskID
			}
			res := tx.Model(&models.CompletedWork{}).
				Where("id = ? AND report_id = ? AND (deleted = FALSE OR deleted IS NULL)", item.ID, reportID).
				Updates(updates)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return domain.ErrNotFound
			}
			continue
		}

		cw := &models.CompletedWork{
			ID:          uuid.New(),
			ReportID:    &reportID,
			Description: item.Description,
			TaskID:      item.TaskID,
			Deleted:     boolPtr(false),
			CreatedAt:   timePtr(time.Now().UTC()),
			UpdatedAt:   timePtr(time.Now().UTC()),
		}
		if err := tx.Create(cw).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *DailyReportRepository) persistHelpRequests(
	tx *gorm.DB,
	reportID uuid.UUID,
	items []ports.HelpRequestInput,
) error {
	if len(items) == 0 {
		return nil
	}
	for _, item := range items {
		if item.ID != nil {
			updates := map[string]interface{}{"updated_at": timePtr(time.Now().UTC())}
			if item.Description != nil {
				updates["description"] = item.Description
			}
			if item.HelperID != nil {
				updates["helper_id"] = item.HelperID
			}
			if item.Status != nil {
				updates["status"] = item.Status
			}
			res := tx.Model(&models.HelpRequest{}).
				Where("id = ? AND report_id = ? AND (deleted = FALSE OR deleted IS NULL)", item.ID, reportID).
				Updates(updates)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return domain.ErrNotFound
			}
			continue
		}

		req := &models.HelpRequest{
			ID:          uuid.New(),
			ReportID:    &reportID,
			HelperID:    item.HelperID,
			Description: item.Description,
			Status:      item.Status,
			Deleted:     boolPtr(false),
			CreatedAt:   timePtr(time.Now().UTC()),
			UpdatedAt:   timePtr(time.Now().UTC()),
		}
		if err := tx.Create(req).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *DailyReportRepository) persistTomorrowPlans(
	tx *gorm.DB,
	reportID uuid.UUID,
	items []ports.TomorrowPlanInput,
) error {
	if len(items) == 0 {
		return nil
	}
	for _, item := range items {
		if item.ID != nil {
			updates := map[string]interface{}{"updated_at": timePtr(time.Now().UTC())}
			if item.Description != nil {
				updates["description"] = item.Description
			}
			if item.TaskID != nil {
				updates["task_id"] = item.TaskID
			}
			res := tx.Model(&models.TomorrowPlans{}).
				Where("id = ? AND report_id = ? AND (deleted = FALSE OR deleted IS NULL)", item.ID, reportID).
				Updates(updates)
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return domain.ErrNotFound
			}
			continue
		}

		plan := &models.TomorrowPlans{
			ID:          uuid.New(),
			ReportID:    &reportID,
			TaskID:      item.TaskID,
			Description: item.Description,
			Deleted:     boolPtr(false),
			CreatedAt:   timePtr(time.Now().UTC()),
			UpdatedAt:   timePtr(time.Now().UTC()),
		}
		if err := tx.Create(plan).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *DailyReportRepository) persistReportProblems(
	tx *gorm.DB,
	reportID uuid.UUID,
	desired []uuid.UUID,
) error {
	if desired == nil {
		return nil
	}
	var current []models.ReportProblem
	if err := tx.Model(&models.ReportProblem{}).
		Where("report_id = ? AND (deleted = FALSE OR deleted IS NULL)", reportID).
		Find(&current).Error; err != nil {
		return err
	}

	desiredSet := make(map[uuid.UUID]struct{}, len(desired))
	for _, id := range desired {
		desiredSet[id] = struct{}{}
	}

	now := time.Now().UTC()
	for _, item := range current {
		if _, ok := desiredSet[item.ProblemID]; !ok {
			if err := tx.Model(&models.ReportProblem{}).
				Where("report_id = ? AND problem_id = ? AND (deleted = FALSE OR deleted IS NULL)", reportID, item.ProblemID).
				Updates(map[string]interface{}{"deleted": boolPtr(true), "updated_at": &now}).Error; err != nil {
				return err
			}
		}
	}

	currentSet := make(map[uuid.UUID]struct{}, len(current))
	for _, item := range current {
		currentSet[item.ProblemID] = struct{}{}
	}

	for _, problemID := range desired {
		if _, ok := currentSet[problemID]; ok {
			continue
		}
		relation := &models.ReportProblem{
			ReportID:  reportID,
			ProblemID: problemID,
			CreatedAt: &now,
			UpdatedAt: &now,
			Deleted:   boolPtr(false),
		}
		if err := tx.Create(relation).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *DailyReportRepository) getCompletedWork(
	ctx context.Context,
	id uuid.UUID,
) (*models.CompletedWork, error) {
	var cw models.CompletedWork
	err := r.db.WithContext(ctx).
		Model(&models.CompletedWork{}).
		Preload("Report", "reports.deleted = FALSE OR reports.deleted IS NULL").
		Where("completed_works.id = ? AND (completed_works.deleted = FALSE OR completed_works.deleted IS NULL)", id).
		First(&cw).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &cw, nil
}

func (r *DailyReportRepository) getHelpRequest(
	ctx context.Context,
	id uuid.UUID,
) (*models.HelpRequest, error) {
	var hr models.HelpRequest
	err := r.db.WithContext(ctx).
		Model(&models.HelpRequest{}).
		Where("help_requests.id = ? AND (help_requests.deleted = FALSE OR help_requests.deleted IS NULL)", id).
		First(&hr).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &hr, nil
}

func (r *DailyReportRepository) getTomorrowPlan(
	ctx context.Context,
	id uuid.UUID,
) (*models.TomorrowPlans, error) {
	var plan models.TomorrowPlans
	err := r.db.WithContext(ctx).
		Model(&models.TomorrowPlans{}).
		Where("tomorrow_plans.id = ? AND (tomorrow_plans.deleted = FALSE OR tomorrow_plans.deleted IS NULL)", id).
		First(&plan).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &plan, nil
}

func boolPtr(value bool) *bool {
	return &value
}

func preloadReportRelations(
	db *gorm.DB,
) *gorm.DB {
	return db.
		Preload("User", "users.deleted = FALSE OR users.deleted IS NULL").
		Preload("CompletedWork", "completed_works.deleted = FALSE OR completed_works.deleted IS NULL").
		Preload("CompletedWork.Task").
		Preload("HelpRequests", "help_requests.deleted = FALSE OR help_requests.deleted IS NULL").
		Preload("TomorrowPlans", "tomorrow_plans.deleted = FALSE OR tomorrow_plans.deleted IS NULL").
		Preload("ReportProblems", "report_problems.deleted = FALSE OR report_problems.deleted IS NULL").
		Preload("ReportProblems.Problem", "problems.deleted = FALSE OR problems.deleted IS NULL")
}

var _ ports.DailyReportRepository = (*DailyReportRepository)(nil)
