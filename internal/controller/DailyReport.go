package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterReportRoutes(e *echo.Group) {
	reportGroup := e.Group("/report")
	{
		reportGroup.GET("", GetAllReports)
		reportGroup.GET("/:id", GetReport)
		reportGroup.GET("/task/:id", GetReportsByTaskId)
		reportGroup.GET("/project/:id", GetReportsByProjectId)
		reportGroup.POST("", CreateReport)
		reportGroup.PUT("/:id", UpdateReport)
		reportGroup.DELETE("/:id", DeleteReport)
		reportGroup.PATCH("/help-request/:id", updateHelpRequest)
		reportGroup.PATCH("/completed-work/:id", updateCompletedWork)
		reportGroup.PATCH("/tomorrow-plans/:id", updateTomorrowPlans)
	}
}

func GetAllReports(c echo.Context) error {
	var req request.ReportListRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.DailyReport{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count reports): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете отчетов",
		})
	}

	var reports []models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&reports).Error; err != nil {
		log.Printf("DB error (find reports): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении отчетов из базы данных",
		})
	}

	reportList := response.ReportListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, report := range reports {
		var userId string
		if report.UserID != nil {
			userId = report.UserID.String()
		}

		var reportDate time.Time
		if report.ReportDate != nil {
			reportDate = *report.ReportDate
		}

		var status string
		if report.Status != nil {
			status = *report.Status
		}

		var taskId string
		if report.TaskID != nil {
			taskId = report.TaskID.String()
		}

		var createdAt time.Time
		if report.CreatedAt != nil {
			createdAt = *report.CreatedAt
		}

		var updatedAt time.Time
		if report.UpdatedAt != nil {
			updatedAt = *report.UpdatedAt
		}

		var user models.User
		result = dbConn.Session(&gorm.Session{}).Where("id = ? and deleted = ?", userId, false).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Пользователь не найден",
				})
			}
			log.Printf("DB error (find user by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении пользователя из базы данных",
			})
		}

		var firstName string
		if user.FirstName != nil {
			firstName = *user.FirstName
		}

		var lastName string
		if user.LastName != nil {
			lastName = *user.LastName
		}

		userInfo := response.UserShort{
			ID:        userId,
			FirstName: firstName,
			LastName:  lastName,
		}

		var helpReq models.HelpRequest
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? AND deleted = ?", report.ID, false).First(&helpReq)

		var hrId string
		if helpReq.ID != nil {
			hrId = helpReq.ID.String()
		}

		var helperId string
		if helpReq.HelperID != nil {
			helperId = helpReq.HelperID.String()
		}

		var description string
		if helpReq.Description != nil {
			description = *helpReq.Description
		}

		helpRequest := response.HelpRequestItem{
			ID:          hrId,
			HelperID:    helperId,
			Description: description,
		}

		var completedWork []models.CompletedWork
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", report.ID, false).Find(&completedWork)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Выполненная работа не найдена",
				})
			}
			log.Printf("DB error (find help requst by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении выполненной работы из базы данных",
			})
		}

		var complWork []response.WorkItem
		for _, workItem := range completedWork {
			var iD string
			if workItem.ID != nil {
				iD = workItem.ID.String()
			}

			var desc string
			if workItem.Description != nil {
				desc = *workItem.Description
			}
			complWork = append(complWork, response.WorkItem{
				ID:          iD,
				Description: desc,
			})
		}

		var tomorrowPlan []models.TomorrowPlans
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", report.ID, false).Find(&tomorrowPlan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "План на завтра не найден",
				})
			}
			log.Printf("DB error (find help requst by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении плана на завтра из базы данных",
			})
		}

		var tomorrowPlans []response.WorkItem
		for _, workItem := range tomorrowPlan {
			var iD string
			if workItem.ID != nil {
				iD = workItem.ID.String()
			}

			var desc string
			if workItem.Description != nil {
				desc = *workItem.Description
			}
			tomorrowPlans = append(tomorrowPlans, response.WorkItem{
				ID:          iD,
				Description: desc,
			})
		}

		var reportProblems []models.ReportProblem
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", report.ID, false).Find(&reportProblems)
		if result.Error != nil {
			log.Printf("DB error (find report-problem by id): %v", result.Error)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Ошибка при получении связи отчет-проблема из базы данных",
				})
			}
		}

		var problemsId []uuid.UUID
		for _, reportProblem := range reportProblems {
			if reportProblem.ReportID != nil {
				problemsId = append(problemsId, *reportProblem.ReportID)
			}
		}

		var problems []models.Problem
		result = dbConn.Session(&gorm.Session{}).Model(models.Problem{}).Where("id in (?)", problemsId).Find(&problems)
		if result.Error != nil {
			log.Printf("DB error (find problem by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении проблем из базы данных",
			})
		}

		var problemsResp []response.ProblemResponse
		for _, problem := range problems {
			var name string
			if problem.Name != nil {
				name = *problem.Name
			}
			var description []string
			if problem.Description != nil {
				description = *problem.Description
			}
			var creatorId string
			if problem.CreatorID != nil {
				creatorId = problem.CreatorID.String()
			}
			var createdAt time.Time
			if problem.CreatedAt != nil {
				createdAt = *problem.CreatedAt
			}
			var updatedAt time.Time
			if problem.UpdatedAt != nil {
				updatedAt = *problem.UpdatedAt
			}
			problemsResp = append(problemsResp, response.ProblemResponse{
				ID:          problem.ID.String(),
				Name:        name,
				Description: description,
				CreatorId:   creatorId,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			})
		}

		resp := response.ReportResponse{
			ID:            report.ID.String(),
			UserID:        userId,
			ReportDate:    reportDate,
			CompletedWork: complWork,
			PlanTomorrow:  tomorrowPlans,
			HelpRequest:   helpRequest,
			Status:        status,
			TaskId:        taskId,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
			UserInfo:      userInfo,
			Problems:      problemsResp,
		}

		reportList.Reports = append(reportList.Reports, resp)
	}

	return c.JSON(http.StatusOK, reportList)
}

