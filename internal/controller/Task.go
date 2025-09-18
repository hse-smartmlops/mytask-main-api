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
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
)

func RegisterTaskRoutes(e *echo.Echo) {
	taskGroup := e.Group("/task")
	taskGroup.Use(KeycloakAuthMiddleware)
	{
		taskGroup.GET("/all/:page/:pagesize", getAllTasks)
		taskGroup.GET("/:id", getTaskByID)
		taskGroup.GET("/project/:projectId/:page/:pagesize", getTasksByProjectID)
		taskGroup.POST("", createTask)
		taskGroup.PATCH("/:id", updateTask)
		taskGroup.DELETE("/:id", deleteTask)
		taskGroup.GET("/user/:id/:page/:pagesize", getTasksByUserId)
	}
}

// getAllTasks godoc
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
func getAllTasks(c echo.Context) error {
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
	if err := dbConn.Session(&gorm.Session{}).Model(models.Task{}).
		Preload("StatusTasks", "deleted = ?", false).
		Preload("StatusTasks.Status", "deleted = ?", false).
		Where("deleted = ?", false).
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
    taskResponse := response.TaskShort{
        ID:        task.ID.String(),
        Name:      utils.GetString(task.Name),
        ProjectID: task.ProjectID.String(),
        Priority:  utils.GetInt16(task.Priority),
        StartDate: utils.GetTime(task.StartDate),
        Deadline:  utils.GetTime(task.Deadline),
        UpdatedAt: utils.GetTime(task.UpdatedAt),
    }

    for _, st := range task.StatusTasks {
        if st.Status != nil {
            taskResponse.Statuses = append(taskResponse.Statuses, response.StatusResponse{
                ID:        st.Status.ID.String(),
                Key:       utils.GetString(st.Status.Key),
                Name:      utils.GetString(st.Status.Name),
                Color:     utils.GetString(st.Status.Color),
                IsDefault: utils.GetBool(st.Status.IsDefault),
                IsActive:  utils.GetBool(st.Status.IsActive),
                IsOpen:    utils.GetBool(st.Status.IsOpen),
                CreatedAt: utils.GetTime(st.Status.CreatedAt),
                UpdatedAt: utils.GetTime(st.Status.UpdatedAt),
            })
        }
    }

    taskList.Tasks = append(taskList.Tasks, taskResponse)
}
	return c.JSON(http.StatusOK, taskList)
}

// getTaskByID godoc
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
func getTaskByID(c echo.Context) error {
    if err := authorize(c); err != nil {
        return err
    }

    taskID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Некорректный идентификатор задачи",
        })
    }

    var task models.Task
    // Подгружаем статусы и пользователей сразу
    if err := dbConn.Session(&gorm.Session{}).Model(models.Task{}).
        Preload("StatusTasks", "deleted = ?", false).
        Preload("StatusTasks.Status", "deleted = ?", false).
        Preload("CreatedByUser", "deleted = ?", false).
        Preload("AssignedToUser", "deleted = ?", false).
        Where("id = ? AND deleted = ?", taskID, false).
        First(&task).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
        }
        log.Printf("DB error (find task by id): %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задачи"})
    }

    // Формируем creator и assigner
    creatorInfo := response.UserShort{}
    if task.CreatedByUser != nil {
        creatorInfo = response.UserShort{
            ID:        task.CreatedByUser.ID.String(),
            FirstName: utils.GetString(task.CreatedByUser.FirstName),
            LastName:  utils.GetString(task.CreatedByUser.LastName),
        }
    }

    assignerInfo := response.UserShort{}
    if task.AssignedToUser != nil {
        assignerInfo = response.UserShort{
            ID:        task.AssignedToUser.ID.String(),
            FirstName: utils.GetString(task.AssignedToUser.FirstName),
            LastName:  utils.GetString(task.AssignedToUser.LastName),
        }
    }

	statuses := []response.StatusResponse{}
	for _, st := range task.StatusTasks {
		if st.Status != nil {
			statuses = append(statuses, response.StatusResponse{
				ID:        st.Status.ID.String(),
				Key:       utils.GetString(st.Status.Key),
				Name:      utils.GetString(st.Status.Name),
				Color:     utils.GetString(st.Status.Color),
				IsDefault: utils.GetBool(st.Status.IsDefault),
				IsActive:  utils.GetBool(st.Status.IsActive),
				IsOpen:    utils.GetBool(st.Status.IsOpen),
				CreatedAt: utils.GetTime(st.Status.CreatedAt),
				UpdatedAt: utils.GetTime(st.Status.UpdatedAt),
			})
		}
	}

    taskResponse := response.GetTaskByIDResponse{
        ID:         task.ID.String(),
        ProjectID:  task.ProjectID.String(),
        Name:       utils.GetString(task.Name),
        Description: utils.GetString(task.Description),
        Priority:   utils.GetInt16(task.Priority),
        CreatedBy:  creatorInfo,
        AssignedTo: assignerInfo,
        Deadline:   utils.GetTime(task.Deadline),
        StartDate:  utils.GetTime(task.StartDate),
        TimeSpent:  utils.GetString(task.TimeSpent),
        GitlabIssueID: utils.GetInt(task.GitlabIssueID),
        Category:   utils.GetInt8(task.Category),
        UpdatedAt:  utils.GetTime(task.UpdatedAt),
    	Statuses:      statuses,
    }

    return c.JSON(http.StatusOK, taskResponse)
}


