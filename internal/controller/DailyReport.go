package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	utils "emplacc-api/internal/utils"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
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

	// Пагинация
	page, _ := strconv.Atoi(c.Param("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Param("pagesize"))
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// Считаем общее количество отчетов
	var totalCount int64
	if err := dbConn.Model(&models.DailyReport{}).Session(&gorm.Session{}).
		Where("deleted = ?", false).
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count reports): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете отчетов",
		})
	}

	// Загружаем отчеты с preloaded связями
	var reports []models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).Model(models.DailyReport{}).
		Preload("User", "deleted = ?", false).
		Preload("Task", "deleted = ?", false).
		Preload("HelpRequest", "deleted = ?", false).
		Preload("CompletedWork", "deleted = ?", false).
		Preload("TomorrowPlans", "deleted = ?", false).
		Preload("ReportProblems", "deleted = ?", false).
		Preload("ReportProblems.Problem", "deleted = ?", false).
		Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&reports).Error; err != nil {
		log.Printf("DB error (find reports): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении отчетов из базы данных",
		})
	}

	// Формируем ответ
	reportList := response.ReportListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, report := range reports {
		if report.Deleted != nil && *report.Deleted {
			continue
		}

		// Пользователь
		user := report.User
		userInfo := response.UserShort{
			ID:        report.UserID.String(),
			FirstName: utils.GetString(user.FirstName),
			LastName:  utils.GetString(user.LastName),
		}

		// Выполненная работа
		var complWork []response.WorkItem
		for _, workItem := range report.CompletedWork {
			complWork = append(complWork, response.WorkItem{
				ID:          workItem.ID.String(),
				Description: utils.GetString(workItem.Description),
			})
		}

		// План на завтра
		var tomorrowPlans []response.WorkItem
		for _, plan := range report.TomorrowPlans {
			tomorrowPlans = append(tomorrowPlans, response.WorkItem{
				ID:          plan.ID.String(),
				Description: utils.GetString(plan.Description),
			})
		}

		// Запрос на помощь
		var helpRequest response.HelpRequestItem
		if report.HelpRequest != nil {
			helpRequest = response.HelpRequestItem{
				ID:          report.HelpRequest.ID.String(),
				HelperID:    utils.GetUUIDString(report.HelpRequest.HelperID),
				Description: utils.GetString(report.HelpRequest.Description),
			}
		}

		// Проблемы
		var problemsResp []response.ProblemResponse
		for _, rp := range report.ReportProblems {
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

		resp := response.ReportResponse{
			ID:            report.ID.String(),
			UserID:        report.UserID.String(),
			ReportDate:    utils.GetTime(report.ReportDate),
			CompletedWork: complWork,
			PlanTomorrow:  tomorrowPlans,
			HelpRequest:   helpRequest,
			TaskId:        utils.GetUUIDString(report.TaskID),
			CreatedAt:     utils.GetTime(report.CreatedAt),
			UpdatedAt:     utils.GetTime(report.UpdatedAt),
			UserInfo:      userInfo,
			Problems:      problemsResp,
		}

		reportList.Reports = append(reportList.Reports, resp)
	}

	return c.JSON(http.StatusOK, reportList)
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

	// Парсим UUID отчета
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор отчета",
		})
	}

	// Загружаем отчет с предзагрузкой всех связанных сущностей
	var report models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).Model(models.DailyReport{}).
		Preload("User", "deleted = ?", false).
		Preload("Task", "deleted = ?", false).
		Preload("HelpRequest", "deleted = ?", false).
		Preload("CompletedWork", "deleted = ?", false).
		Preload("TomorrowPlans", "deleted = ?", false).
		Preload("ReportProblems", "deleted = ?", false).
		Preload("ReportProblems.Problem", "deleted = ?", false).
		Where("deleted = ? AND id = ?", false, reportID).
		First(&report).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Отчет не найден"})
		}
		log.Printf("DB error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении отчета"})
	}

	// Пользователь
	user := report.User
	userInfo := response.UserShort{
		ID:        report.UserID.String(),
		FirstName: utils.GetString(user.FirstName),
		LastName:  utils.GetString(user.LastName),
	}

	// Выполненная работа
	var complWork []response.WorkItem
	for _, work := range report.CompletedWork {
		complWork = append(complWork, response.WorkItem{
			ID:          work.ID.String(),
			Description: utils.GetString(work.Description),
		})
	}

	// План на завтра
	var tomorrowPlans []response.WorkItem
	for _, plan := range report.TomorrowPlans {
		tomorrowPlans = append(tomorrowPlans, response.WorkItem{
			ID:          plan.ID.String(),
			Description: utils.GetString(plan.Description),
		})
	}

	// Запрос на помощь
	var helpRequest response.HelpRequestItem
	if report.HelpRequest != nil {
		helpRequest = response.HelpRequestItem{
			ID:          report.HelpRequest.ID.String(),
			HelperID:    utils.GetUUIDString(report.HelpRequest.HelperID),
			Description: utils.GetString(report.HelpRequest.Description),
		}
	}

	// Проблемы
	var problemsResp []response.ProblemResponse
	for _, rp := range report.ReportProblems {
		if rp.Problem != nil {
			problemsResp = append(problemsResp, response.ProblemResponse{
				ID:          rp.Problem.ID.String(),
				Name:        utils.GetString(rp.Problem.Name),
				Description: rp.Problem.Description, // pq.StringArray уже []string
				CreatorId:   utils.GetUUIDString(rp.Problem.CreatorID),
				CreatedAt:   utils.GetTime(rp.Problem.CreatedAt),
				UpdatedAt:   utils.GetTime(rp.Problem.UpdatedAt),
			})
		}
	}

	// TaskID (nullable)
	taskID := utils.GetUUIDString(report.TaskID)

	// Формируем ответ
	resp := response.ReportResponse{
		ID:            report.ID.String(),
		UserID:        report.UserID.String(),
		ReportDate:    utils.GetTime(report.ReportDate),
		CompletedWork: complWork,
		PlanTomorrow:  tomorrowPlans,
		HelpRequest:   helpRequest,
		TaskId:        taskID,
		CreatedAt:     utils.GetTime(report.CreatedAt),
		UpdatedAt:     utils.GetTime(report.UpdatedAt),
		UserInfo:      userInfo,
		Problems:      problemsResp,
	}

	return c.JSON(http.StatusOK, resp)
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

	taskIDParam := c.Param("id")
	taskUUID, err := uuid.Parse(taskIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID задачи"})
	}

	var reports []models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).Model(models.DailyReport{}).
		Preload("User", "deleted = ?", false).
		Preload("HelpRequest", "deleted = ?", false).
		Preload("CompletedWork", "deleted = ?", false).
		Preload("TomorrowPlans", "deleted = ?", false).
		Preload("ReportProblems", "deleted = ?", false).
		Preload("ReportProblems.Problem", "deleted = ?", false).
		Where("deleted = ? AND task_id = ?", false, taskUUID).
		Find(&reports).Error; err != nil {
		log.Printf("DB error (find reports): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении отчетов"})
	}

	reportList := response.ReportListByTaskId{
		TaskID: taskIDParam,
	}

	for _, report := range reports {
		if report.Deleted != nil && *report.Deleted {
			continue
		}

		userInfo := response.UserShort{
			ID:        report.UserID.String(),
			FirstName: utils.GetString(report.User.FirstName),
			LastName:  utils.GetString(report.User.LastName),
		}

		// Выполненная работа
		var complWork []response.WorkItem
		for _, work := range report.CompletedWork {
			complWork = append(complWork, response.WorkItem{
				ID:          work.ID.String(),
				Description: utils.GetString(work.Description),
			})
		}

		// План на завтра
		var tomorrowPlans []response.WorkItem
		for _, plan := range report.TomorrowPlans {
			tomorrowPlans = append(tomorrowPlans, response.WorkItem{
				ID:          plan.ID.String(),
				Description: utils.GetString(plan.Description),
			})
		}

		// Запрос на помощь
		var helpRequest response.HelpRequestItem
		if report.HelpRequest != nil {
			helpRequest = response.HelpRequestItem{
				ID:          report.HelpRequest.ID.String(),
				HelperID:    utils.GetUUIDString(report.HelpRequest.HelperID),
				Description: utils.GetString(report.HelpRequest.Description),
			}
		}

		// Проблемы
		var problemsResp []response.ProblemResponse
		for _, rp := range report.ReportProblems {
			if rp.Problem != nil {
				problemsResp = append(problemsResp, response.ProblemResponse{
					ID:          rp.Problem.ID.String(),
					Name:        utils.GetString(rp.Problem.Name),
					Description: rp.Problem.Description, // pq.StringArray уже []string
					CreatorId:   utils.GetUUIDString(rp.Problem.CreatorID),
					CreatedAt:   utils.GetTime(rp.Problem.CreatedAt),
					UpdatedAt:   utils.GetTime(rp.Problem.UpdatedAt),
				})
			}
		}

		taskID := utils.GetUUIDString(report.TaskID)

		resp := response.ReportResponse{
			ID:            report.ID.String(),
			UserID:        report.UserID.String(),
			ReportDate:    utils.GetTime(report.ReportDate),
			CompletedWork: complWork,
			PlanTomorrow:  tomorrowPlans,
			HelpRequest:   helpRequest,
			TaskId:        taskID,
			CreatedAt:     utils.GetTime(report.CreatedAt),
			UpdatedAt:     utils.GetTime(report.UpdatedAt),
			UserInfo:      userInfo,
			Problems:      problemsResp,
		}

		reportList.Reports = append(reportList.Reports, resp)
	}

	return c.JSON(http.StatusOK, reportList)
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

	projectID := c.Param("id")
	var tasks []models.Task
	if err := dbConn.Session(&gorm.Session{}).Model(models.Task{}).Where("project_id = ? AND deleted = ?", projectID, false).Find(&tasks).Error; err != nil {
		log.Printf("DB error (find tasks by project id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задач"})
	}

	var taskIDs []uuid.UUID
	for _, t := range tasks {
		taskIDs = append(taskIDs, t.ID)
	}

	var reports []models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).Model(models.DailyReport{}).
		Preload("User", "deleted = ?", false).
		Preload("HelpRequest", "deleted = ?", false).
		Preload("CompletedWork", "deleted = ?", false).
		Preload("TomorrowPlans", "deleted = ?", false).
		Preload("ReportProblems", "deleted = ?", false).
		Preload("ReportProblems.Problem", "deleted = ?", false).
		Where("task_id IN (?)", taskIDs).
		Find(&reports).Error; err != nil {
		log.Printf("DB error (find reports by project id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении отчетов"})
	}

	reportList := response.ReportListByProjectId{ProjectID: projectID}

	for _, report := range reports {
		if report.Deleted != nil && *report.Deleted {
			continue
		}

		userInfo := response.UserShort{
			ID:        report.UserID.String(),
			FirstName: utils.GetString(report.User.FirstName),
			LastName:  utils.GetString(report.User.LastName),
		}

		var complWork []response.WorkItem
		for _, w := range report.CompletedWork {
			complWork = append(complWork, response.WorkItem{
				ID:          w.ID.String(),
				Description: utils.GetString(w.Description),
			})
		}

		var tomorrowPlans []response.WorkItem
		for _, t := range report.TomorrowPlans {
			tomorrowPlans = append(tomorrowPlans, response.WorkItem{
				ID:          t.ID.String(),
				Description: utils.GetString(t.Description),
			})
		}

		var helpRequest response.HelpRequestItem
		if report.HelpRequest != nil {
			helpRequest = response.HelpRequestItem{
				ID:          report.HelpRequest.ID.String(),
				HelperID:    utils.GetUUIDString(report.HelpRequest.HelperID),
				Description: utils.GetString(report.HelpRequest.Description),
			}
		}

		var problemsResp []response.ProblemResponse
		for _, rp := range report.ReportProblems {
			if rp.Problem != nil {
				problemsResp = append(problemsResp, response.ProblemResponse{
					ID:          rp.Problem.ID.String(),
					Name:        utils.GetString(rp.Problem.Name),
					Description: rp.Problem.Description, // уже []string
					CreatorId:   utils.GetUUIDString(rp.Problem.CreatorID),
					CreatedAt:   utils.GetTime(rp.Problem.CreatedAt),
					UpdatedAt:   utils.GetTime(rp.Problem.UpdatedAt),
				})
			}
		}

		reportList.Reports = append(reportList.Reports, response.ReportResponse{
			ID:            report.ID.String(),
			UserID:        report.UserID.String(),
			ReportDate:    utils.GetTime(report.ReportDate),
			CompletedWork: complWork,
			PlanTomorrow:  tomorrowPlans,
			HelpRequest:   helpRequest,
			TaskId:        utils.GetUUIDString(report.TaskID),
			CreatedAt:     utils.GetTime(report.CreatedAt),
			UpdatedAt:     utils.GetTime(report.UpdatedAt),
			UserInfo:      userInfo,
			Problems:      problemsResp,
		})
	}

	return c.JSON(http.StatusOK, reportList)
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
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	now := time.Now()

	var userId uuid.UUID
	userId, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("ParseUserId error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось распарсить userId ",
		})
	}

	var taskId uuid.UUID
	taskId, err = uuid.Parse(req.TaskId)
	if err != nil {
		log.Printf("ParseTaskId error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось распарсить taskId ",
		})
	}

	del := false

	report := models.DailyReport{
		ID:         newUUID,
		UserID:     userId,
		TaskID:     &taskId,
		ReportDate: req.ReportDate,
		CreatedAt:  &now,
		Deleted: &del,
	}

	result := dbConn.Session(&gorm.Session{}).Model(models.DailyReport{}).Create(&report)
	if result.Error != nil {
		log.Printf("DB error (create report): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании отчета",
		})
	}

	if req.CompleteWork != nil {
		for _, completedWorks := range *req.CompleteWork {
			tempUUID := uuid.New()

		
			complWork := models.CompletedWork{
				ID:          tempUUID,
				Description: &completedWorks.Description,
				ReportID:    &newUUID,
				Deleted: &del,
			}

			result = dbConn.Session(&gorm.Session{}).Model(models.CompletedWork{}).Create(&complWork)
			if result.Error != nil {
				log.Printf("DB error (create report): %v", result.Error)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Ошибка при создании выполненной работы",
				})
			}
		}
	}

	if req.PlanTomorrow != nil {
		for _, tomorrowPlan := range *req.PlanTomorrow {
			tempUUID := uuid.New()

			tomPlan := models.TomorrowPlans{
				ID:          tempUUID,
				Description: &tomorrowPlan.Description,
				ReportID:    &newUUID,
				Deleted: &del,
			}

			result = dbConn.Session(&gorm.Session{}).Model(models.TomorrowPlans{}).Create(&tomPlan)
			if result.Error != nil {
				log.Printf("DB error (create report): %v", result.Error)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Ошибка при создании плана на завтра",
				})
			}
		}
	}

	if req.Problems != nil {
		for _, problem := range *req.Problems {
			problemId, err := uuid.Parse(problem.ID)
			if err != nil {
				log.Printf("ParseProblemId error: %v", err)
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "Ошибка при парсинге problemId",
				})
			}

			reportProblem := models.ReportProblem{
				ReportID:  newUUID,
				ProblemID: problemId,
				Deleted: &del,
			}

			result = dbConn.Session(&gorm.Session{}).Model(models.ReportProblem{}).Create(&reportProblem)
			if result.Error != nil {
				log.Printf("DB error (create report): %v", result.Error)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Ошибка при создании связи отчет-проблема",
				})
			}
		}
	}

	if req.Help != nil {
		tempUUID := uuid.New()

		helperId, err := uuid.Parse(req.Help.HelperID)
		if err != nil {
			log.Printf("ParseHelperId error: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Ошибка при парсинге helperId",
			})
		}

		help := models.HelpRequest{
			ID:          tempUUID,
			HelperID:    &helperId,
			Description: &req.Help.Description,
			ReportID:    &newUUID,
			Deleted: &del,
			CreatedAt: &now,
		}

		result = dbConn.Session(&gorm.Session{}).Model(models.HelpRequest{}).Create(&help)
		if result.Error != nil {
			log.Printf("DB error (create report): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при создании запроса на помощь",
			})
		}
	}

	resp := response.ReportUniversalResponse{
		ID:      newUUID.String(),
		Message: "Успешно создано",
	}

	return c.JSON(http.StatusCreated, resp)
}