func GetReport(c echo.Context) error {
	id := c.Param("id")
	reportId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор отчета",
		})
	}

	var report models.DailyReport
	result := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).First(&report, "id = ?", reportId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Отчет не найден",
			})
		}
		log.Printf("DB error (find report by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении отчета из базы данных",
		})
	}

	if report.ID == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Отчет не найден",
		})
	}

	var userId string
	if report.UserID != nil {
		userId = report.UserID.String()
	}

	var reportDate time.Time
	if report.ReportDate != nil {
		reportDate = *report.ReportDate
	}

	var status string
	if report.Status != nil {
		status = *report.Status
	}

	var taskId string
	if report.TaskID != nil {
		taskId = report.TaskID.String()
	}

	var createdAt time.Time
	if report.CreatedAt != nil {
		createdAt = *report.CreatedAt
	}

	var updatedAt time.Time
	if report.UpdatedAt != nil {
		updatedAt = *report.UpdatedAt
	}

	var user models.User
	result = dbConn.Session(&gorm.Session{}).Where("id = ? and deleted = ?", userId, false).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Пользователь не найден",
			})
		}
		log.Printf("DB error (find user by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении пользователя из базы данных",
		})
	}

	var firstName string
	if user.FirstName != nil {
		firstName = *user.FirstName
	}

	var lastName string
	if user.LastName != nil {
		lastName = *user.LastName
	}

	userInfo := response.UserShort{
		ID:        userId,
		FirstName: firstName,
		LastName:  lastName,
	}

	var helpReq models.HelpRequest
	result = dbConn.Session(&gorm.Session{}).Where("report_id = ? AND deleted = ?", reportId, false).First(&helpReq)

	var hrId string
	if helpReq.ID != nil {
		hrId = helpReq.ID.String()
	}

	var helperId string
	if helpReq.HelperID != nil {
		helperId = helpReq.HelperID.String()
	}

	var description string
	if helpReq.Description != nil {
		description = *helpReq.Description
	}

	helpRequest := response.HelpRequestItem{
		ID:          hrId,
		HelperID:    helperId,
		Description: description,
	}

	var completedWork []models.CompletedWork
	result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", reportId, false).Find(&completedWork)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Выполненная работа не найдена",
			})
		}
		log.Printf("DB error (find help requst by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении выполненной работы из базы данных",
		})
	}

	var complWork []response.WorkItem
	for _, workItem := range completedWork {
		var iD string
		if workItem.ID != nil {
			iD = workItem.ID.String()
		}

		var desc string
		if workItem.Description != nil {
			desc = *workItem.Description
		}
		complWork = append(complWork, response.WorkItem{
			ID:          iD,
			Description: desc,
		})
	}

	var tomorrowPlan []models.TomorrowPlans
	result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", reportId, false).Find(&tomorrowPlan)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "План на завтра не найден",
			})
		}
		log.Printf("DB error (find help requst by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении плана на завтра из базы данных",
		})
	}

	var tomorrowPlans []response.WorkItem
	for _, workItem := range tomorrowPlan {
		var iD string
		if workItem.ID != nil {
			iD = workItem.ID.String()
		}

		var desc string
		if workItem.Description != nil {
			desc = *workItem.Description
		}
		tomorrowPlans = append(tomorrowPlans, response.WorkItem{
			ID:          iD,
			Description: desc,
		})
	}

	var reportProblems []models.ReportProblem
	result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", id, false).Find(&reportProblems)
	if result.Error != nil {
		log.Printf("DB error (find report-problem by id): %v", result.Error)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Ошибка при получении связи отчет-проблема из базы данных",
			})
		}
	}

	var problemsId []uuid.UUID
	for _, reportProblem := range reportProblems {
		if reportProblem.ReportID != nil {
			problemsId = append(problemsId, *reportProblem.ReportID)
		}
	}

	var problems []models.Problem
	result = dbConn.Session(&gorm.Session{}).Model(models.Problem{}).Where("id in (?)", problemsId).Find(&problems)
	if result.Error != nil {
		log.Printf("DB error (find problem by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проблем из базы данных",
		})
	}

	var problemsResp []response.ProblemResponse
	for _, problem := range problems {
		var name string
		if problem.Name != nil {
			name = *problem.Name
		}
		var description []string
		if problem.Description != nil {
			description = *problem.Description
		}
		var creatorId string
		if problem.CreatorID != nil {
			creatorId = problem.CreatorID.String()
		}
		var createdAt time.Time
		if problem.CreatedAt != nil {
			createdAt = *problem.CreatedAt
		}
		var updatedAt time.Time
		if problem.UpdatedAt != nil {
			updatedAt = *problem.UpdatedAt
		}
		problemsResp = append(problemsResp, response.ProblemResponse{
			ID:          problem.ID.String(),
			Name:        name,
			Description: description,
			CreatorId:   creatorId,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}

	resp := response.ReportResponse{
		ID:            id,
		UserID:        userId,
		ReportDate:    reportDate,
		CompletedWork: complWork,
		PlanTomorrow:  tomorrowPlans,
		HelpRequest:   helpRequest,
		Status:        status,
		TaskId:        taskId,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		UserInfo:      userInfo,
		Problems:      problemsResp,
	}
	return c.JSON(http.StatusOK, resp)
}

func GetReportsByTaskId(c echo.Context) error {
	taskId := c.Param("id")
	var reports []models.DailyReport
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ? and task_id = ?", false, taskId).Find(&reports).Error; err != nil {
		log.Printf("DB error (find reports): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении отчетов из базы данных",
		})
	}

	reportList := response.ReportListByTaskId{
		TaskID: taskId,
	}

	for _, report := range reports {
		var userId string
		if report.UserID != nil {
			userId = report.UserID.String()
		}

		var reportDate time.Time
		if report.ReportDate != nil {
			reportDate = *report.ReportDate
		}

		var status string
		if report.Status != nil {
			status = *report.Status
		}

		var taskId string
		if report.TaskID != nil {
			taskId = report.TaskID.String()
		}

		var createdAt time.Time
		if report.CreatedAt != nil {
			createdAt = *report.CreatedAt
		}

		var updatedAt time.Time
		if report.UpdatedAt != nil {
			updatedAt = *report.UpdatedAt
		}

		var user models.User
		result := dbConn.Session(&gorm.Session{}).Where("id = ? and deleted = ?", userId, false).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Пользователь не найден",
				})
			}
			log.Printf("DB error (find user by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении пользователя из базы данных",
			})
		}

		var firstName string
		if user.FirstName != nil {
			firstName = *user.FirstName
		}

		var lastName string
		if user.LastName != nil {
			lastName = *user.LastName
		}

		userInfo := response.UserShort{
			ID:        userId,
			FirstName: firstName,
			LastName:  lastName,
		}

		var helpReq models.HelpRequest
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? AND deleted = ?", report.ID, false).First(&helpReq)

		var hrId string
		if helpReq.ID != nil {
			hrId = helpReq.ID.String()
		}

		var helperId string
		if helpReq.HelperID != nil {
			helperId = helpReq.HelperID.String()
		}

		var description string
		if helpReq.Description != nil {
			description = *helpReq.Description
		}

		helpRequest := response.HelpRequestItem{
			ID:          hrId,
			HelperID:    helperId,
			Description: description,
		}

		var completedWork []models.CompletedWork
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", report.ID, false).Find(&completedWork)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Выполненная работа не найдена",
				})
			}
			log.Printf("DB error (find help requst by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении выполненной работы из базы данных",
			})
		}

		var complWork []response.WorkItem
		for _, workItem := range completedWork {
			var iD string
			if workItem.ID != nil {
				iD = workItem.ID.String()
			}

			var desc string
			if workItem.Description != nil {
				desc = *workItem.Description
			}
			complWork = append(complWork, response.WorkItem{
				ID:          iD,
				Description: desc,
			})
		}

		var tomorrowPlan []models.TomorrowPlans
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", report.ID, false).Find(&tomorrowPlan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "План на завтра не найден",
				})
			}
			log.Printf("DB error (find help requst by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении плана на завтра из базы данных",
			})
		}

		var tomorrowPlans []response.WorkItem
		for _, workItem := range tomorrowPlan {
			var iD string
			if workItem.ID != nil {
				iD = workItem.ID.String()
			}

			var desc string
			if workItem.Description != nil {
				desc = *workItem.Description
			}
			tomorrowPlans = append(tomorrowPlans, response.WorkItem{
				ID:          iD,
				Description: desc,
			})
		}

		var reportProblems []models.ReportProblem
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", report.ID, false).Find(&reportProblems)
		if result.Error != nil {
			log.Printf("DB error (find report-problem by id): %v", result.Error)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Ошибка при получении связи отчет-проблема из базы данных",
				})
			}
		}

		var problemsId []uuid.UUID
		for _, reportProblem := range reportProblems {
			if reportProblem.ReportID != nil {
				problemsId = append(problemsId, *reportProblem.ReportID)
			}
		}

		var problems []models.Problem
		result = dbConn.Session(&gorm.Session{}).Model(models.Problem{}).Where("id in (?)", problemsId).Find(&problems)
		if result.Error != nil {
			log.Printf("DB error (find problem by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении проблем из базы данных",
			})
		}

		var problemsResp []response.ProblemResponse
		for _, problem := range problems {
			var name string
			if problem.Name != nil {
				name = *problem.Name
			}
			var description []string
			if problem.Description != nil {
				description = *problem.Description
			}
			var creatorId string
			if problem.CreatorID != nil {
				creatorId = problem.CreatorID.String()
			}
			var createdAt time.Time
			if problem.CreatedAt != nil {
				createdAt = *problem.CreatedAt
			}
			var updatedAt time.Time
			if problem.UpdatedAt != nil {
				updatedAt = *problem.UpdatedAt
			}
			problemsResp = append(problemsResp, response.ProblemResponse{
				ID:          problem.ID.String(),
				Name:        name,
				Description: description,
				CreatorId:   creatorId,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			})
		}

		resp := response.ReportResponse{
			ID:            report.ID.String(),
			UserID:        userId,
			ReportDate:    reportDate,
			CompletedWork: complWork,
			PlanTomorrow:  tomorrowPlans,
			HelpRequest:   helpRequest,
			Status:        status,
			TaskId:        taskId,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
			UserInfo:      userInfo,
			Problems:      problemsResp,
		}

		reportList.Reports = append(reportList.Reports, resp)
	}

	return c.JSON(http.StatusOK, reportList)
}