// getTasksByProjectID godoc
// @Summary Получение задач по ID проекта
// @Description Получает список задач, связанных с указанным проектом
// @Tags Tasks
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TaskListResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проекта"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/project/{projectId}/{page}/{pagesize} [get]
func getTasksByProjectID(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	// projectID
	projectIDParam := c.Param("projectId")
	projectUUID, err := uuid.Parse(projectIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
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

	var tasks []models.Task
	if err := dbConn.Session(&gorm.Session{}).Model(models.Task{}).
		Preload("StatusTasks", "deleted = ?", false).
		Preload("StatusTasks.Status", "deleted = ?", false).
		Where("project_id = ? AND deleted = ?", projectUUID, false).
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		log.Printf("DB error (find tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении задач из базы данных",
		})
	}

	var totalCount int64
	if err := dbConn.Model(&models.Task{}).Session(&gorm.Session{}).
		Where("project_id = ? AND deleted = ?", projectUUID, false).
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count tasks): %v", err)
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
		if task.Deleted != nil && *task.Deleted {
			continue
		}

		taskResponse := response.TaskShort{
			ID:        task.ID.String(),
			Name:      utils.GetString(task.Name),
			ProjectID: task.ProjectID.String(),
			Priority:  utils.GetInt16(task.Priority),
			StartDate: utils.GetTime(task.StartDate),
			Deadline:  utils.GetTime(task.Deadline),
			UpdatedAt: utils.GetTime(task.UpdatedAt),
		}

		for _, st := range task.StatusTasks {
			if st.Status != nil {
				taskResponse.Statuses = append(taskResponse.Statuses, response.StatusResponse{
					ID:        st.Status.ID.String(),
					Key:       utils.GetString(st.Status.Key),
					Name:      utils.GetString(st.Status.Name),
					Color:     utils.GetString(st.Status.Color),
					IsDefault: utils.GetBool(st.Status.IsDefault),
					IsActive:  utils.GetBool(st.Status.IsActive),
					IsOpen:    utils.GetBool(st.Status.IsOpen),
					CreatedAt: utils.GetTime(st.Status.CreatedAt),
					UpdatedAt: utils.GetTime(st.Status.UpdatedAt),
				})
			}
		}

		taskList.Tasks = append(taskList.Tasks, taskResponse)
	}

	return c.JSON(http.StatusOK, taskList)
}

