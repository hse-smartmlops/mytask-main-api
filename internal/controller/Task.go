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
	"gorm.io/gorm/clause"

	"github.com/labstack/echo/v4"
)

func RegisterTaskRoutes(e *echo.Echo) {
	taskGroup := e.Group("/task")
	taskGroup.Use(KeycloakAuthMiddleware)
	{
		taskGroup.GET("/all/:page/:pagesize", GetAllTasks)
		taskGroup.GET("/:id", GetTaskByID)
		taskGroup.GET("/project/:projectId/:page/:pagesize", GetTasksByProjectID)
		taskGroup.POST("", CreateTask)
		taskGroup.PATCH("/:id", UpdateTask)
		taskGroup.DELETE("/:id", DeleteTask)
		taskGroup.GET("/user/:id/:page/:pagesize", GetTasksByUserId)
		taskGroup.POST("/move", TaskMoveFunc)
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
// @Success 200 {object} response.TaskListResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/all/{page}/{pagesize} [get]
func GetAllTasks(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Task{}).
		Where("deleted = ?", false).
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте задач"})
	}

	var tasks []models.Task
	if err := DBConn.Session(&gorm.Session{}).
		Where("deleted = ?", false).
		// НЕ нужно Preload("Status"), потому что TaskShort не включает Status — только StatusID
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		log.Printf("DB error (find tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задач"})
	}

	taskList := response.TaskListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		Tasks:      make([]response.TaskShort, 0, len(tasks)),
	}

	for _, task := range tasks {
		taskList.Tasks = append(taskList.Tasks, response.TaskShort{
			ID:        task.ID.String(),
			StatusID:  task.StatusID.String(), // ← uuid.UUID → string
			Name:      utils.GetString(task.Name),
			Priority:  utils.GetInt16(task.Priority),
			StartDate: utils.GetTime(task.StartDate),   // ← важно: обработка nil
			Deadline:  utils.GetTime(task.Deadline),    // ← важно: обработка nil
			CreatedAt: utils.GetTime(task.CreatedAt),   // ← важно
			UpdatedAt: utils.GetTime(task.UpdatedAt),   // ← важно
		})
	}

	return c.JSON(http.StatusOK, taskList)
}

// GetTaskByID godoc
// @Summary Получение задачи по ID
// @Description Получает данные задачи по её уникальному идентификатору, включая статус и пользователей
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Security BearerAuth
// @Success 200 {object} response.GetTaskByIDResponse "Задача успешно получена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор задачи"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задачи"
// @Router /task/{id} [get]
func GetTaskByID(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор задачи",
		})
	}

	var task models.Task
	if err := DBConn.Session(&gorm.Session{}).
		Where("id = ? AND deleted = ?", taskID, false).
		Preload("Status", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		Preload("CreatedByUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		Preload("AssignedToUser", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}
		log.Printf("DB error (find task by id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задачи"})
	}

	// Создатель
	creatorInfo := response.UserShort{}
	if task.CreatedByUser != nil {
		creatorInfo = response.UserShort{
			ID:        task.CreatedByUser.ID.String(),
			FirstName: task.CreatedByUser.FirstName,
			LastName:  task.CreatedByUser.LastName,
		}
	}

	// Исполнитель
	var assignerInfo response.UserShort
	if task.AssignedToUser != nil {
		assignerInfo = response.UserShort{
			ID:        task.AssignedToUser.ID.String(),
			FirstName: task.AssignedToUser.FirstName,
			LastName:  task.AssignedToUser.LastName,
		}
	}

	taskResponse := response.GetTaskByIDResponse{
		ID:            task.ID.String(),
		StatusID:      task.StatusID.String(), // ← только ID статуса
		Name:          utils.GetString(task.Name),
		Description:   utils.GetString(task.Description),
		Priority:      utils.GetInt16(task.Priority),
		CreatedBy:     creatorInfo,
		AssignedTo:    assignerInfo, // ← указатель, чтобыomitempty работал
		Deadline:      utils.GetTime(task.Deadline),
		TimeSpent:     utils.GetString(task.TimeSpent),
		StartDate:     utils.GetTime(task.StartDate),
		GitlabIssueID: utils.GetInt(task.GitlabIssueID),
		Category:      utils.GetInt8(task.Category),
		UpdatedAt:     utils.GetTime(task.UpdatedAt),
		CreatedAt:     utils.GetTime(task.CreatedAt),
	}

	return c.JSON(http.StatusOK, taskResponse)
}