func GetReportsByProjectId(c echo.Context) error {
	projectId := c.Param("id")

	var task []models.Task
	result := dbConn.Session(&gorm.Session{}).Where("project_id = ? and deleted = ?", projectId, false).Find(&task)
	if result.Error != nil {
		log.Printf("DB error (find tasks by project id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении задач из базы данных",
		})
	}

	var tasksId []uuid.UUID
	for _, taskItem := range task {
		if taskItem.ID != nil {
			tasksId = append(tasksId, *taskItem.ID)
		}
	}

	var reports []models.DailyReport
	result = dbConn.Session(&gorm.Session{}).Where("task_id in (?)", tasksId).Find(&reports)
	if result.Error != nil {
		log.Printf("DB error (find reports by project id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении отчетов из базы данных",
		})
	}

	reportList := response.ReportListByProjectId{
		ProjectID: projectId,
	}

	for _, report := range reports {
		var userId string
		if report.UserID != nil {
			userId = report.UserID.String()
		}

		var reportDate time.Time
		if report.ReportDate != nil {
			reportDate = *report.ReportDate
		}

		var status string
		if report.Status != nil {
			status = *report.Status
		}

		var taskId string
		if report.TaskID != nil {
			taskId = report.TaskID.String()
		}

		var createdAt time.Time
		if report.CreatedAt != nil {
			createdAt = *report.CreatedAt
		}

		var updatedAt time.Time
		if report.UpdatedAt != nil {
			updatedAt = *report.UpdatedAt
		}

		var user models.User
		result := dbConn.Session(&gorm.Session{}).Where("id = ? and deleted = ?", userId, false).First(&user)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Пользователь не найден",
				})
			}
			log.Printf("DB error (find user by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении пользователя из базы данных",
			})
		}

		var firstName string
		if user.FirstName != nil {
			firstName = *user.FirstName
		}

		var lastName string
		if user.LastName != nil {
			lastName = *user.LastName
		}

		userInfo := response.UserShort{
			ID:        userId,
			FirstName: firstName,
			LastName:  lastName,
		}

		var helpReq models.HelpRequest
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? AND deleted = ?", report.ID, false).First(&helpReq)

		var hrId string
		if helpReq.ID != nil {
			hrId = helpReq.ID.String()
		}

		var helperId string
		if helpReq.HelperID != nil {
			helperId = helpReq.HelperID.String()
		}

		var description string
		if helpReq.Description != nil {
			description = *helpReq.Description
		}

		helpRequest := response.HelpRequestItem{
			ID:          hrId,
			HelperID:    helperId,
			Description: description,
		}

		var completedWork []models.CompletedWork
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", report.ID, false).Find(&completedWork)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Выполненная работа не найдена",
				})
			}
			log.Printf("DB error (find help requst by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении выполненной работы из базы данных",
			})
		}

		var complWork []response.WorkItem
		for _, workItem := range completedWork {
			var iD string
			if workItem.ID != nil {
				iD = workItem.ID.String()
			}

			var desc string
			if workItem.Description != nil {
				desc = *workItem.Description
			}
			complWork = append(complWork, response.WorkItem{
				ID:          iD,
				Description: desc,
			})
		}

		var tomorrowPlan []models.TomorrowPlans
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", report.ID, false).Find(&tomorrowPlan)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "План на завтра не найден",
				})
			}
			log.Printf("DB error (find help requst by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении плана на завтра из базы данных",
			})
		}

		var tomorrowPlans []response.WorkItem
		for _, workItem := range tomorrowPlan {
			var iD string
			if workItem.ID != nil {
				iD = workItem.ID.String()
			}

			var desc string
			if workItem.Description != nil {
				desc = *workItem.Description
			}
			tomorrowPlans = append(tomorrowPlans, response.WorkItem{
				ID:          iD,
				Description: desc,
			})
		}

		var reportProblems []models.ReportProblem
		result = dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", report.ID, false).Find(&reportProblems)
		if result.Error != nil {
			log.Printf("DB error (find report-problem by id): %v", result.Error)
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Ошибка при получении связи отчет-проблема из базы данных",
				})
			}
		}

		var problemsId []uuid.UUID
		for _, reportProblem := range reportProblems {
			if reportProblem.ReportID != nil {
				problemsId = append(problemsId, *reportProblem.ReportID)
			}
		}

		var problems []models.Problem
		result = dbConn.Session(&gorm.Session{}).Model(models.Problem{}).Where("id in (?)", problemsId).Find(&problems)
		if result.Error != nil {
			log.Printf("DB error (find problem by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении проблем из базы данных",
			})
		}

		var problemsResp []response.ProblemResponse
		for _, problem := range problems {
			var name string
			if problem.Name != nil {
				name = *problem.Name
			}
			var description []string
			if problem.Description != nil {
				description = *problem.Description
			}
			var creatorId string
			if problem.CreatorID != nil {
				creatorId = problem.CreatorID.String()
			}
			var createdAt time.Time
			if problem.CreatedAt != nil {
				createdAt = *problem.CreatedAt
			}
			var updatedAt time.Time
			if problem.UpdatedAt != nil {
				updatedAt = *problem.UpdatedAt
			}
			problemsResp = append(problemsResp, response.ProblemResponse{
				ID:          problem.ID.String(),
				Name:        name,
				Description: description,
				CreatorId:   creatorId,
				CreatedAt:   createdAt,
				UpdatedAt:   updatedAt,
			})
		}

		resp := response.ReportResponse{
			ID:            report.ID.String(),
			UserID:        userId,
			ReportDate:    reportDate,
			CompletedWork: complWork,
			PlanTomorrow:  tomorrowPlans,
			HelpRequest:   helpRequest,
			Status:        status,
			TaskId:        taskId,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
			UserInfo:      userInfo,
			Problems:      problemsResp,
		}

		reportList.Reports = append(reportList.Reports, resp)
	}

	return c.JSON(http.StatusOK, reportList)
}

