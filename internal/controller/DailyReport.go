package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"emplacc-api/internal/utils"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func RegisterReportRoutes(e *echo.Echo) {
	reportGroup := e.Group("/report")
	reportGroup.Use(KeycloakAuthMiddleware)
	{
		reportGroup.GET("/all/:page/:pagesize", getAllReports)
		reportGroup.GET("/:id", getReport)
		reportGroup.GET("/task/:id", getReportsByTaskId)
		reportGroup.GET("/project/:id", getReportsByProjectId)
		reportGroup.POST("", createReport)
		reportGroup.PATCH("/:id", updateReport)
		reportGroup.DELETE("/:id", deleteReport)
		reportGroup.PATCH("/help-request/:id", updateHelpRequest)
		reportGroup.PATCH("/completed-work/:id", updateCompletedWork)
		reportGroup.PATCH("/tomorrow-plans/:id", updateTomorrowPlans)
	}
}

// getAllReports godoc
// @Summary Получение списка всех отчетов
// @Description Получает список всех отчетов с учетом пагинации, исключая удаленные
// @Tags Reports
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportListResponse "Список отчетов успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 404 {object} map[string]string "Пользователь, запрос на помощь, выполненная работа или план на завтра не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении отчетов"
// @Router /report/all/{page}/{pagesize} [get]
func getAllReports(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	page, _ := strconv.Atoi(c.Param("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Param("pagesize"))
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.DailyReport{}).
		Where("deleted = FALSE").
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count reports): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете отчетов"})
	}

	var reports []models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.DailyReport{}).
		Preload("User", "deleted = FALSE").
		Preload("HelpRequests", "deleted = FALSE").
		Preload("CompletedWork", "deleted = FALSE").
		Preload("TomorrowPlans", "deleted = FALSE").
		Preload("ReportProblems", "deleted = FALSE").
		Preload("ReportProblems.Problem", "deleted = FALSE").
		Where("deleted = FALSE").
		Order("report_date DESC NULLS LAST, created_at DESC NULLS LAST").
		Limit(pageSize).Offset(offset).
		Find(&reports).Error; err != nil {
		log.Printf("DB error (find reports): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении отчетов"})
	}

	out := response.ReportListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}
	for _, r := range reports {
		user := r.User
		userInfo := response.UserShort{
			ID:        r.UserID.String(),
			FirstName: utils.GetString(user.FirstName),
			LastName:  utils.GetString(user.LastName),
		}

		var complWork []response.CompletedWork
		for _, w := range r.CompletedWork {
			complWork = append(complWork, response.CompletedWork{
				ID:          w.ID.String(),
				Description: utils.GetString(w.Description),
			})
		}

		var plans []response.TomorrowPlans
		for _, p := range r.TomorrowPlans {
			plans = append(plans, response.TomorrowPlans{
				ID:          p.ID.String(),
				Description: utils.GetString(p.Description),
			})
		}

		var problemsResp []response.ProblemResponse
		for _, rp := range r.ReportProblems {
			if rp.Problem != nil {
				problemsResp = append(problemsResp, response.ProblemResponse{
					ID:          rp.Problem.ID.String(),
					Name:        utils.GetString(rp.Problem.Name),
					Description: rp.Problem.Description,
					CreatorId:   utils.GetUUIDString(rp.Problem.CreatorID),
					CreatedAt:   utils.GetTime(rp.Problem.CreatedAt),
					UpdatedAt:   utils.GetTime(rp.Problem.UpdatedAt),
				})
			}
		}

		helpResp := []response.HelpRequestItem{}
		for _, hr := range r.HelpRequests{
			helpResp = append(helpResp, response.HelpRequestItem{
				ID:          hr.ID.String(),
				HelperID:    utils.GetUUIDString(hr.HelperID),
				Description: utils.GetString(hr.Description),
				Status:      utils.GetString(hr.Status),
			})
		}

		out.Reports = append(out.Reports, response.ReportResponse{
			ID:            r.ID.String(),
			UserID:        r.UserID.String(),
			ReportDate:    utils.GetTime(r.ReportDate),
			CompletedWork: complWork,
			PlanTomorrow:  plans,
			HelpRequest:   helpResp,
			CreatedAt:     utils.GetTime(r.CreatedAt),
			UpdatedAt:     utils.GetTime(r.UpdatedAt),
			UserInfo:      userInfo,
			Problems:      problemsResp,
			Checked:       utils.GetInt8(r.Checked),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// getReport godoc
// @Summary Получение отчета по ID
// @Description Получает данные отчета по его уникальному идентификатору
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID отчета"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportResponse "Отчет успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор отчета"
// @Failure 404 {object} map[string]string "Отчет, пользователь, запрос на помощь, выполненная работа или план на завтра не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении отчета"
// @Router /report/{id} [get]
func getReport(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор отчета"})
	}

	var r models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.DailyReport{}).
		Preload("User", "deleted = FALSE").
		Preload("HelpRequest", "deleted = FALSE").
		Preload("CompletedWork", "deleted = FALSE").
		Preload("TomorrowPlans", "deleted = FALSE").
		Preload("ReportProblems", "deleted = FALSE").
		Preload("ReportProblems.Problem", "deleted = FALSE").
		Where("deleted = FALSE AND id = ?", reportID).
		First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Отчет не найден"})
		}
		log.Printf("DB error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении отчета"})
	}

	userInfo := response.UserShort{
		ID:        r.UserID.String(),
		FirstName: utils.GetString(r.User.FirstName),
		LastName:  utils.GetString(r.User.LastName),
	}

	var complWork []response.CompletedWork
	for _, w := range r.CompletedWork {
		complWork = append(complWork, response.CompletedWork{
			ID:          w.ID.String(),
			Description: utils.GetString(w.Description),
		})
	}

	var plans []response.TomorrowPlans
	for _, p := range r.TomorrowPlans {
		plans = append(plans, response.TomorrowPlans{
			ID:          p.ID.String(),
			Description: utils.GetString(p.Description),
		})
	}

	var problemsResp []response.ProblemResponse
	for _, rp := range r.ReportProblems {
		if rp.Problem != nil {
			problemsResp = append(problemsResp, response.ProblemResponse{
				ID:          rp.Problem.ID.String(),
				Name:        utils.GetString(rp.Problem.Name),
				Description: rp.Problem.Description,
				CreatorId:   utils.GetUUIDString(rp.Problem.CreatorID),
				CreatedAt:   utils.GetTime(rp.Problem.CreatedAt),
				UpdatedAt:   utils.GetTime(rp.Problem.UpdatedAt),
			})
		}
	}

	helpResp := []response.HelpRequestItem{}
	for _, hr := range r.HelpRequests{
		helpResp = append(helpResp, response.HelpRequestItem{
			ID:          hr.ID.String(),
			HelperID:    utils.GetUUIDString(hr.HelperID),
			Description: utils.GetString(hr.Description),
			Status:      utils.GetString(hr.Status),
		})
	}

	return c.JSON(http.StatusOK, response.ReportResponse{
		ID:            r.ID.String(),
		UserID:        r.UserID.String(),
		ReportDate:    utils.GetTime(r.ReportDate),
		CompletedWork: complWork,
		PlanTomorrow:  plans,
		HelpRequest:   helpResp,
		CreatedAt:     utils.GetTime(r.CreatedAt),
		UpdatedAt:     utils.GetTime(r.UpdatedAt),
		UserInfo:      userInfo,
		Problems:      problemsResp,
		Checked:       utils.GetInt8(r.Checked),
	})
}

