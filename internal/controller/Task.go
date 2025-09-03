package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
)

func RegisterTaskRoutes(e *echo.Echo) {
	taskGroup := e.Group("/task")
	taskGroup.Use(KeycloakAuthMiddleware)
	{
		taskGroup.GET("/all/:page/:pagesize", GetAllTasks)
		taskGroup.GET("/:id", GetTaskByID)
		taskGroup.GET("/project/:projectId", GetTasksByProjectID)
		taskGroup.GET("/filter", GetTasksByFilter)
		taskGroup.POST("", CreateTask)
		taskGroup.PATCH("/:id", UpdateTask)
		taskGroup.DELETE("/:id", DeleteTask)
	}
}

// GetAllTasks godoc
// @Summary Получение списка всех задач
// @Description Получает список всех задач с учетом пагинации, исключая удаленные
// @Tags Tasks
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TaskListResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/all/{page}/{pagesize} [get]
func GetAllTasks(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	pageReq := c.Param("page")
	pageSizeReq := c.Param("pagesize")
	// Значения по умолчанию
	page, err := strconv.Atoi(pageReq)
	if err != nil{
		log.Printf("failed to parse page: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге страницы",
		})
	}
	if page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeReq)
	if err != nil{
		log.Printf("failed to parse pagesize: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге номера страницы",
		})
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var tasks []models.Task
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		log.Printf("DB error (find tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении задач из базы данных",
		})
	}

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.Task{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count tasks): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете задач",
		})
	}

	taskList := response.TaskListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, task := range tasks {
		var id string = task.ID.String()

		var projectId string = task.ProjectID.String()

		var name string
		if task.Name != nil {
			name = *task.Name
		}

		var status string
		if task.Status != nil {
			status = *task.Status
		}

		var priority int16
		if task.Priority != nil {
			priority = *task.Priority
		}

		var startTime time.Time
		if task.StartDate != nil {
			startTime = *task.StartDate
		}

		var deadLine time.Time
		if task.Deadline != nil {
			deadLine = *task.Deadline
		}

		var updatedAt time.Time
		if task.UpdatedAt != nil {
			updatedAt = *task.UpdatedAt
		}

		taskList.Tasks = append(taskList.Tasks, response.TaskShort{
			ID:        id,
			Name:      name,
			ProjectID: projectId,
			Status:    status,
			Priority:  priority,
			StartDate: startTime,
			Deadline:  deadLine,
			UpdatedAt: updatedAt,
		})
	}
	return c.JSON(http.StatusOK, taskList)
}

// GetTaskByID godoc
// @Summary Получение задачи по ID
// @Description Получает данные задачи по её уникальному идентификатору
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.GetTaskByIDResponse "Задача успешно получена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор задачи"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задачи"
// @Router /task/{id} [get]
func GetTaskByID(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	taskId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор задачи",
		})
	}

	var task models.Task
	result := dbConn.Session(&gorm.Session{}).First(&task, "id = ? AND deleted = ?", taskId, false)
	if result.Error != nil {
		log.Printf("DB error %v", err)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Задача не найдена",
			})
		}
		log.Printf("DB error (find project by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении задачи из базы данных",
		})
	}

	if task.Deleted != nil {
		if *task.Deleted {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Задача не найдена",
			})
		}
	}

	creator := models.User{}
	creatorId := *task.CreatedBy
	err = dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", creatorId, false).First(&creator).Error
	if err != nil {
		log.Printf("DB error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении пользователя из базы данных",
		})
	}

	var creatorInfo response.UserShort
	if creator.Deleted != nil {
		if *creator.Deleted {
			var creatorID string = creator.ID.String()

			var creatorFirstName string
			if creator.FirstName != nil {
				creatorFirstName = *creator.FirstName
			}

			var creatorLastName string
			if creator.LastName != nil {
				creatorLastName = *creator.LastName
			}
			creatorInfo = response.UserShort{ID: creatorID, FirstName: creatorFirstName, LastName: creatorLastName}
		}
	}

	assigner := models.User{}
	assignerId := task.AssignedTo
	err = dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", assignerId, false).First(&assigner).Error
	if err != nil {
		log.Printf("DB error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении пользователя из базы данных",
		})
	}

	var assignerInfo response.UserShort
	if assigner.Deleted != nil {
		if *assigner.Deleted {
			var assignerID = assigner.ID.String()

			var assignerFirstName string
			if assigner.FirstName != nil {
				assignerFirstName = *assigner.FirstName
			}

			var assignerLastName string
			if assigner.LastName != nil {
				assignerLastName = *assigner.LastName
			}

			assignerInfo = response.UserShort{ID: assignerID, FirstName: assignerFirstName, LastName: assignerLastName}
		}
	}

	var updatedAt time.Time
	if task.UpdatedAt != nil {
		updatedAt = *task.UpdatedAt
	}

	var Id string = task.ID.String()

	var projectId string = task.ProjectID.String()

	var name string
	if task.Name != nil {
		name = *task.Name
	}

	var description string
	if task.Description != nil {
		description = *task.Description
	}

	var status string
	if task.Status != nil {
		status = *task.Status
	}

	var priority int16
	if task.Priority != nil {
		priority = *task.Priority
	}

	var startTime time.Time
	if task.StartDate != nil {
		startTime = *task.StartDate
	}

	var deadLine time.Time
	if task.Deadline != nil {
		deadLine = *task.Deadline
	}

	var timeSpent string
	if task.TimeSpent != nil {
		timeSpent = *task.TimeSpent
	}

	var gitLabIssueId int
	if task.GitlabIssueID != nil {
		gitLabIssueId = *task.GitlabIssueID
	}

	var category int8
	if task.Category != nil {
		category = *task.Category
	}
	taskResponse := response.GetTaskByIDResponse{
		ID:            Id,
		ProjectID:     projectId,
		Name:          name,
		Description:   description,
		Status:        status,
		Priority:      priority,
		CreatedBy:     creatorInfo,
		AssignedTo:    assignerInfo,
		Deadline:      deadLine,
		TimeSpent:     timeSpent,
		StartDate:     startTime,
		GitlabIssueID: gitLabIssueId,
		Category:      category,
		UpdatedAt:     updatedAt,
	}
	return c.JSON(http.StatusOK, taskResponse)
}

