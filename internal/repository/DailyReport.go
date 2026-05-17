package repository

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReportRepository interface {
	GetAllReports(limit, offset int) ([]models.DailyReport, int64, error)
	GetAllReportsByUserId(userID uuid.UUID, limit, offset int) ([]models.DailyReport, int64, error)
	GetReport(reportID uuid.UUID) (*models.DailyReport, error)
	GetReportsByTaskId(taskID uuid.UUID) ([]models.DailyReport, error)
	GetReportsByProjectId(projectID uuid.UUID) ([]models.DailyReport, error)
	CreateReportWithRelations(rep models.DailyReport, req request.ReportCreateRequest) error
	UpdateReport(report models.DailyReport, updateData map[string]interface{}) error
	UpdateReportRelations(reportID uuid.UUID, req request.ReportReplaceRequest, now time.Time) error
	DeleteReport(reportID uuid.UUID) (bool, error)
	UpdateHelpRequest(helpID uuid.UUID, updateData map[string]interface{}) (bool, error)
	UpdateCompletedWork(cwID uuid.UUID, updateData map[string]interface{}) (bool, error)
	UpdateTomorrowPlans(tpID uuid.UUID, updateData map[string]interface{}) (bool, error)
	GetHelpRequestsForUser(userID uuid.UUID) ([]models.HelpRequest, error)
	DeleteHelpRequest(requestID uuid.UUID) (bool, error)
	Transaction(txFunc func(ReportRepository) error) error
	GetReportByDateInXLSX(startDate, endDate time.Time) ([]models.DailyReport, error)
	GetLatestTomorrowPlans() ([]models.TomorrowPlans, error)
}

type reportRepository struct {
	db *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{
		db: db,
	}
}

