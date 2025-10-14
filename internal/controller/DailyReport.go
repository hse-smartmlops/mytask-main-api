package controller

import (
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"emplacc-api/internal/service"
	"emplacc-api/internal/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type ReportController struct {
	reportService service.ReportService
}

func NewReportController(reportService service.ReportService) *ReportController {
	return &ReportController{
		reportService: reportService,
	}
}

func RegisterReportRoutes(e *echo.Echo, reportService service.ReportService) {
	controller := NewReportController(reportService)
	reportGroup := e.Group("/report")
	{
		reportGroup.GET("/all/:page/:pagesize", controller.GetAllReports)
		reportGroup.GET("/:id", controller.GetReport)
		reportGroup.GET("/task/:id", controller.GetReportsByTaskId)
		reportGroup.GET("/project/:id", controller.GetReportsByProjectId)
		reportGroup.POST("", controller.CreateReport)
		reportGroup.PATCH("/:id", controller.UpdateReport)
		reportGroup.DELETE("/:id", controller.DeleteReport)
		reportGroup.PATCH("/help-request/:id", controller.UpdateHelpRequest)
		reportGroup.PATCH("/completed-work/:id", controller.UpdateCompletedWork)
		reportGroup.PATCH("/tomorrow-plans/:id", controller.UpdateTomorrowPlans)
		reportGroup.GET("/user/:id/:page/:pagesize", controller.GetAllReportsByUserId)
		reportGroup.GET("/help-requests-by-user-id/:id", controller.GetHelpRequestsForUser)
		reportGroup.DELETE("/help-request/:id", controller.DeleteHelpRequest)
	}
}