// GetTasksByProjectID godoc
// @Summary Получение задач по ID проекта
// @Description Получает список задач, связанных с указанным проектом
// @Tags Tasks
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TaskListResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проекта"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/project/{projectId} [get]
func GetTasksByProjectID(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	// Получаем projectID из параметров URL
	projectIDParam := c.Param("projectId")
	projectUUID, err := uuid.Parse(projectIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
	}
	var req request.TaskListRequest
	if err = c.Bind(&req); err != nil {
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

	var tasks []models.Task
	if err = dbConn.Session(&gorm.Session{}).Where("project_id = ? AND deleted = ?", projectUUID, false).
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		log.Printf("DB error (find tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении задач из базы данных",
		})
	}

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.Task{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count tasks): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете задач",
		})
	}

	taskList := response.TaskListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, task := range tasks {
		if task.Deleted != nil {
			if !*task.Deleted {
				var id string = task.ID.String()

				var projectId string = task.ProjectID.String()

				var name string
				if task.Name != nil {
					name = *task.Name
				}

				var status string
				if task.Status != nil {
					status = *task.Status
				}

				var priority int16
				if task.Priority != nil {
					priority = *task.Priority
				}

				var startTime time.Time
				if task.StartDate != nil {
					startTime = *task.StartDate
				}

				var deadLine time.Time
				if task.Deadline != nil {
					deadLine = *task.Deadline
				}

				var updatedAt time.Time
				if task.UpdatedAt != nil {
					updatedAt = *task.UpdatedAt
				}

				taskList.Tasks = append(taskList.Tasks, response.TaskShort{
					ID:        id,
					Name:      name,
					ProjectID: projectId,
					Status:    status,
					Priority:  priority,
					StartDate: startTime,
					Deadline:  deadLine,
					UpdatedAt: updatedAt,
				})
			}
		}
	}
	return c.JSON(http.StatusOK, taskList)
}


func GetTasksByFilter(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
		if err := authorize(c); err != nil {
			return err
		}
	filter := new(request.TaskFilter)
	if err := c.Bind(filter); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid filter parameters"})
	}
	return c.JSON(http.StatusOK, nil)
}