// GetTasksByProjectID godoc
// @Summary Получение задач по ID проекта
// @Description Получает список задач, связанных с указанным проектом через доски и статусы
// @Tags Tasks
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.TaskListResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проекта"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/project/{projectId}/{page}/{pagesize} [get]
func GetTasksByProjectID(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	// Подзапрос: все status_id, принадлежащие доскам этого проекта
	var statusIDs []uuid.UUID
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Select("statuses.id").
		Joins("JOIN boards ON statuses.board_id = boards.id").
		Where("boards.project_id = ? AND boards.deleted = ? AND statuses.deleted = ?", projectID, false, false).
		Scan(&statusIDs).Error; err != nil {
		log.Printf("DB error (get status IDs): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов проекта"})
	}

	var totalCount int64
	if len(statusIDs) == 0 {
		totalCount = 0
	} else {
		if err := DBConn.Session(&gorm.Session{}).
			Model(&models.Task{}).
			Where("status_id IN ? AND deleted = ?", statusIDs, false).
			Count(&totalCount).Error; err != nil {
			log.Printf("DB error (count tasks): %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте задач"})
		}
	}

	var tasks []models.Task
	if len(statusIDs) > 0 {
		if err := DBConn.Session(&gorm.Session{}).
			Where("status_id IN ? AND deleted = ?", statusIDs, false).
			Limit(pageSize).
			Offset(offset).
			Find(&tasks).Error; err != nil {
			log.Printf("DB error (find tasks): %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задач"})
		}
	}

	taskList := response.TaskListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		Tasks:      make([]response.TaskShort, 0, len(tasks)),
	}

	for _, task := range tasks {
		taskList.Tasks = append(taskList.Tasks, response.TaskShort{
			ID:        task.ID.String(),
			StatusID:  task.StatusID.String(),
			Name:      utils.GetString(task.Name),
			Priority:  utils.GetInt16(task.Priority),
			StartDate: utils.GetTime(task.StartDate),
			Deadline:  utils.GetTime(task.Deadline),
			CreatedAt: utils.GetTime(task.CreatedAt),
			UpdatedAt: utils.GetTime(task.UpdatedAt),
		})
	}

	return c.JSON(http.StatusOK, taskList)
}