func CreateReport(c echo.Context) error {
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

	report := models.DailyReport{
		ID:         &newUUID,
		UserID:     &userId,
		Status:     req.Status,
		TaskID:     &taskId,
		ReportDate: req.ReportDate,
		CreatedAt:  &now,
	}

	result := dbConn.Session(&gorm.Session{}).Create(&report)
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
				ID:          &tempUUID,
				Description: &completedWorks.Description,
				ReportID:    &newUUID,
			}

			result = dbConn.Session(&gorm.Session{}).Create(&complWork)
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
				ID:          &tempUUID,
				Description: &tomorrowPlan.Description,
				ReportID:    &newUUID,
			}

			result = dbConn.Session(&gorm.Session{}).Create(&tomPlan)
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
			tempUUID := uuid.New()

			curProblem := models.Problem{
				ID:          &tempUUID,
				Description: &problem.Description,
				CreatorID:   &userId,
				CreatedAt:   &now,
			}

			result = dbConn.Session(&gorm.Session{}).Create(&curProblem)
			if result.Error != nil {
				log.Printf("DB error (create report): %v", result.Error)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Ошибка при создании проблемы",
				})
			}

			reportProblem := models.ReportProblem{
				ReportID:  &newUUID,
				ProblemID: &tempUUID,
			}

			result = dbConn.Session(&gorm.Session{}).Create(&reportProblem)
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
			ID:          &tempUUID,
			HelperID:    &helperId,
			Description: &req.Help.Description,
			ReportID:    &newUUID,
		}

		result = dbConn.Session(&gorm.Session{}).Create(&help)
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

	return c.JSON(http.StatusOK, resp)
}

