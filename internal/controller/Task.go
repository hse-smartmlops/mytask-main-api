package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"errors"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func RegisterTaskRoutes(e *echo.Echo) {
	taskGroup := e.Group("/task")
	{
		taskGroup.GET("", GetAllTasks)
		taskGroup.GET("/:id", GetTaskByID)
		taskGroup.GET("/project/:projectId", GetTasksByProjectID)
		taskGroup.GET("/filter", GetTasksByFilter)
		taskGroup.POST("", CreateTask)
		taskGroup.PUT("/:id", UpdateTask)
		taskGroup.DELETE("/:id", DeleteTask)
	}
}

// GetAllTasks godoc
// @Summary Получение списка задач
// @Description Возвращает список всех задач (заглушка)
// @Tags Tasks
// @Produce json
// @Success 200 {object} response.TaskList
// @Router /tasks [get]
func GetAllTasks(c echo.Context) error {
	var req request.TaskListRequest
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

	taskList := response.TaskListResponce{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, task := range tasks {
		var id string
		if task.ID != nil {
			id = task.ID.String()
		}

		var projectId string
		if task.ID != nil {
			projectId = task.ProjectID.String()
		}

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
// @Description Возвращает задачу по ID (заглушка)
// @Tags Tasks
// @Param id path string true "ID задачи"
// @Produce json
// @Success 200 {object} response.TaskDetail
// @Failure 404 {object} response.ErrorResponse
// @Router /tasks/{id} [get]
func GetTaskByID(c echo.Context) error {
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
			var creatorID string
			if creator.ID != nil {
				creatorID = creator.ID.String()
			}

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
			var assignerID string
			if assigner.ID != nil {
				assignerID = assigner.ID.String()
			}

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

	var Id string
	if task.ID != nil {
		Id = task.ID.String()
	}

	var projectId string
	if task.ID != nil {
		projectId = task.ProjectID.String()
	}

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
	taskResponse := response.GetTaskByIDResponce{
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
// @Summary Получение задач по проекту
// @Description Возвращает задачи для указанного проекта (заглушка)
// @Tags Tasks
// @Param projectId path string true "ID проекта"
// @Produce json
// @Success 200 {object} response.TaskList
// @Router /tasks/project/{projectId} [get]
func GetTasksByProjectID(c echo.Context) error {
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
	if err = dbConn.Session(&gorm.Session{}).Where("project_id = ? AND deleted = ?", projectUUID).
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

	taskList := response.TaskListResponce{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, task := range tasks {
		if task.Deleted != nil {
			if !*task.Deleted {
				var id string
				if task.ID != nil {
					id = task.ID.String()
				}

				var projectId string
				if task.ID != nil {
					projectId = task.ProjectID.String()
				}

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

// GetTasksByFilter godoc
// @Summary Фильтрация задач
// @Description Возвращает отфильтрованный список задач (заглушка)
// @Tags Tasks
// @Param project_id query string false "ID проекта"
// @Param status query string false "Статус задачи"
// @Param priority query int false "Приоритет задачи"
// @Produce json
// @Success 200 {object} response.TaskList
// @Router /tasks/filter [get]
func GetTasksByFilter(c echo.Context) error {
	filter := new(request.TaskFilter)
	if err := c.Bind(filter); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid filter parameters"})
	}
	return c.JSON(http.StatusOK, nil)
}

// CreateTask godoc
// @Summary Создание задачи
// @Description Создает новую задачу (заглушка)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param input body request.CreateTask true "Данные задачи"
// @Success 201 {object} response.TaskOperation
// @Failure 400 {object} response.ErrorResponse
// @Router /tasks [post]
func CreateTask(c echo.Context) error {
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

	var assigner models.User
	result := dbConn.Session(&gorm.Session{}).First(&assigner, "id = ? AND deleted = ?", req.AssignedTo, false)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Исполнитель не найден",
			})
		}
		log.Printf("DB error (find project by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении исполнителя из базы данных",
		})
	}

	var assignedTo *uuid.UUID
	if assigner.ID != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Данный поручитель удален",
		})
	} else {
		assignedTo = assigner.ID
	}

	var creator models.User
	result = dbConn.Session(&gorm.Session{}).First(&creator, "id = ? AND deleted = ?", req.CreatorID, false)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Поручитель исполнитель не найден",
			})
		}
		log.Printf("DB error (find project by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении поручителя задачи из базы данных",
		})
	}

	var creatorID *uuid.UUID
	if creator.ID != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Данный поручитель удален",
		})
	} else {
		creatorID = creator.ID
	}

	var projectId *uuid.UUID
	temp1, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{})
	}
	projectId = &temp1

	var category *int8
	if req.Category != nil {
		category = req.Category
	}

	task := models.Task{
		ID:            &newUUID,
		Priority:      priority,
		Name:          name,
		Description:   description,
		Status:        status,
		CreatedBy:     creatorID,
		AssignedTo:    assignedTo,
		Deadline:      deadline,
		StartDate:     startDate,
		GitlabIssueID: gitLabIssueID,
		ProjectID:     projectId,
		Category:      category,
	}

	result = dbConn.Session(&gorm.Session{}).Create(&task)
	if result.Error != nil {
		log.Printf("DB error (create task): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{})
	}

	createResp := response.TaskUniversaResponce{
		ID:      task.ID.String(),
		Message: "Задача создана",
	}

	return c.JSON(http.StatusCreated, createResp)
}

// UpdateTask godoc
// @Summary Обновление задачи
// @Description Обновляет существующую задачу (заглушка)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Param input body request.UpdateTask true "Обновляемые данные"
// @Success 200 {object} response.TaskOperation
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /tasks/{id} [put]
func UpdateTask(c echo.Context) error {
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
	if req.Community != nil {
		updates["community"] = *req.Community
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

	result := dbConn.Session(&gorm.Session{}).Model(&models.Task{}).Where("id = ?", taskID).Updates(updates)
	if result.Error != nil {
		log.Printf("DB error (update task): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении задачи"})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
	}

	return c.JSON(http.StatusOK, response.TaskUniversaResponce{
		ID:      taskID.String(),
		Message: "Задача изменена",
	})
}

// DeleteTask godoc
// @Summary Удаление задачи
// @Description Удаляет задачу по ID (заглушка)
// @Tags Tasks
// @Param id path string true "ID задачи"
// @Produce json
// @Success 200 {object} response.TaskOperation
// @Failure 404 {object} response.ErrorResponse
// @Router /tasks/{id} [delete]
func DeleteTask(c echo.Context) error {
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	updateData["updated_at"] = time.Now()
	result := dbConn.Session(&gorm.Session{}).Model(models.Task{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete task): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении задачи",
		})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не удалено",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.DailyReport{}).Where("task_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete report: %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении отчета",
		})
	}

	return c.JSON(http.StatusOK, response.TaskUniversaResponce{
		ID:      id,
		Message: "Задача удалена",
	})
}