// getReportsByTaskId godoc
// @Summary Получение отчетов по ID задачи
// @Description Получает список отчетов, связанных с указанной задачей
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportListByTaskId "Список отчетов успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор задачи"
// @Failure 404 {object} map[string]string "Пользователь, запрос на помощь, выполненная работа или план на завтра не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении отчетов"
// @Router /report/task/{id} [get]
func getReportsByTaskId(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID задачи"})
	}

	// находим все completed_work по задаче -> собираем report_id
	var cw []models.CompletedWork
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.CompletedWork{}).
		Where("deleted = FALSE AND task_id = ?", taskID).
		Find(&cw).Error; err != nil {
		log.Printf("DB error (find completed_work by task): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении работ по задаче"})
	}
	reportIDs := make(map[uuid.UUID]struct{})
	for _, w := range cw {
		if w.ReportID != nil {
			reportIDs[*w.ReportID] = struct{}{}
		}
	}
	if len(reportIDs) == 0 {
		return c.JSON(http.StatusOK, response.ReportListByTaskId{TaskID: taskID.String()})
	}

	ids := make([]uuid.UUID, 0, len(reportIDs))
	for id := range reportIDs {
		ids = append(ids, id)
	}

	var reports []models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.DailyReport{}).
		Preload("User", "deleted = FALSE").
		Preload("HelpRequest", "deleted = FALSE").
		Preload("CompletedWork", "deleted = FALSE").
		Preload("TomorrowPlans", "deleted = FALSE").
		Preload("ReportProblems", "deleted = FALSE").
		Preload("ReportProblems.Problem", "deleted = FALSE").
		Where("deleted = FALSE AND id IN ?", ids).
		Find(&reports).Error; err != nil {
		log.Printf("DB error (find reports by task): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении отчетов"})
	}

	out := response.ReportListByTaskId{TaskID: taskID.String()}
	for _, r := range reports {
		userInfo := response.UserShort{
			ID:        r.UserID.String(),
			FirstName: utils.GetString(r.User.FirstName),
			LastName:  utils.GetString(r.User.LastName),
		}

		var complWork []response.CompletedWork
		for _, w := range r.CompletedWork {
			complWork = append(complWork, response.CompletedWork{
				ID:          w.ID.String(),
				Description: utils.GetString(w.Description),
			})
		}

		var plans []response.TomorrowPlans
		for _, p := range r.TomorrowPlans {
			plans = append(plans, response.TomorrowPlans{
				ID:          p.ID.String(),
				Description: utils.GetString(p.Description),
			})
		}

		helpResp := []response.HelpRequestItem{}
		for _, hr := range r.HelpRequests{
			helpResp = append(helpResp, response.HelpRequestItem{
				ID:          hr.ID.String(),
				HelperID:    utils.GetUUIDString(hr.HelperID),
				Description: utils.GetString(hr.Description),
				Status:      utils.GetString(hr.Status),
			})
		}

		var problemsResp []response.ProblemResponse
		for _, rp := range r.ReportProblems {
			if rp.Problem != nil {
				problemsResp = append(problemsResp, response.ProblemResponse{
					ID:          rp.Problem.ID.String(),
					Name:        utils.GetString(rp.Problem.Name),
					Description: rp.Problem.Description,
					CreatorId:   utils.GetUUIDString(rp.Problem.CreatorID),
					CreatedAt:   utils.GetTime(rp.Problem.CreatedAt),
					UpdatedAt:   utils.GetTime(rp.Problem.UpdatedAt),
				})
			}
		}

		out.Reports = append(out.Reports, response.ReportResponse{
			ID:            r.ID.String(),
			UserID:        r.UserID.String(),
			ReportDate:    utils.GetTime(r.ReportDate),
			CompletedWork: complWork,
			PlanTomorrow:  plans,
			HelpRequest:   helpResp,
			CreatedAt:     utils.GetTime(r.CreatedAt),
			UpdatedAt:     utils.GetTime(r.UpdatedAt),
			UserInfo:      userInfo,
			Problems:      problemsResp,
			Checked:       utils.GetInt8(r.Checked),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// getReportsByProjectId godoc
// @Summary Получение отчетов по ID проекта
// @Description Получает список отчетов, связанных с задачами указанного проекта
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID проекта"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportListByProjectId "Список отчетов успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проекта"
// @Failure 404 {object} map[string]string "Пользователь, запрос на помощь, выполненная работа или план на завтра не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении отчетов"
// @Router /report/project/{id} [get]
func getReportsByProjectId(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	projectUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
	}

	var tasks []models.Task
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.Task{}).
		Where("project_id = ? AND deleted = FALSE", projectUUID).
		Find(&tasks).Error; err != nil {
		log.Printf("DB error (find tasks by project id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задач"})
	}

	if len(tasks) == 0 {
		return c.JSON(http.StatusOK, response.ReportListByProjectId{ProjectID: projectUUID.String()})
	}

	taskIDs := make([]uuid.UUID, 0, len(tasks))
	for _, t := range tasks {
		taskIDs = append(taskIDs, t.ID)
	}

	// по задачам проекта берём completed_work, вытягиваем report_id
	var cw []models.CompletedWork
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.CompletedWork{}).
		Where("deleted = FALSE AND task_id IN ?", taskIDs).
		Find(&cw).Error; err != nil {
		log.Printf("DB error (find completed_work by tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении работ"})
	}
	reportIDs := make(map[uuid.UUID]struct{})
	for _, w := range cw {
		if w.ReportID != nil {
			reportIDs[*w.ReportID] = struct{}{}
		}
	}
	if len(reportIDs) == 0 {
		return c.JSON(http.StatusOK, response.ReportListByProjectId{ProjectID: projectUUID.String()})
	}
	ids := make([]uuid.UUID, 0, len(reportIDs))
	for id := range reportIDs {
		ids = append(ids, id)
	}

	var reports []models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.DailyReport{}).
		Preload("User", "deleted = FALSE").
		Preload("HelpRequest", "deleted = FALSE").
		Preload("CompletedWork", "deleted = FALSE").
		Preload("TomorrowPlans", "deleted = FALSE").
		Preload("ReportProblems", "deleted = FALSE").
		Preload("ReportProblems.Problem", "deleted = FALSE").
		Where("deleted = FALSE AND id IN ?", ids).
		Find(&reports).Error; err != nil {
		log.Printf("DB error (find reports by project id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении отчетов"})
	}

	out := response.ReportListByProjectId{ProjectID: projectUUID.String()}
	for _, r := range reports {
		userInfo := response.UserShort{
			ID:        r.UserID.String(),
			FirstName: utils.GetString(r.User.FirstName),
			LastName:  utils.GetString(r.User.LastName),
		}

		var complWork []response.CompletedWork
		for _, w := range r.CompletedWork {
			complWork = append(complWork, response.CompletedWork{
				ID:          w.ID.String(),
				Description: utils.GetString(w.Description),
			})
		}

		var plans []response.TomorrowPlans
		for _, p := range r.TomorrowPlans {
			plans = append(plans, response.TomorrowPlans{
				ID:          p.ID.String(),
				Description: utils.GetString(p.Description),
			})
		}

		helpResp := []response.HelpRequestItem{}
		for _, hr := range r.HelpRequests{
			helpResp = append(helpResp, response.HelpRequestItem{
				ID:          hr.ID.String(),
				HelperID:    utils.GetUUIDString(hr.HelperID),
				Description: utils.GetString(hr.Description),
				Status:      utils.GetString(hr.Status),
			})
		}

		var problemsResp []response.ProblemResponse
		for _, rp := range r.ReportProblems {
			if rp.Problem != nil {
				problemsResp = append(problemsResp, response.ProblemResponse{
					ID:          rp.Problem.ID.String(),
					Name:        utils.GetString(rp.Problem.Name),
					Description: rp.Problem.Description,
					CreatorId:   utils.GetUUIDString(rp.Problem.CreatorID),
					CreatedAt:   utils.GetTime(rp.Problem.CreatedAt),
					UpdatedAt:   utils.GetTime(rp.Problem.UpdatedAt),
				})
			}
		}

		out.Reports = append(out.Reports, response.ReportResponse{
			ID:            r.ID.String(),
			UserID:        r.UserID.String(),
			ReportDate:    utils.GetTime(r.ReportDate),
			CompletedWork: complWork,
			PlanTomorrow:  plans,
			HelpRequest:   helpResp,
			CreatedAt:     utils.GetTime(r.CreatedAt),
			UpdatedAt:     utils.GetTime(r.UpdatedAt),
			UserInfo:      userInfo,
			Problems:      problemsResp,
			Checked:       utils.GetInt8(r.Checked),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// createReport godoc
// @Summary Создание нового отчета
// @Description Создает новый отчет с указанными параметрами, включая выполненную работу, планы на завтра, проблемы и запрос на помощь
// @Tags Reports
// @Accept json
// @Produce json
// @Param report body request.ReportCreateRequest true "Данные для создания отчета"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.ReportUniversalResponse "Отчет успешно создан"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании отчета"
// @Router /report [post]
func createReport(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	var req request.ReportCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("ParseUserId error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось распарсить userId"})
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

	txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		// создаём сам отчёт
		if res := tx.Session(&gorm.Session{}).Model(&models.DailyReport{}).Create(&rep); res.Error != nil {
			return res.Error
		}

		// CompletedWork: создаём столько, сколько пришло
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
				// task_id опционально
				if w.TaskID != nil {
					if tid, err := uuid.Parse(*w.TaskID); err == nil {
						item.TaskID = &tid
					}
				}
				batch = append(batch, item)
			}
			if res := tx.Session(&gorm.Session{}).Model(&models.CompletedWork{}).Create(&batch); res.Error != nil {
				return res.Error
			}
		}

		// TomorrowPlans
		if len(req.PlanTomorrow) > 0 {
			batch := make([]models.TomorrowPlans, 0, len(req.PlanTomorrow))
			for _, p := range req.PlanTomorrow {
				batch = append(batch, models.TomorrowPlans{
					ID:        uuid.New(),
					Description: &p.Description,
					ReportID:  &rep.ID,
					Deleted:   &delFalse,
					CreatedAt: &now,
				})
			}
			if res := tx.Session(&gorm.Session{}).Model(&models.TomorrowPlans{}).Create(&batch); res.Error != nil {
				return res.Error
			}
		}

		// Problems (pivot)
		if len(req.Problems) > 0 {
			batch := make([]models.ReportProblem, 0, len(req.Problems))
			for _, pr := range req.Problems {
				pid, err := uuid.Parse(pr)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "Некорректный problemId")
				}
				batch = append(batch, models.ReportProblem{
					ReportID:  rep.ID,
					ProblemID: pid,
					Deleted:   &delFalse,
					CreatedAt: &now,
				})
			}
			if res := tx.Session(&gorm.Session{}).Model(&models.ReportProblem{}).Create(&batch); res.Error != nil {
				return res.Error
			}
		}

		// HelpRequest: поддержка массива Helps и одиночного Help
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
					return echo.NewHTTPError(http.StatusBadRequest, "Некорректный helperId")
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
			if res := tx.Session(&gorm.Session{}).Model(&models.HelpRequest{}).Create(&batch); res.Error != nil {
				return res.Error
			}
		}

		return nil
	})
	if txErr != nil {
		log.Printf("DB transaction error (create report): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании отчета"})
	}

	return c.JSON(http.StatusCreated, response.ReportUniversalResponse{
		ID:      rep.ID.String(),
		Message: "Успешно создано",
	})
}