// updateReport godoc
// @Summary Обновление отчета
// @Description Обновляет данные отчета по его ID
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID отчета"
// @Param report body request.ReportUpdateRequest true "Данные для обновления отчета"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportUniversalResponse "Отчет успешно обновлен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор отчета или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении отчета"
// @Router /report/{id} [patch]
func updateReport(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	reportId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор отчета",
		})
	}

	var req request.ReportUpdateRequest 
	if err = c.Bind(&req); err != nil{
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.UserId != nil{
		updateData["user_id"] = *req.UserId
	}
	if req.ReportDate != nil{
		updateData["report_date"] = *req.ReportDate
	}
	updateData["updated_at"] = time.Now()

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.DailyReport{}).Where("id = ? and deleted = ?", reportId, false).Updates(updateData)
		if res.Error != nil{
			log.Printf("DB error (update report): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0{
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil{
		log.Printf("Transaction error (update report): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении отчета",
		})
	}

	updateResponse := response.ReportUniversalResponse{
		ID: id,
		Message: "Отчет успешно изменен",
	}

	return c.JSON(http.StatusOK, updateResponse)
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
	id := c.Param("id")
	updateData := map[string]interface{}{"deleted": true}

	txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).Model(&models.DailyReport{}).Where("id = ?", id).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete report): %v", res.Error)
			return res.Error
		} else if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}

		tables := []interface{}{
			&models.HelpRequest{},
			&models.CompletedWork{},
			&models.TomorrowPlans{},
		}

		for _, table := range tables {
			if res := tx.Session(&gorm.Session{}).Model(table).Where("report_id = ?", id).Updates(updateData); res.Error != nil {
				log.Printf("DB error (delete related table %T): %v", table, res.Error)
				return res.Error
			}
		}

		var reportProblems []models.ReportProblem
		if err := tx.Session(&gorm.Session{}).Model(models.ReportProblem{}).Where("report_id = ? AND deleted = ?", id, false).Find(&reportProblems).Error; err != nil {
			log.Printf("DB error (get report-problem): %v", err)
			return err
		}

		if res := tx.Session(&gorm.Session{}).Model(&models.ReportProblem{}).Where("report_id = ?", id).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete report-problem): %v", res.Error)
			return res.Error
		}

		return nil
	})

	if txErr != nil {
		if httpErr, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(httpErr.Code, httpErr.Message)
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении отчета"})
	}

	resp := response.ReportUniversalResponse{
		ID:      id,
		Message: "Отчет успешно удален",
	}

	return c.JSON(http.StatusOK, resp)
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
func updateHelpRequest(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	helpRequestId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор запроса помощи",
		})
	}

	var req request.HelpRequestUpdateRequest
	if err = c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.Description != nil{
		updateData["description"] = *req.Description
	}
	if req.HelperID != nil{
		updateData["helper_id"] = *req.HelperID
	}

	if len(updateData) == 0{
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	result := dbConn.Session(&gorm.Session{}).Model(models.HelpRequest{}).Where("id = ?", helpRequestId).Updates(updateData)
	if result.Error != nil{
		log.Printf("DB error (update help request): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении запроса на помощь",
		})
	}

	if result.RowsAffected == 0{
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не обновлено",
		})
	}

	updateResponse := response.ReportUniversalResponse{
		ID:      id,
		Message: "Проект с ID " + id + " обновлен",
	}

	return c.JSON(http.StatusOK, updateResponse)
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
func updateCompletedWork(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	completedWorkId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор выполненной работы",
		})
	}

	var req request.CompletedWorkUpdateRequest
	if err = c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.Description != nil{
		updateData["description"] = *req.Description
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	result := dbConn.Session(&gorm.Session{}).Model(models.CompletedWork{}).Where("id = ?", completedWorkId).Updates(updateData)
	if result.Error != nil{
		log.Printf("DB error (update completedWork): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении проделанной работы",
		})
	}

	if result.RowsAffected == 0{
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не обновлено",
		})
	}	

	updateResponse := response.ReportUniversalResponse{
		ID:      id,
		Message: "Выполненная работа успешно обновлена",
	}

	return c.JSON(http.StatusOK, updateResponse)
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
func updateTomorrowPlans(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	tomorrowPlansId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор планов на завтра",
		})
	}

	var req request.TomorrowPlansUpdateRequest
	if err = c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.Description != nil{
		updateData["description"] = *req.Description
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	result := dbConn.Session(&gorm.Session{}).Model(models.TomorrowPlans{}).Where("id = ?", tomorrowPlansId).Updates(updateData)
	if result.Error != nil{
		log.Printf("DB error (update completedWork): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении планов на завтра",
		})
	}

	if result.RowsAffected == 0{
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не обновлено",
		})
	}

	updateResponse := response.ReportUniversalResponse{
		ID:      id,
		Message: "Планы на завтра успешно обновлены",
	}

	return c.JSON(http.StatusOK, updateResponse)
}	