func UpdateReport(c echo.Context) error {
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
		updateData["user_id "] = *req.UserId
	}
	if req.ReportDate != nil{
		updateData["report_date"] = *req.ReportDate
	}
	updateData["updated_at"] = time.Now()

	if err = dbConn.Session(&gorm.Session{}).Model(models.DailyReport{}).Where("id = ?", reportId).Updates(updateData).Error ;err != nil{
		log.Printf("DB error (update report): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении отчета",
		})
	}

	updateResponse := response.ReportUniversalResponse{
		ID: id,
		Message: "Проект успешно изменен",
	}

	return c.JSON(http.StatusOK, updateResponse)
}

func DeleteReport(c echo.Context) error {
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	updateData["updated_at"] = time.Now()
	result := dbConn.Session(&gorm.Session{}).Model(models.DailyReport{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete report): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении отчета",
		})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не удалено",
		})
	}	

	result = dbConn.Session(&gorm.Session{}).Model(models.HelpRequest{}).Where("report_id = ?").Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete report): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении запроса на помощь",
		})
	}
	
	result = dbConn.Session(&gorm.Session{}).Model(models.CompletedWork{}).Where("report_id = ?").Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete report): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении выполненной за день работы",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.TomorrowPlans{}).Where("report_id = ?").Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete report): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении планов на завтра",
		})
	}

	var reportProblems []models.ReportProblem
	if err := dbConn.Session(&gorm.Session{}).Where("report_id = ? and deleted = ?", id, false).Find(&reportProblems); err != nil{
		log.Printf("DB error (get report-problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении связи отчет-проблема",
		})
	}

	var problemsId []uuid.UUID
	for _, reportProblem := range reportProblems{
		problemsId = append(problemsId, *reportProblem.ProblemID)
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.Problem{}).Where("id in (?)", problemsId).Updates(updateData)
	if result.Error != nil{
		log.Printf("DB error (delete problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении проблем",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.ReportProblem{}).Where("report_id = ?", id).Updates(updateData)
	if result.Error != nil{
		log.Printf("DB error (delete report-problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении связей отчет-проблема",
		})
	}
	
	resp := response.ReportUniversalResponse{
		ID: id,
		Message: "Отчет успешно удален",
	}	

	return c.JSON(http.StatusOK, resp)
}