func (r *reportRepository) GetAllReports(limit, offset int) ([]models.DailyReport, int64, error) {
	var totalCount int64
	if err := r.db.
		Table("daily_reports").
		Where("daily_reports.deleted = FALSE").
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var reports []models.DailyReport
	if err := r.db.
		Model(&models.DailyReport{}).
		Preload("User", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("HelpRequests", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("CompletedWork", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("TomorrowPlans", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("ReportProblems", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("ReportProblems.Problem", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Where("daily_reports.deleted = FALSE").
		Order("daily_reports.report_date DESC NULLS LAST, daily_reports.created_at DESC NULLS LAST").
		Limit(limit).Offset(offset).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, totalCount, nil
}

func (r *reportRepository) GetReportByDateInXLSX(startDate, endDate time.Time) ([]models.DailyReport, error) {
	var reports []models.DailyReport

	if err := r.db.
		Model(&models.DailyReport{}).
		Preload("User", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("CompletedWork", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("CompletedWork.Task.Status.Board.Project").
		Where("daily_reports.deleted = FALSE AND daily_reports.report_date BETWEEN ? AND ?", startDate, endDate).
		Order("daily_reports.report_date DESC, daily_reports.created_at DESC").
		Find(&reports).Error; err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *reportRepository) GetAllReportsByUserId(userID uuid.UUID, limit, offset int) ([]models.DailyReport, int64, error) {
	var totalCount int64
	if err := r.db.
		Model(&models.DailyReport{}).
		Where("daily_reports.deleted = FALSE and daily_reports.user_id = ?", userID).
		Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	var reports []models.DailyReport
	if err := r.db.
		Model(&models.DailyReport{}).
		Preload("User", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("HelpRequests", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("CompletedWork", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("TomorrowPlans", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("ReportProblems", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("ReportProblems.Problem", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Where("daily_reports.deleted = FALSE and daily_reports.user_id = ?", userID).
		Order("daily_reports.report_date DESC NULLS LAST, daily_reports.created_at DESC NULLS LAST").
		Limit(limit).Offset(offset).
		Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, totalCount, nil
}

func (r *reportRepository) GetReport(reportID uuid.UUID) (*models.DailyReport, error) {
	var report models.DailyReport
	if err := r.db.
		Model(&models.DailyReport{}).
		Preload("User", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("HelpRequests", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("CompletedWork", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("CompletedWork.Task", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("CompletedWork.Task.Status", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("CompletedWork.Task.Status.Board", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("CompletedWork.Task.Status.Board.Project", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("TomorrowPlans", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("TomorrowPlans.Task", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("TomorrowPlans.Task.Status", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("TomorrowPlans.Task.Status.Board", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("TomorrowPlans.Task.Status.Board.Project", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("ReportProblems", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Preload("ReportProblems.Problem", func(db *gorm.DB) *gorm.DB {return db.Where("deleted = FALSE")}).
		Where("daily_reports.deleted = FALSE AND daily_reports.id = ?", reportID).
		First(&report).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("report not found")
		}
		return nil, err
	}
	return &report, nil
}

func (r *reportRepository) GetLatestTomorrowPlans() ([]models.TomorrowPlans, error) {
    var plans []models.TomorrowPlans
    
    subquery := r.db.
        Model(&models.DailyReport{}).
        Select("id").
        Where("daily_reports.deleted = ?", false).
        Where("(daily_reports.user_id, daily_reports.report_date) IN (?)", 
            r.db.
                Model(&models.DailyReport{}).
                Select("user_id, MAX(report_date)").
                Where("daily_reports.deleted = ?", false).
                Group("user_id"),
        )

    err := r.db.
        Model(&models.TomorrowPlans{}).
        Where("tomorrow_plans.deleted = ? AND tomorrow_plans.report_id IN (?)", false, subquery).
        Preload("Report", func(db *gorm.DB) *gorm.DB {
            return db.Where("deleted = ?", false)
        }).
        Preload("Report.User", func(db *gorm.DB) *gorm.DB {
            return db.Where("deleted = ?", false)
        }).
        Preload("Task", func(db *gorm.DB) *gorm.DB {
            return db.Where("deleted = ?", false)
        }).
        Preload("Task.Status", func(db *gorm.DB) *gorm.DB {
            return db.Where("deleted = ?", false)
        }).
        Preload("Task.Status.Board", func(db *gorm.DB) *gorm.DB {
            return db.Where("deleted = ?", false)
        }).
        Preload("Task.Status.Board.Project", func(db *gorm.DB) *gorm.DB {
            return db.Where("deleted = ?", false)
        }).
        Order("report_id, created_at").
        Find(&plans).Error

    if err != nil {
        return nil, err
    }

    return plans, nil
}

func (r *reportRepository) GetReportsByTaskId(taskID uuid.UUID) ([]models.DailyReport, error) {
	var cw []models.CompletedWork
	if err := r.db.
		Model(&models.CompletedWork{}).
		Where("completed_work.deleted = FALSE AND completed_work.task_id = ?", taskID).
		Find(&cw).Error; err != nil {
		return nil, err
	}

	reportIDs := make(map[uuid.UUID]struct{})
	for _, w := range cw {
		if w.ReportID != nil {
			reportIDs[*w.ReportID] = struct{}{}
		}
	}

	if len(reportIDs) == 0 {
		return []models.DailyReport{}, nil
	}

	ids := make([]uuid.UUID, 0, len(reportIDs))
	for id := range reportIDs {
		ids = append(ids, id)
	}

	var reports []models.DailyReport
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.DailyReport{}).
		Preload("User", "deleted = FALSE").
		Preload("HelpRequests", "deleted = FALSE").
		Preload("CompletedWork", "deleted = FALSE").
		Preload("TomorrowPlans", "deleted = FALSE").
		Preload("ReportProblems", "deleted = FALSE").
		Preload("ReportProblems.Problem", "deleted = FALSE").
		Where("deleted = FALSE AND id IN ?", ids).
		Find(&reports).Error; err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *reportRepository) GetReportsByProjectId(projectID uuid.UUID) ([]models.DailyReport, error) {
	var statusIDs []uuid.UUID
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Select("statuses.id").
		Joins("JOIN boards ON statuses.board_id = boards.id").
		Where("boards.project_id = ? AND boards.deleted = ? AND statuses.deleted = ?", projectID, false, false).
		Scan(&statusIDs).Error; err != nil {
		return nil, err
	}

	if len(statusIDs) == 0 {
		return []models.DailyReport{}, nil
	}

	var tasks []models.Task
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.Task{}).
		Where("status_id IN ? AND deleted = ?", statusIDs, false).
		Find(&tasks).Error; err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return []models.DailyReport{}, nil
	}

	taskIDs := make([]uuid.UUID, len(tasks))
	for i, t := range tasks {
		taskIDs[i] = t.ID
	}

	var completedWorks []models.CompletedWork
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.CompletedWork{}).
		Where("task_id IN ? AND deleted = ?", taskIDs, false).
		Find(&completedWorks).Error; err != nil {
		return nil, err
	}

	if len(completedWorks) == 0 {
		return []models.DailyReport{}, nil
	}

	reportIDSet := make(map[uuid.UUID]struct{})
	for _, cw := range completedWorks {
		if cw.ReportID != nil {
			reportIDSet[*cw.ReportID] = struct{}{}
		}
	}

	if len(reportIDSet) == 0 {
		return []models.DailyReport{}, nil
	}

	reportIDs := make([]uuid.UUID, 0, len(reportIDSet))
	for id := range reportIDSet {
		reportIDs = append(reportIDs, id)
	}

	var reports []models.DailyReport
	if err := r.db.Session(&gorm.Session{}).
		Preload("User", "deleted = FALSE").
		Preload("HelpRequests", "deleted = FALSE").
		Preload("CompletedWork", "deleted = FALSE").
		Preload("TomorrowPlans", "deleted = FALSE").
		Preload("ReportProblems", "deleted = FALSE").
		Preload("ReportProblems.Problem", "deleted = FALSE").
		Where("id IN ? AND deleted = ?", reportIDs, false).
		Find(&reports).Error; err != nil {
		return nil, err
	}

	return reports, nil
}

func (r *reportRepository) CreateReportWithRelations(rep models.DailyReport, req request.ReportCreateRequest) error {
	now := time.Now()
	delFalse := false

	return r.db.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).Table("LDailyReports").Create(&rep); res.Error != nil {
			return res.Error
		}

		if len(req.CompleteWork) > 0 {
			batch := make([]models.CompletedWork, 0, len(req.CompleteWork))
			for _, w := range req.CompleteWork {
				item := models.CompletedWork{
					ID:          uuid.New(),
					Description: &w.Description,
					ReportID:    &rep.ID,
					Deleted:     &delFalse,
					CreatedAt:   &now,
				}
				if w.TaskID != "" {
					if tid, err := uuid.Parse(w.TaskID); err == nil {
						item.TaskID = &tid
					}
				}
				batch = append(batch, item)
			}
			if res := tx.Session(&gorm.Session{}).Table("LCompletedWorks").Create(&batch); res.Error != nil {
				return res.Error
			}
		}

		if len(req.PlanTomorrow) > 0 {
			batch := make([]models.TomorrowPlans, 0, len(req.PlanTomorrow))
			for _, p := range req.PlanTomorrow {
				item := models.TomorrowPlans{
					ID:          uuid.New(),
					Description: &p.Description,
					ReportID:    &rep.ID,
					Deleted:     &delFalse,
					CreatedAt:   &now,
				}
				if p.TaskId != "" {
					if tid, err := uuid.Parse(p.TaskId); err == nil {
						item.TaskID = &tid
					}
				}
				batch = append(batch, item)
			}
			if res := tx.Session(&gorm.Session{}).Table("LTomorrowPlanss").Create(&batch); res.Error != nil {
				return res.Error
			}
		}

		if len(req.Problems) > 0 {
			batch := make([]models.ReportProblem, 0, len(req.Problems))
			for _, pr := range req.Problems {
				pid, err := uuid.Parse(pr)
				if err != nil {
					return errors.New("invalid problemId")
				}
				batch = append(batch, models.ReportProblem{
					ReportID:  rep.ID,
					ProblemID: pid,
					Deleted:   &delFalse,
					CreatedAt: &now,
				})
			}
			if res := tx.Session(&gorm.Session{}).Table("LReportProblems").Create(&batch); res.Error != nil {
				return res.Error
			}
		}

		type helpItem struct {
			HelperID    string
			Description string
			Status      *string
		}
		items := []helpItem{}
		if len(req.Helps) > 0 {
			for _, h := range req.Helps {
				items = append(items, helpItem{HelperID: h.HelperID, Description: h.Description, Status: h.Status})
			}
		}
		if len(items) > 0 {
			batch := make([]models.HelpRequest, 0, len(items))
			for _, hi := range items {
				hid, err := uuid.Parse(hi.HelperID)
				if err != nil {
					return errors.New("invalid helperId")
				}
				batch = append(batch, models.HelpRequest{
					ID:          uuid.New(),
					HelperID:    &hid,
					Description: &hi.Description,
					Status:      hi.Status,
					ReportID:    &rep.ID,
					Deleted:     &delFalse,
					CreatedAt:   &now,
				})
			}
			if res := tx.Session(&gorm.Session{}).Table("LHelpRequests").Create(&batch); res.Error != nil {
				return res.Error
			}
		}

		return nil
	})
}

func (r *reportRepository) UpdateReport(report models.DailyReport, updateData map[string]interface{}) error {
	return r.db.Model(&report).Updates(updateData).Error
}

func (r *reportRepository) UpdateReportRelations(reportID uuid.UUID, req request.ReportReplaceRequest, now time.Time) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var existingCWs []models.CompletedWork
		var existingTPs []models.TomorrowPlans
		var existingHelps []models.HelpRequest
		var existingProblems []models.ReportProblem

		tx.Where("report_id = ?", reportID).Find(&existingCWs)
		tx.Where("report_id = ?", reportID).Find(&existingTPs)
		tx.Where("report_id = ?", reportID).Find(&existingHelps)
		tx.Where("report_id = ?", reportID).Find(&existingProblems)

		cwMap := make(map[uuid.UUID]models.CompletedWork)
		for _, cw := range existingCWs {
			cwMap[cw.ID] = cw
		}
		tpMap := make(map[uuid.UUID]models.TomorrowPlans)
		for _, tp := range existingTPs {
			tpMap[tp.ID] = tp
		}
		helpMap := make(map[uuid.UUID]models.HelpRequest)
		for _, h := range existingHelps {
			helpMap[h.ID] = h
		}
		problemMap := make(map[uuid.UUID]models.ReportProblem)
		for _, rp := range existingProblems {
			problemMap[rp.ProblemID] = rp
		}

		var cwInserts []models.CompletedWork
		var cwUpdates []map[string]interface{}
		for _, cw := range req.CompleteWork {
			var cwID uuid.UUID
			if cw.ID != nil && *cw.ID != "" {
				parsedID, err := uuid.Parse(*cw.ID)
				if err != nil {
					return errors.New("invalid id in completed_work")
				}
				cwID = parsedID
			} else {
				cwID = uuid.New()
			}
			var taskUUID *uuid.UUID
			if cw.TaskID != nil && *cw.TaskID != "" {
				if parsed, err := uuid.Parse(*cw.TaskID); err == nil {
					taskUUID = &parsed
				} else {
					return errors.New("invalid task_id in completed_work")
				}
			}
			if _, ok := cwMap[cwID]; ok {
				cwUpdates = append(cwUpdates, map[string]interface{}{
					"id":          cwID,
					"description": cw.Description,
					"task_id":     taskUUID,
					"updated_at":  now,
				})
			} else {
				cwInserts = append(cwInserts, models.CompletedWork{
					ID:          cwID,
					Description: &cw.Description,
					ReportID:    &reportID,
					TaskID:      taskUUID,
					CreatedAt:   &now,
					UpdatedAt:   &now,
					Deleted:     func() *bool { b := false; return &b }(),
				})
			}
		}

		if len(cwInserts) > 0 {
			if err := tx.Create(&cwInserts).Error; err != nil {
				return err
			}
		}
		if len(cwUpdates) > 0 {
			tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"description", "task_id", "updated_at"}),
			}).Create(&cwUpdates)
		}

		var tpInserts []models.TomorrowPlans
		var tpUpdates []map[string]interface{}
		for _, tp := range req.PlanTomorrow {
			var tpID uuid.UUID
			if tp.ID != nil && *tp.ID != "" {
				parsedID, err := uuid.Parse(*tp.ID)
				if err != nil {
					return errors.New("invalid id in plan_tomorrow")
				}
				tpID = parsedID
			} else {
				tpID = uuid.New()
			}
			if _, ok := tpMap[tpID]; ok {
				tpUpdates = append(tpUpdates, map[string]interface{}{
					"id":          tpID,
					"description": tp.Description,
					"updated_at":  now,
				})
			} else {
				tpInserts = append(tpInserts, models.TomorrowPlans{
					ID:          tpID,
					Description: &tp.Description,
					ReportID:    &reportID,
					CreatedAt:   &now,
					UpdatedAt:   &now,
					Deleted:     func() *bool { b := false; return &b }(),
				})
			}
		}

		if len(tpInserts) > 0 {
			if err := tx.Create(&tpInserts).Error; err != nil {
				return err
			}
		}
		if len(tpUpdates) > 0 {
			tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"description", "updated_at"}),
			}).Create(&tpUpdates)
		}

		var helpInserts []models.HelpRequest
		var helpUpdates []map[string]interface{}
		for _, h := range req.Helps {
			var hID uuid.UUID
			if h.ID != nil && *h.ID != "" {
				parsedID, err := uuid.Parse(*h.ID)
				if err != nil {
					return errors.New("invalid id in help")
				}
				hID = parsedID
			} else {
				hID = uuid.New()
			}
			var helperUUID *uuid.UUID
			if h.HelperID != nil && *h.HelperID != "" {
				if parsed, err := uuid.Parse(*h.HelperID); err == nil {
					helperUUID = &parsed
				} else {
					return errors.New("invalid helper_id in help")
				}
			}
			if _, ok := helpMap[hID]; ok {
				helpUpdates = append(helpUpdates, map[string]interface{}{
					"id":          hID,
					"description": h.Description,
					"helper_id":   helperUUID,
					"status":      h.Status,
					"updated_at":  now,
				})
			} else {
				helpInserts = append(helpInserts, models.HelpRequest{
					ID:          hID,
					HelperID:    helperUUID,
					Description: &h.Description,
					Status:      h.Status,
					ReportID:    &reportID,
					CreatedAt:   &now,
					UpdatedAt:   &now,
					Deleted:     func() *bool { b := false; return &b }(),
				})
			}
		}

		if len(helpInserts) > 0 {
			if err := tx.Create(&helpInserts).Error; err != nil {
				return err
			}
		}
		if len(helpUpdates) > 0 {
			tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"description", "helper_id", "status", "updated_at"}),
			}).Create(&helpUpdates)
		}

		var rpInserts []models.ReportProblem
		for _, pIDstr := range req.Problems {
			if pIDstr == "" {
				continue
			}
			pid, err := uuid.Parse(pIDstr)
			if err != nil {
				return errors.New("invalid id in problems")
			}
			if _, ok := problemMap[pid]; !ok {
				rpInserts = append(rpInserts, models.ReportProblem{
					ReportID:  reportID,
					ProblemID: pid,
					CreatedAt: &now,
					UpdatedAt: &now,
					Deleted:   func() *bool { b := false; return &b }(),
				})
			}
		}

		if len(rpInserts) > 0 {
			if err := tx.Create(&rpInserts).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *reportRepository) DeleteReport(reportID uuid.UUID) (bool, error) {
	delTrue := true
	now := time.Now()
	update := map[string]interface{}{
		"deleted":    &delTrue,
		"updated_at": &now,
	}

	var affected int64
	txErr := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.DailyReport{}).
			Where("id = ?", reportID).
			Updates(update)
		if res.Error != nil {
			return res.Error
		}
		affected = res.RowsAffected

		if affected == 0 {
			return errors.New("report not found")
		}

		// связанные сущности по report_id
		for _, tbl := range []interface{}{
			&models.HelpRequest{}, &models.CompletedWork{}, &models.TomorrowPlans{}, &models.ReportProblem{},
		} {
			if res := tx.Session(&gorm.Session{}).
				Model(tbl).
				Where("report_id = ?", reportID).
				Updates(update); res.Error != nil {
				return res.Error
			}
		}
		return nil
	})

	if txErr != nil {
		if txErr.Error() == "report not found" {
			return false, nil
		}
		return false, txErr
	}

	return affected > 0, nil
}

func (r *reportRepository) UpdateHelpRequest(helpID uuid.UUID, updateData map[string]interface{}) (bool, error) {
	res := r.db.Session(&gorm.Session{}).
		Model(&models.HelpRequest{}).
		Where("id = ? AND deleted = FALSE", helpID).
		Updates(updateData)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *reportRepository) UpdateCompletedWork(cwID uuid.UUID, updateData map[string]interface{}) (bool, error) {
	res := r.db.Session(&gorm.Session{}).
		Model(&models.CompletedWork{}).
		Where("id = ? AND deleted = FALSE", cwID).
		Updates(updateData)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *reportRepository) UpdateTomorrowPlans(tpID uuid.UUID, updateData map[string]interface{}) (bool, error) {
	res := r.db.Session(&gorm.Session{}).
		Model(&models.TomorrowPlans{}).
		Where("id = ? AND deleted = FALSE", tpID).
		Updates(updateData)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *reportRepository) GetHelpRequestsForUser(userID uuid.UUID) ([]models.HelpRequest, error) {
	var helpRequests []models.HelpRequest
	if err := r.db.Session(&gorm.Session{}).
		Model(&models.HelpRequest{}).
		Preload("Report", "deleted = FALSE").
		Preload("Report.User", "deleted = FALSE").
		Where("deleted = false AND helper_id = ?", userID).
		Find(&helpRequests).Error; err != nil {
		return nil, err
	}
	return helpRequests, nil
}

func (r *reportRepository) DeleteHelpRequest(requestID uuid.UUID) (bool, error) {
	delTrue := true
	now := time.Now()
	update := map[string]interface{}{
		"deleted":    &delTrue,
		"updated_at": &now,
	}

	txErr := r.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.HelpRequest{}).
			Where("id = ?", requestID).
			Updates(update)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("help request not found")
		}
		return nil
	})

	if txErr != nil {
		if txErr.Error() == "help request not found" {
			return false, nil
		}
		return false, txErr
	}

	return true, nil
}

func (r *reportRepository) Transaction(txFunc func(ReportRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &reportRepository{db: tx}
		return txFunc(txRepo)
	})
}