// updateReport godoc
// @Summary Обновление отчета и связанных данных, ЕСЛИ НЕ УКАЗЫВАТЬ ID ВО ВСПОМОГАТЕЛЬНЫХ СУЩНОСТЯХ, СОЗДАЕТ НОВЫЕ
// @Tags Reports
// @Description Обновляет поля отчета, а также связанные CompletedWork, TomorrowPlans, HelpRequests и ReportProblems, ЕСЛИ НЕ УКАЗЫВАТЬ ID ВО ВСПОМОГАТЕЛЬНЫХ СУЩНОСТЯХ, СОЗДАЕТ НОВЫЕ
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID отчета"
// @Param report body request.ReportReplaceRequest true "Данные для обновления отчета и связанных сущностей"
// @Security BearerAuth
// @Success 200 {object} response.ReportUniversalResponse "Отчет успешно обновлен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор отчета или ошибка в запросе"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Отчет не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении отчета"
// @Router /report/{id} [patch]
func updateReport(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор отчета"})
	}

	var req request.ReportReplaceRequest
	if err = c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	now := time.Now()

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		var report models.DailyReport
		if err := tx.Where("id = ? AND deleted = FALSE", reportID).First(&report).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Отчет не найден"})
			}
			return err
		}

		updateData := map[string]interface{}{"updated_at": &now}
		if req.UserId != "" {
			if uid, err := uuid.Parse(req.UserId); err == nil {
				updateData["user_id"] = uid
			} else {
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "Некорректный user_id"})
			}
		}
		if req.ReportDate != nil {
			updateData["report_date"] = req.ReportDate
		}
		if req.Checked != nil {
			updateData["checked"] = req.Checked
		}
		if err := tx.Model(&report).Updates(updateData).Error; err != nil {
			return err
		}

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
				cwID, err = uuid.Parse(*cw.ID)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "Некорректный id в completed_work"})
				}
			} else {
				cwID = uuid.New()
			}

			var taskUUID *uuid.UUID
			if cw.TaskID != nil && *cw.TaskID != "" {
				if parsed, err := uuid.Parse(*cw.TaskID); err == nil {
					taskUUID = &parsed
				} else {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "Некорректный task_id в completed_work"})
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
				tpID, err = uuid.Parse(*tp.ID)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "Некорректный id в plan_tomorrow"})
				}
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
				hID, err = uuid.Parse(*h.ID)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "Некорректный id в help"})
				}
			} else {
				hID = uuid.New()
			}

			var helperUUID *uuid.UUID
			if h.HelperID != nil && *h.HelperID != "" {
				if parsed, err := uuid.Parse(*h.HelperID); err == nil {
					helperUUID = &parsed
				} else {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "Некорректный helper_id в help"})
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
				return echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": "Некорректный id в problems"})
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
	}); txErr != nil {
		log.Printf("Transaction error (update report and relations): %v", txErr)
		if httpErr, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(httpErr.Code, httpErr.Message)
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении отчета"})
	}

	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      reportID.String(),
		Message: "Отчет успешно изменен",
	})
}