// GetAllReports godoc
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
func (rc *ReportController) GetAllReports(c echo.Context) error {
	page, _ := strconv.Atoi(c.Param("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Param("pagesize"))
	if pageSize <= 0 {
		pageSize = 10
	}

	reports, totalCount, err := rc.reportService.GetAllReports(page, pageSize)
	if err != nil {
		log.Printf("service error (get all reports): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете отчетов"})
	}

	out := response.ReportListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, r := range reports {
		userInfo := response.UserShort{
			ID:        r.UserID.String(),
			FirstName: r.User.FirstName,
			LastName:  r.User.LastName,
		}
		var complWork []response.CompletedWork
		for _, w := range r.CompletedWork {
			complWork = append(complWork, response.CompletedWork{
				ID:          w.ID.String(),
				Description: utils.GetString(w.Description),
				TaskId:  utils.GetUUIDString(w.TaskID),
			})
		}
		var plans []response.TomorrowPlans
		for _, p := range r.TomorrowPlans {
			plans = append(plans, response.TomorrowPlans{
				ID:          p.ID.String(),
				Description: utils.GetString(p.Description),
				TaskId:  utils.GetUUIDString(p.TaskID),
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
		for _, hr := range r.HelpRequests {
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

// GetAllReportsByUserId godoc
// @Summary Получение списка отчетов по ID пользователя
// @Description Получает список отчетов, созданных конкретным пользователем, с пагинацией и исключением удаленных записей
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя (UUID)"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.ReportListResponse "Список отчетов успешно получен"
// @Failure 400 {object} map[string]string "Некорректный ID пользователя"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении отчетов"
// @Router /report/user/{id}/{page}/{pagesize} [get]
func (rc *ReportController) GetAllReportsByUserId(c echo.Context) error {
	page, _ := strconv.Atoi(c.Param("page"))
	if page <= 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Param("pagesize"))
	if pageSize <= 0 {
		pageSize = 10
	}
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID пользователя"})
	}

	reports, totalCount, err := rc.reportService.GetAllReportsByUserId(userID, page, pageSize)
	if err != nil {
		log.Printf("service error (get reports by user id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете отчетов"})
	}

	out := response.ReportListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, r := range reports {
		userInfo := response.UserShort{
			ID:        r.UserID.String(),
			FirstName: r.User.FirstName,
			LastName:  r.User.LastName,
		}
		var complWork []response.CompletedWork
		for _, w := range r.CompletedWork {
			complWork = append(complWork, response.CompletedWork{
				ID:          w.ID.String(),
				Description: utils.GetString(w.Description),
				TaskId: utils.GetUUIDString(w.TaskID),
			})
		}
		var plans []response.TomorrowPlans
		for _, p := range r.TomorrowPlans {
			plans = append(plans, response.TomorrowPlans{
				ID:          p.ID.String(),
				Description: utils.GetString(p.Description),
				TaskId:  utils.GetUUIDString(p.TaskID),
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
		for _, hr := range r.HelpRequests {
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

// GetReport godoc
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
func (rc *ReportController) GetReport(c echo.Context) error {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор отчета"})
	}

	r, err := rc.reportService.GetReport(reportID)
	if err != nil {
		if err.Error() == "report not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Отчет не найден"})
		}
		log.Printf("service error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении отчета"})
	}

	userInfo := response.UserShort{
		ID:        r.UserID.String(),
		FirstName: r.User.FirstName,
		LastName:  r.User.LastName,
	}
	var complWork []response.CompletedWork
	for _, w := range r.CompletedWork {
		complWork = append(complWork, response.CompletedWork{
			ID:          w.ID.String(),
			Description: utils.GetString(w.Description),
			TaskId:  utils.GetUUIDString(w.TaskID),
		})
	}
	var plans []response.TomorrowPlans
	for _, p := range r.TomorrowPlans {
		plans = append(plans, response.TomorrowPlans{
			ID:          p.ID.String(),
			Description: utils.GetString(p.Description),
			TaskId:  utils.GetUUIDString(p.TaskID),
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
	for _, hr := range r.HelpRequests {
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

// GetReportsByTaskId godoc
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
func (rc *ReportController) GetReportsByTaskId(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID задачи"})
	}

	reports, err := rc.reportService.GetReportsByTaskId(taskID)
	if err != nil {
		log.Printf("service error (get reports by task): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении работ по задаче"})
	}

	out := response.ReportListByTaskId{TaskID: taskID.String()}
	for _, r := range reports {
		userInfo := response.UserShort{
			ID:        r.UserID.String(),
			FirstName: r.User.FirstName,
			LastName:  r.User.LastName,
		}
		var complWork []response.CompletedWork
		for _, w := range r.CompletedWork {
			complWork = append(complWork, response.CompletedWork{
				ID:          w.ID.String(),
				Description: utils.GetString(w.Description),
				TaskId:  utils.GetUUIDString(w.TaskID),
			})
		}
		var plans []response.TomorrowPlans
		for _, p := range r.TomorrowPlans {
			plans = append(plans, response.TomorrowPlans{
				ID:          p.ID.String(),
				Description: utils.GetString(p.Description),
				TaskId:  utils.GetUUIDString(p.TaskID),
			})
		}
		helpResp := []response.HelpRequestItem{}
		for _, hr := range r.HelpRequests {
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

// GetReportsByProjectId godoc
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
func (rc *ReportController) GetReportsByProjectId(c echo.Context) error {
	projectUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
	}

	reports, err := rc.reportService.GetReportsByProjectId(projectUUID)
	if err != nil {
		log.Printf("service error (get reports by project id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задач"})
	}

	out := response.ReportListByProjectId{ProjectID: projectUUID.String()}
	for _, r := range reports {
		userInfo := response.UserShort{
			ID:        r.UserID.String(),
			FirstName: r.User.FirstName,
			LastName:  r.User.LastName,
		}
		var complWork []response.CompletedWork
		for _, w := range r.CompletedWork {
			complWork = append(complWork, response.CompletedWork{
				ID:          w.ID.String(),
				Description: utils.GetString(w.Description),
				TaskId:  utils.GetUUIDString(w.TaskID),
			})
		}
		var plans []response.TomorrowPlans
		for _, p := range r.TomorrowPlans {
			plans = append(plans, response.TomorrowPlans{
				ID:          p.ID.String(),
				Description: utils.GetString(p.Description),
				TaskId:  utils.GetUUIDString(p.TaskID),
			})
		}
		helpResp := []response.HelpRequestItem{}
		for _, hr := range r.HelpRequests {
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

// CreateReport godoc
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
func (rc *ReportController) CreateReport(c echo.Context) error {
	var req request.ReportCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	reportID, err := rc.reportService.CreateReport(req)
	if err != nil {
		if err.Error() == "invalid user id" {
			log.Printf("ParseUserId error: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось распарсить userId"})
		}
		log.Printf("service error (create report): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании отчета"})
	}

	return c.JSON(http.StatusCreated, response.ReportUniversalResponse{
		ID:      reportID.String(),
		Message: "Успешно создано",
	})
}

// UpdateReport godoc
// @Summary Обновление отчета и связанных данных, ЕСЛИ НЕ УКАЗЫВАТЬ ID ВО ВСПОМОГАТЕЛЬНЫХ СУЩНОСТЯХ, СОЗДАЕТ НОВЫЕ
// @Tags Reports
// @Description Обновляет поля отчета, а также связанные CompletedWork, TomorrowPlans, HelpRequests и ReportProblems, ЕСЛИ НЕ УКАЗЫВАТЬ ID ВО ВСПОМОГАТЕЛЬНЫХ СУЩНОСТЯХ, СОЗДАЕТ НОВЫЕ
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
func (rc *ReportController) UpdateReport(c echo.Context) error {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор отчета"})
	}
	var req request.ReportReplaceRequest
	if err = c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	err = rc.reportService.UpdateReport(reportID, req)
	if err != nil {
		if err.Error() == "report not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Отчет не найден"})
		}
		log.Printf("service error (update report): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении отчета"})
	}

	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      reportID.String(),
		Message: "Отчет успешно изменен",
	})
}

// DeleteReport godoc
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
func (rc *ReportController) DeleteReport(c echo.Context) error {
	reportID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор отчета"})
	}

	err = rc.reportService.DeleteReport(reportID)
	if err != nil {
		if err.Error() == "report not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		log.Printf("service error (delete report): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении отчета"})
	}

	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      reportID.String(),
		Message: "Отчет успешно удален",
	})
}