// CreateTask godoc
// @Summary Создание новой задачи
// @Description Создает новую задачу с указанными параметрами
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task body request.TaskCreateRequest true "Данные для создания задачи"
// @Security BearerAuth
// @Success 201 {object} response.TaskUniversaResponse "Задача успешно создана"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Исполнитель, поручитель или статус не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании задачи"
// @Router /task [post]
func CreateTask(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	var req request.TaskCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	// === Валидация и парсинг AssignedTo (исполнитель) ===
	if req.AssignedTo == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Отсутствует идентификатор исполнителя"})
	}
	assigneeID, err := uuid.Parse(*req.AssignedTo)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор исполнителя"})
	}

	// Проверка существования исполнителя
	var assigneeCount int64
	if err := DBConn.Session(&gorm.Session{}).Model(&models.User{}).
		Where("id = ? AND deleted = ?", assigneeID, false).
		Count(&assigneeCount).Error; err != nil {
		log.Printf("DB error (check assignee): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке исполнителя"})
	}
	if assigneeCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Исполнитель не найден"})
	}

	// === Валидация и парсинг CreatorID (поручитель) ===
	if req.CreatorID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Отсутствует идентификатор поручителя"})
	}
	creatorID, err := uuid.Parse(*req.CreatorID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор поручителя"})
	}

	// Проверка существования поручителя
	var creatorCount int64
	if err := DBConn.Session(&gorm.Session{}).Model(&models.User{}).
		Where("id = ? AND deleted = ?", creatorID, false).
		Count(&creatorCount).Error; err != nil {
		log.Printf("DB error (check creator): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке поручителя"})
	}
	if creatorCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Поручитель не найден"})
	}

	// === Валидация StatusID ===
	statusID, err := uuid.Parse(req.StatusID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор статуса"})
	}

	// Проверка существования статуса и его принадлежности к неудалённой доске
	var statusExists bool
	if err := DBConn.Session(&gorm.Session{}).Model(&models.Status{}).
		Joins("INNER JOIN boards ON statuses.board_id = boards.id").
		Where("statuses.id = ? AND statuses.deleted = ? AND boards.deleted = ?", statusID, false, false).
		Scan(&statusExists).Error; err != nil {
		log.Printf("DB error (check status): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке статуса"})
	}
	if !statusExists {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Статус не найден или привязан к удалённой доске"})
	}

	// === Создание задачи ===
	now := time.Now()
	deleted := false

	task := models.Task{
		ID:            uuid.New(),
		StatusID:      statusID,
		Priority:      req.Priority,
		Name:          req.Name,
		Description:   req.Description,
		CreatedBy:     &creatorID,
		AssignedTo:    &assigneeID,
		Deadline:      req.Deadline,
		StartDate:     req.StartDate,
		GitlabIssueID: req.GitlabIssueID,
		Category:      req.Category,
		Deleted:       &deleted,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}

	if err := DBConn.Session(&gorm.Session{FullSaveAssociations: false}).
		Omit(clause.Associations).
		Create(&task).Error; err != nil {
		log.Printf("DB error (create task): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании задачи"})
	}

	return c.JSON(http.StatusCreated, response.TaskUniversaResponse{
		ID:      task.ID.String(),
		Message: "Задача создана",
	})
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
	if err := Authorize(c); err != nil {
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

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.Task{}).Where("id = ? AND deleted = FALSE", taskID).Updates(updates)
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
	if err := Authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	updateData := map[string]interface{}{"deleted": true}
	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.Task{}).Where("id = ?", id).Updates(updateData)
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

// GetTasksByUserId godoc
// @Summary Получение задач по ID пользователя
// @Description Получение списка задач, назначенных на конкретного пользователя, с пагинацией (deleted = false)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.TaskListResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров или некорректный ID пользователя"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/user/{id}/{page}/{pagesize} [get]
func GetTasksByUserId(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор пользователя"})
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Task{}).
		Where("assigned_to = ? AND deleted = ?", userID, false).
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте задач"})
	}

	var tasks []models.Task
	if err := DBConn.Session(&gorm.Session{}).
		Where("assigned_to = ? AND deleted = ?", userID, false).
		// НЕ нужно Preload("Status"), потому что TaskShort содержит только StatusID
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		log.Printf("DB error (find tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задач"})
	}

	taskList := response.TaskListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		Tasks:      make([]response.TaskShort, 0, len(tasks)),
	}

	for _, task := range tasks {
		taskList.Tasks = append(taskList.Tasks, response.TaskShort{
			ID:        task.ID.String(),
			StatusID:  task.StatusID.String(), // ← uuid.UUID → string
			Name:      utils.GetString(task.Name),
			Priority:  utils.GetInt16(task.Priority),
			StartDate: utils.GetTime(task.StartDate),
			Deadline:  utils.GetTime(task.Deadline),
			CreatedAt: utils.GetTime(task.CreatedAt),
			UpdatedAt: utils.GetTime(task.UpdatedAt),
		})
	}

	return c.JSON(http.StatusOK, taskList)
}