// deleteReport godoc
// @Summary Удаление отчета
// @Description Логическое удаление отчета по ID, включая связанные данные (поле deleted = true)
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID отчета"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportUniversalResponse "Отчет успешно удален"
// @Failure 404 {object} map[string]string "Отчет не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении отчета"
// @Router /report/{id} [delete]
func deleteReport(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор отчета"})
	}

	delTrue := true
	now := time.Now()
	update := map[string]interface{}{
		"deleted":    &delTrue,
		"updated_at": &now,
	}

	txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.DailyReport{}).
			Where("id = ?", reportID).
			Updates(update)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
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
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении отчета"})
	}

	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      reportID.String(),
		Message: "Отчет успешно удален",
	})
}

// updateHelpRequest godoc
// @Summary Обновление запроса на помощь
// @Description Обновляет данные запроса на помощь по его ID
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID запроса на помощь"
// @Param helpRequest body request.HelpRequestUpdateRequest true "Данные для обновления запроса на помощь"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportUniversalResponse "Запрос на помощь успешно обновлен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении запроса на помощь"
// @Router /report/help-request/{id} [patch]
func updateHelpRequest(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	helpID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор запроса помощи"})
	}

	var req request.HelpRequestUpdateRequest
	if err = c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	updateData := map[string]interface{}{}
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
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
	}
	now := time.Now()
	updateData["updated_at"] = &now

	res := dbConn.Session(&gorm.Session{}).
		Model(&models.HelpRequest{}).
		Where("id = ? AND deleted = FALSE", helpID).
		Updates(updateData)
	if res.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении запроса на помощь"})
	}
	if res.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
	}
	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      helpID.String(),
		Message: "Запрос на помощь обновлен",
	})
}