func updateHelpRequest(c echo.Context) error{
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

	updateData["created_at"] = time.Now()

	if err = dbConn.Session(&gorm.Session{}).Model(models.HelpRequest{}).Where("id = ?", helpRequestId).Error; err != nil{
		log.Printf("DB error (update help request): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении запроса на помощь",
		})
	}

	updateResponse := response.ReportUniversalResponse{
		ID:      id,
		Message: "Проект с ID " + id + " обновлен",
	}

	return c.JSON(http.StatusOK, updateResponse)
}

func updateCompletedWork(c echo.Context) error{
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

	if err = dbConn.Session(&gorm.Session{}).Model(models.CompletedWork{}).Where("id = ?", completedWorkId).Error; err != nil{
		log.Printf("DB error (update completedWork): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении проделанной работы",
		})
	}

	updateResponse := response.ReportUniversalResponse{
		ID:      id,
		Message: "Выполненная работа успешно обновлена",
	}

	return c.JSON(http.StatusOK, updateResponse)
}

func updateTomorrowPlans(c echo.Context) error{
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

	if err = dbConn.Session(&gorm.Session{}).Model(models.TomorrowPlans{}).Where("id = ?", tomorrowPlansId).Error; err != nil{
		log.Printf("DB error (update completedWork): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении планов на завтра",
		})
	}

	updateResponse := response.ReportUniversalResponse{
		ID:      id,
		Message: "Выполненная работа успешно обновлена",
	}

	return c.JSON(http.StatusOK, updateResponse)
}	