// CreateTask godoc
// @Summary Создание новой задачи
// @Description Создает новую задачу с указанными параметрами
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task body request.TaskCreateRequest true "Данные для создания задачи"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.TaskUniversaResponse "Задача успешно создана"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 404 {object} map[string]string "Исполнитель или поручитель не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании задачи"
// @Router /task [post]
func CreateTask(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	var req request.TaskCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	var name *string
	if req.Name != "" {
		name = &req.Name
	}

	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	var status *string
	if req.Status != nil {
		status = req.Status
	}

	var priority *int16
	if req.Priority != nil {
		priority = req.Priority
	}

	var deadline *time.Time
	if req.Deadline != nil {
		deadline = req.Deadline
	}

	var startDate *time.Time
	if req.StartDate != nil {
		startDate = req.StartDate
	}

	var gitLabIssueID *int
	if req.GitlabIssueID != nil {
		gitLabIssueID = req.GitlabIssueID
	}

	// Обрабатываем исполнителя (assigned_to): поле опционально
	var assignedTo *uuid.UUID = nil
	if req.AssignedTo != nil && *req.AssignedTo != "" {
		// Валидируем UUID до обращения к БД
		assignedUUID, err := uuid.Parse(*req.AssignedTo)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Некорректный идентификатор исполнителя",
			})
		}
		var assigner models.User
		result := dbConn.Session(&gorm.Session{}).First(&assigner, "id = ? AND deleted = ?", assignedUUID, false)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Исполнитель не найден",
				})
			}
			log.Printf("DB error (find assignee by id): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении исполнителя из базы данных",
			})
		}
		assignedTo = &assigner.ID
	}

	// Валидируем и получаем поручителя (creator_id)
	if req.CreatorID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Отсутствует идентификатор поручителя",
		})
	}
	creatorUUID, err := uuid.Parse(req.CreatorID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор поручителя",
		})
	}
	var creator models.User
	res := dbConn.Session(&gorm.Session{}).First(&creator, "id = ? AND deleted = ?", creatorUUID, false)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Поручитель исполнитель не найден",
			})
		}
		log.Printf("DB error (find creator by id): %v", res.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении поручителя задачи из базы данных",
		})
	}

	var creatorID *uuid.UUID = &creator.ID


	// Валидируем идентификатор проекта
	temp1, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	var category *int8
	if req.Category != nil {
		category = req.Category
	}

	del := false

	now := time.Now()

	task := models.Task{
		ID:            newUUID,
		Priority:      priority,
		Name:          name,
		Description:   description,
		Status:        status,
		CreatedBy:     creatorID,
		AssignedTo:    assignedTo,
		Deadline:      deadline,
		StartDate:     startDate,
		GitlabIssueID: gitLabIssueID,
		ProjectID:     temp1,
		Category:      category,
		Deleted: &del,
		CreatedAt: &now,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Create(&task); res.Error != nil {
			log.Printf("DB error (create task): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create task): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании задачи"})
	}

	createResp := response.TaskUniversaResponse{
		ID:      task.ID.String(),
		Message: "Задача создана",
	}

	return c.JSON(http.StatusCreated, createResp)
}

// UpdateTask godoc
// @Summary Обновление задачи
// @Description Обновляет данные задачи по её ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Param task body request.TaskUpdateRequest true "Данные для обновления задачи"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TaskUniversaResponse "Задача успешно обновлена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении задачи"
// @Router /task/{id} [patch]
func UpdateTask(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	taskID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор задачи"})
	}

	var req request.TaskUpdateRequest
	if err = c.Bind(&req); err != nil {
		log.Printf("Bind error (update task): %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные запроса"})
	}

	var updates = make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Deadline != nil {
		updates["deadline"] = *req.Deadline
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.GitlabIssueID != nil {
		updates["gitlab_issue_id"] = *req.GitlabIssueID
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.AssignedTo != nil {
		if *req.AssignedTo == "" {
			// Очистить назначенного исполнителя
			updates["assigned_to"] = nil
		} else {
			assignedUUID, err := uuid.Parse(*req.AssignedTo)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор исполнителя"})
			}
			updates["assigned_to"] = assignedUUID
		}
	}

	if len(updates) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Нет данных для обновления"})
	}

	updates["updated_at"] = time.Now()

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Task{}).Where("id = ? and deleted = ?", taskID, false).Updates(updates)
		if res.Error != nil {
			log.Printf("DB error (update task): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (update task): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении задачи"})
	}

	return c.JSON(http.StatusOK, response.TaskUniversaResponse{
		ID:      taskID.String(),
		Message: "Задача изменена",
	})
}

// DeleteTask godoc
// @Summary Удаление задачи
// @Description Логическое удаление задачи по ID, включая связанные отчеты (поле deleted = true)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TaskUniversaResponse "Задача успешно удалена"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении задачи"
// @Router /task/{id} [delete]
func DeleteTask(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(models.Task{}).Where("id = ?", id).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete task): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete task): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении задачи"})
	}

	return c.JSON(http.StatusOK, response.TaskUniversaResponse{
		ID:      id,
		Message: "Задача удалена",
	})
}