// updateCompletedWork godoc
// @Summary Обновление выполненной работы
// @Description Обновляет данные выполненной работы по ее ID
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID выполненной работы"
// @Param completedWork body request.CompletedWorkUpdateRequest true "Данные для обновления выполненной работы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportUniversalResponse "Выполненная работа успешно обновлена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении выполненной работы"
// @Router /report/completed-work/{id} [patch]
func updateCompletedWork(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	cwID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор выполненной работы"})
	}

	var req request.CompletedWorkUpdateRequest
	if err = c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	updateData := map[string]interface{}{}
	if req.Description != nil {
		updateData["description"] = *req.Description
	}
	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
	}
	now := time.Now()
	updateData["updated_at"] = &now

	res := dbConn.Session(&gorm.Session{}).
		Model(&models.CompletedWork{}).
		Where("id = ? AND deleted = FALSE", cwID).
		Updates(updateData)
	if res.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении проделанной работы"})
	}
	if res.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
	}
	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      cwID.String(),
		Message: "Выполненная работа обновлена",
	})
}

// updateTomorrowPlans godoc
// @Summary Обновление планов на завтра
// @Description Обновляет данные планов на завтра по их ID
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID планов на завтра"
// @Param tomorrowPlans body request.TomorrowPlansUpdateRequest true "Данные для обновления планов на завтра"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportUniversalResponse "Планы на завтра успешно обновлены"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении планов на завтра"
// @Router /report/tomorrow-plans/{id} [patch]
func updateTomorrowPlans(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	tpID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор планов на завтра"})
	}

	var req request.TomorrowPlansUpdateRequest
	if err = c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	updateData := map[string]interface{}{}
	if req.Description != nil {
		updateData["description"] = *req.Description
	}
	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
	}
	now := time.Now()
	updateData["updated_at"] = &now

	res := dbConn.Session(&gorm.Session{}).
		Model(&models.TomorrowPlans{}).
		Where("id = ? AND deleted = FALSE", tpID).
		Updates(updateData)
	if res.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении планов на завтра"})
	}
	if res.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
	}
	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      tpID.String(),
		Message: "Планы на завтра обновлены",
	})
}