// createTask godoc
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
func createTask(c echo.Context) error {
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

	if req.AssignedTo == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Отсутствует идентификатор исполнителя",
		})
	}
	assignerUUID, err := uuid.Parse(*req.AssignedTo)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор исполнителя",
		})
	}

	assignerID := &assignerUUID

	var assigner models.User
	result := dbConn.Session(&gorm.Session{}).Model(models.User{}).First(&assigner, "id = ? AND deleted = ?", assignerUUID, false)
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

	// Валидируем и получаем поручителя (creator_id)
	if req.CreatorID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Отсутствует идентификатор поручителя",
		})
	}
	creatorUUID, err := uuid.Parse(*req.CreatorID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор поручителя",
		})
	}

	var creator models.User
	result = dbConn.Session(&gorm.Session{}).Model(models.User{}).First(&creator, "id = ? AND deleted = ?", creatorUUID, false)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Поручитель исполнитель не найден",
			})
		}
		log.Printf("DB error (find creator by id): %v", result.Error)
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
		Priority:      req.Priority,
		Name:          req.Name,
		Description:   req.Description,
		CreatedBy:     creatorID,
		AssignedTo:    assignerID,
		Deadline:      req.Deadline,
		StartDate:     req.StartDate,
		GitlabIssueID: req.GitlabIssueID,
		ProjectID:     temp1,
		Category:      category,
		Deleted: &del,
		CreatedAt: &now,
		StatusTasks: []models.StatusTask{},
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).Model(models.Task{}).Create(&task); res.Error != nil {
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

// updateTask godoc
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
func updateTask(c echo.Context) error {
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
		res := tx.Session(&gorm.Session{}).Model(models.Task{}).Where("id = ? and deleted = ?", taskID, false).Updates(updates)
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

// deleteTask godoc
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
func deleteTask(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.Task{}).Where("id = ?", id).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete task): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		if res = tx.Session(&gorm.Session{}).Model(models.StatusTask{}).Where("task_id = ?", id).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete status_task): %v", res.Error)
			return res.Error
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

// getTasksByUserId godoc
// @Summary Получение задач по ID пользователя
// @Description Получение списка задач, назначенных на конкретного пользователя, с пагинацией (deleted = false)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров или некорректный ID пользователя"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Success 200 {object} response.TaskListResponse "Список задач успешно получен"
// @Router /task/user/{id}/{page}/{pagesize} [get]
func getTasksByUserId(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	userIDParam := c.Param("id")
	userUUID, err := uuid.Parse(userIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор пользователя"})
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

	var tasks []models.Task
	if err := dbConn.Session(&gorm.Session{}).Model(models.Task{}).
		Preload("StatusTasks", "deleted = ?", false).
		Preload("StatusTasks.Status", "deleted = ?", false).
		Where("assigned_to = ? AND deleted = ?", userUUID, false).
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		log.Printf("DB error (find tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении задач из базы данных",
		})
	}

	var totalCount int64
	if err := dbConn.Model(&models.Task{}).Session(&gorm.Session{}).
		Where("assigned_to = ? AND deleted = ?", userUUID, false).
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count tasks): %v", err)
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
		if task.Deleted != nil && *task.Deleted {
			continue
		}

		taskResponse := response.TaskShort{
			ID:        task.ID.String(),
			Name:      utils.GetString(task.Name),
			ProjectID: task.ProjectID.String(),
			Priority:  utils.GetInt16(task.Priority),
			StartDate: utils.GetTime(task.StartDate),
			Deadline:  utils.GetTime(task.Deadline),
			UpdatedAt: utils.GetTime(task.UpdatedAt),
		}

		for _, st := range task.StatusTasks {
			if st.Status != nil {
				taskResponse.Statuses = append(taskResponse.Statuses, response.StatusResponse{
					ID:        st.Status.ID.String(),
					Key:       utils.GetString(st.Status.Key),
					Name:      utils.GetString(st.Status.Name),
					Color:     utils.GetString(st.Status.Color),
					IsDefault: utils.GetBool(st.Status.IsDefault),
					IsActive:  utils.GetBool(st.Status.IsActive),
					IsOpen:    utils.GetBool(st.Status.IsOpen),
					CreatedAt: utils.GetTime(st.Status.CreatedAt),
					UpdatedAt: utils.GetTime(st.Status.UpdatedAt),
				})
			}
		}

		taskList.Tasks = append(taskList.Tasks, taskResponse)
	}

	return c.JSON(http.StatusOK, taskList)
}