// UpdateHelpRequest godoc
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
func (rc *ReportController) UpdateHelpRequest(c echo.Context) error {
	helpID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор запроса помощи"})
	}
	var req request.HelpRequestUpdateRequest
	if err = c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	err = rc.reportService.UpdateHelpRequest(helpID, req)
	if err != nil {
		if err.Error() == "no fields to update" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
		}
		if err.Error() == "help request not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		log.Printf("service error (update help request): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении запроса на помощь"})
	}

	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      helpID.String(),
		Message: "Запрос на помощь обновлен",
	})
}

// UpdateCompletedWork godoc
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
func (rc *ReportController) UpdateCompletedWork(c echo.Context) error {
	cwID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор выполненной работы"})
	}
	var req request.CompletedWorkUpdateRequest
	if err = c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	err = rc.reportService.UpdateCompletedWork(cwID, req)
	if err != nil {
		if err.Error() == "no fields to update" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
		}
		if err.Error() == "completed work not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		log.Printf("service error (update completed work): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении проделанной работы"})
	}

	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      cwID.String(),
		Message: "Выполненная работа обновлена",
	})
}

// UpdateTomorrowPlans godoc
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
func (rc *ReportController) UpdateTomorrowPlans(c echo.Context) error {
	tpID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор планов на завтра"})
	}
	var req request.TomorrowPlansUpdateRequest
	if err = c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	err = rc.reportService.UpdateTomorrowPlans(tpID, req)
	if err != nil {
		if err.Error() == "no fields to update" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
		}
		if err.Error() == "tomorrow plans not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		log.Printf("service error (update tomorrow plans): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении планов на завтра"})
	}

	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      tpID.String(),
		Message: "Планы на завтра обновлены",
	})
}

// GetHelpRequestsForUser godoc
// @Summary Получение запросов на помощь по ID пользователя-помощника
// @Description Возвращает список запросов на помощь, где указанный пользователь назначен в качестве помощника. В ответе также возвращаются имя и фамилия пользователя, создавшего запрос (автора ежедневного отчёта).
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя-помощника (в формате UUID)"
// @Security BearerAuth
// @Success 200 {object} response.HelpRequestsForUser "Список запросов на помощь успешно получен"
// @Failure 400 {object} map[string]string "Некорректный ID пользователя"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении запросов на помощь"
// @Router /report/help-requests-by-user-id/{id} [get]
func (rc *ReportController) GetHelpRequestsForUser(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID пользователя"})
	}

	helpRequests, err := rc.reportService.GetHelpRequestsForUser(userID)
	if err != nil {
		log.Printf("service error (find help_requests): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении запросов на помощь"})
	}

	helpRequestResponse := response.HelpRequestsForUser{}
	for _, helpRequest := range helpRequests {
		var firstName, lastName string
		if helpRequest.Report != nil && helpRequest.Report.User != nil {
			firstName = helpRequest.Report.User.FirstName
			lastName = helpRequest.Report.User.LastName
		}
		helpRequestResp := response.HelpRequestWithAssignerID{
			HelpRequest: response.HelpRequestItem{
				ID:          helpRequest.ID.String(),
				HelperID:    utils.GetUUIDString(helpRequest.HelperID),
				Description: utils.GetString(helpRequest.Description),
				Status:      utils.GetString(helpRequest.Status),
			},
			UserFirstName: firstName,
			UserLastName:  lastName,
		}
		helpRequestResponse.HelpRequests = append(helpRequestResponse.HelpRequests, helpRequestResp)
	}
	return c.JSON(http.StatusOK, helpRequestResponse)
}

// DeleteHelpRequest godoc
// @Summary Удаление запроса на помощь
// @Description Логическое удаление запроса на помощь по ID (установка поля deleted = true)
// @Tags Reports
// @Accept json
// @Produce json
// @Param id path string true "ID запроса на помощь"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ReportUniversalResponse "Запрос на помощь успешно удален"
// @Failure 400 {object} map[string]string "Некорректный идентификатор запроса на помощь"
// @Failure 404 {object} map[string]string "Запрос на помощь не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении запроса на помощь"
// @Router /report/help-request/{id} [delete]
func (rc *ReportController) DeleteHelpRequest(c echo.Context) error {
	requestID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор запроса на помощь"})
	}

	err = rc.reportService.DeleteHelpRequest(requestID)
	if err != nil {
		if err.Error() == "help request not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		log.Printf("service error (delete help request): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении запроса на помощь"})
	}

	return c.JSON(http.StatusOK, response.ReportUniversalResponse{
		ID:      requestID.String(),
		Message: "Запрос на помощь успешно удален",
	})
}