// TaskMoveFunc godoc
// @Summary Переместить задачу в другой статус (столбец)
// @Description Перемещает задачу в указанный статус на той же доске
// @Tags Tasks
// @Accept json
// @Produce json
// @Param request body request.MoveTaskToAnotherStatus true "Данные для перемещения задачи"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Задача успешно перемещена"
// @Failure 400 {object} map[string]string "Некорректные данные запроса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Задача или статус не найдены"
// @Failure 409 {object} map[string]string "Нельзя переместить задачу в статус с другой доски"
// @Failure 500 {object} map[string]string "Ошибка сервера при перемещении задачи"
// @Router /task/move [post]
func TaskMoveFunc(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	var req request.MoveTaskToAnotherStatus
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
	}

	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID задачи"})
	}

	toStatusID, err := uuid.Parse(req.ToStatusID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID статуса"})
	}

	// 1. Получаем задачу с её текущим статусом и доской
	var task models.Task
	if err := DBConn.Session(&gorm.Session{}).Model(&models.Task{}).
		Select("id, status_id, deleted").
		Where("id = ? AND deleted = ?", taskID, false).
		First(&task).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}
		log.Printf("DB error (fetch task): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задачи"})
	}

	// 2. Получаем текущий статус задачи — чтобы узнать board_id
	var currentStatus models.Status
	if err := DBConn.Session(&gorm.Session{}).Model(&models.Status{}).
		Select("board_id").
		Where("id = ? AND deleted = ?", task.StatusID, false).
		First(&currentStatus).Error; err != nil {
		log.Printf("DB error (fetch current status): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Не удалось определить доску задачи"})
	}

	// 3. Проверяем, существует ли целевой статус и принадлежит ли он той же доске
	var targetStatus models.Status
	if err := DBConn.Session(&gorm.Session{}).Model(&models.Status{}).
		Select("id, board_id").
		Where("id = ? AND deleted = ?", toStatusID, false).
		First(&targetStatus).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Целевой статус не найден"})
		}
		log.Printf("DB error (fetch target status): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке целевого статуса"})
	}

	// 4. Запрещаем перемещение между разными досками
	if currentStatus.BoardID != targetStatus.BoardID {
		return c.JSON(http.StatusConflict, map[string]string{"error": "Нельзя переместить задачу в статус с другой доски"})
	}

	// 5. Обновляем статус задачи
	updatedAt := time.Now()
	if err := DBConn.Session(&gorm.Session{}).Model(&models.Task{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status_id":   toStatusID,
			"updated_at":  updatedAt,
		}).Error; err != nil {
		log.Printf("DB error (update task status): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении статуса задачи"})
	}

	// Загружаем статусы доски + задачи к каждому статусу
	var statuses []models.Status
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Where("board_id = ? AND deleted = ?", targetStatus.BoardID, false).
		Order("sort_order ASC").
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		Find(&statuses).Error; err != nil {
		log.Printf("DB error (statuses by board): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов для доски"})
	}

	// Формируем ответ
	resp := response.StatusByBoardIdResponse{
		BoardId:  targetStatus.BoardID.String(),
		Statuses: make([]response.StatusResponse, 0, len(statuses)),
	}

	for _, status := range statuses {
		tasks := make([]response.TaskShort, 0, len(status.Tasks))
		for _, task := range status.Tasks {
			tasks = append(tasks, response.TaskShort{
				ID:          task.ID.String(),
				Name:        utils.GetString(task.Name),
				StatusID:    task.StatusID.String(),
				Priority:    utils.GetInt16(task.Priority),
				CreatedAt:   utils.GetTime(task.CreatedAt),
				UpdatedAt:   utils.GetTime(task.UpdatedAt),
				StartDate:   utils.GetTime(task.StartDate),
				Deadline:    utils.GetTime(task.Deadline),
			})
		}

		resp.Statuses = append(resp.Statuses, response.StatusResponse{
			ID:        status.ID.String(),
			Key:       utils.GetString(status.Key),
			Name:      utils.GetString(status.Name),
			Color:     utils.GetString(status.Color),
			Order:     utils.GetInt(status.SortOrder), // ← не забудь!
			IsDefault: utils.GetBool(status.IsDefault),
			IsActive:  utils.GetBool(status.IsActive),
			IsOpen:    utils.GetBool(status.IsOpen),
			CreatedAt: utils.GetTime(status.CreatedAt),
			UpdatedAt: utils.GetTime(status.UpdatedAt),
			Tasks:     tasks, // ← задачи внутри статуса
		})
	}

	return c.JSON(http.StatusOK, resp)
}
