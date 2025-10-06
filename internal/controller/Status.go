package controller

import (
	"crypto/sha256"
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"emplacc-api/internal/utils"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterStatusRoutes(e *echo.Echo) {
	g := e.Group("/status")
	g.Use(KeycloakAuthMiddleware)
	{
		g.GET("/all/:page/:pagesize", GetAllStatuses)
		g.GET("/:id", GetStatusByID)
		g.POST("", CreateStatus)
		g.PATCH("/:id", UpdateStatus)
		g.DELETE("/:id", DeleteStatus)
		g.GET("/board/:board_id", GetStatusesByBoardId)
	}
}

// GetAllStatuses godoc
// @Summary Получение всех статусов
// @Description Получение списка всех статусов с пагинацией (только неудаленные), отсортированных по полю order
// @Tags Statuses
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.StatusListResponse "Список статусов"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении статусов"
// @Router /status/all/{page}/{pagesize} [get]
func GetAllStatuses(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге размера страницы"})
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Where("deleted = ?", false).
		Count(&total).Error; err != nil {
		log.Printf("DB error (count statuses): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте статусов"})
	}

	var rows []models.Status
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Where("deleted = ?", false).
		Order("sort_order ASC"). // ← ДОБАВЛЕНА СОРТИРОВКА
		Limit(pageSize).Offset(offset).
		Find(&rows).Error; err != nil {
		log.Printf("DB error (find statuses): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов"})
	}

	out := response.StatusListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: int64(total), // ← int(total), так как TotalCount int
	}
	return c.JSON(http.StatusOK, out)
}

// GetStatusesByBoardId godoc
// @Summary Получение статусов по ID доски
// @Description Получение списка статусов, связанных с доской по её ID, включая все задачи для каждого статуса
// @Tags Statuses
// @Produce json
// @Param board_id path string true "ID доски"
// @Security BearerAuth
// @Success 200 {object} response.StatusByBoardIdResponse "Список статусов для доски"
// @Failure 400 {object} map[string]string "Ошибка при парсинге ID доски"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении статусов"
// @Router /status/board/{board_id} [get]
func GetStatusesByBoardId(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	boardID, err := uuid.Parse(c.Param("board_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора доски"})
	}

	// Проверим, существует ли доска (опционально, но хорошо для 404)
	var count int64
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Board{}).
		Where("id = ? AND deleted = ?", boardID, false).
		Count(&count).Error; err != nil {
		log.Printf("DB error (check board existence): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке доски"})
	}
	if count == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Доска не найдена"})
	}

	// Загружаем статусы доски + задачи к каждому статусу
	var statuses []models.Status
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Where("board_id = ? AND deleted = ?", boardID, false).
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
		BoardId:  boardID.String(),
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

// GetStatusByID godoc
// @Summary Получение статуса по ID
// @Description Получение информации о статусе по его ID, включая все связанные задачи
// @Tags Statuses
// @Produce json
// @Param id path string true "ID статуса"
// @Security BearerAuth
// @Success 200 {object} response.StatusResponse "Информация о статусе"
// @Failure 400 {object} map[string]string "Некорректный идентификатор статуса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Статус не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении статуса"
// @Router /status/{id} [get]
func GetStatusByID(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	statusID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор статуса"})
	}

	var s models.Status
	if err := DBConn.Session(&gorm.Session{}).
		Where("id = ? AND deleted = ?", statusID, false).
		Preload("Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Статус не найден"})
		}
		log.Printf("DB error (find status by id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статуса"})
	}

	// Маппинг задач
	tasks := make([]response.TaskShort, 0, len(s.Tasks))
	for _, task := range s.Tasks {
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

	return c.JSON(http.StatusOK, response.StatusResponse{
		ID:        s.ID.String(),
		Key:       utils.GetString(s.Key),
		Name:      utils.GetString(s.Name),
		Color:     utils.GetString(s.Color),
		Order:     utils.GetInt(s.SortOrder), // ← добавлено
		IsDefault: utils.GetBool(s.IsDefault),
		IsActive:  utils.GetBool(s.IsActive),
		IsOpen:    utils.GetBool(s.IsOpen),
		CreatedAt: utils.GetTime(s.CreatedAt),
		UpdatedAt: utils.GetTime(s.UpdatedAt),
		Tasks:     tasks, // ← добавлено
	})
}

// CreateStatus godoc
// @Summary Создание статуса
// @Description Создание нового статуса
// @Tags Statuses
// @Accept json
// @Produce json
// @Param /status body request.CreateStatusRequest true "Данные для создания статуса"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при привязке данных"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.StatusUniversalResponse "Статус успешно создан"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании статуса"
// @Router /status [post]
func CreateStatus(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	var req request.CreateStatusRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при привязке данных запроса"})
	}

	id := uuid.New()
	now := time.Now()
	del := false
	h := sha256.Sum256(id[:])
	key := hex.EncodeToString(h[:])[:8]
	boardId, err := uuid.Parse(req.BoardId)
	if err != nil{
		log.Printf("Uuid parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при получении идентификатора доски"})
	}

	row := models.Status{
		ID:        id,
		Key:       &key,
		Name:      &req.Name,
		Color:     &req.Color,
		IsDefault: &req.IsDefault,
		IsActive:  &req.IsActive,
		IsOpen:    &req.IsOpen,
		CreatedAt: &now,
		Deleted:   &del,
		SortOrder: 	   &req.Order,
		BoardID:   boardId,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		return tx.Session(&gorm.Session{}).
			Model(&models.Status{}).
			Create(&row).Error
	}); txErr != nil {
		log.Printf("DB transaction error (create status): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании статуса"})
	}

	return c.JSON(http.StatusCreated, response.StatusUniversalResponse{
		ID:      id.String(),
		Message: "Статус успешно создан",
	})
}

// UpdateStatus godoc
// @Summary Обновление статуса
// @Description Обновление информации о статусе по ID
// @Tags Statuses
// @Accept json
// @Produce json
// @Param id path string true "ID статуса"
// @Param body body request.UpdateStatusRequest true "Данные для обновления статуса"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Нет данных для обновления или некорректный ID"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.StatusUniversalResponse "Статус успешно обновлен"
// @Failure 404 {object} map[string]string "Статус не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении статуса"
// @Router /status/{id} [patch]
func UpdateStatus(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	statusID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор статуса"})
	}

	var req request.UpdateStatusRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при привязке данных запроса"})
	}

	update := map[string]interface{}{}
	if req.Name != nil      { update["name"] = *req.Name }
	if req.Color != nil     { update["color"] = *req.Color }
	if req.IsDefault != nil { update["is_default"] = *req.IsDefault }
	if req.IsActive != nil  { update["is_active"]  = *req.IsActive }
	if req.IsOpen != nil    { update["is_open"]    = *req.IsOpen }
	if req.BoardId != nil   { update["board_id"]   = *req.BoardId}
	if req.Order != nil     { update["order"]      = *req.Order}
	if len(update) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Нет данных для обновления"})
	}
	now := time.Now()
	update["updated_at"] = &now

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.Status{}).
			Where("id = ? AND deleted = FALSE", statusID).
			Updates(update)
		if res.Error != nil { return res.Error }
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update status): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении статуса"})
	}

	return c.JSON(http.StatusOK, response.StatusUniversalResponse{
		ID:      statusID.String(),
		Message: "Статус успешно обновлён",
	})
}

// DeleteStatus godoc
// @Summary Удаление статуса
// @Description Логическое удаление статуса по ID, включая все связанные задачи (поле deleted = true)
// @Tags Statuses
// @Accept json
// @Produce json
// @Param id path string true "ID статуса"
// @Security BearerAuth
// @Success 200 {object} response.StatusUniversalResponse "Статус успешно удален"
// @Failure 400 {object} map[string]string "Некорректный идентификатор статуса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Статус не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении статуса"
// @Router /status/{id} [delete]
func DeleteStatus(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	statusID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор статуса"})
	}

	deleted := true
	now := time.Now()
	updateData := map[string]interface{}{
		"deleted":    &deleted,
		"updated_at": now,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		// 1. Проверяем, существует ли неудалённый статус
		var count int64
		if err := tx.Session(&gorm.Session{}).
			Model(&models.Status{}).
			Where("id = ? AND deleted = ?", statusID, false).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "Статус не найден или уже удалён"})
		}

		// 2. Удаляем задачи, привязанные к этому статусу
		if err := tx.Session(&gorm.Session{}).
			Model(&models.Task{}).
			Where("status_id = ?", statusID).
			Updates(updateData).Error; err != nil {
			return err
		}

		// 3. Удаляем сам статус
		if err := tx.Session(&gorm.Session{}).
			Model(&models.Status{}).
			Where("id = ?", statusID).
			Updates(updateData).Error; err != nil {
			return err
		}

		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete status): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении статуса"})
	}

	return c.JSON(http.StatusOK, response.StatusUniversalResponse{
		ID:      statusID.String(),
		Message: "Статус успешно удалён",
	})
}

/*
// AddStatusToTask godoc
// @Summary Добавление статуса к задаче
// @Description Создание связи между статусом и задачей
// @Tags Statuses
// @Accept json
// @Produce json
// @Param body body request.AddStatusToTaskRequest true "Данные для добавления статуса к задаче"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при привязке данных или парсинге UUID"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.AddStatusToTaskResponse "Статус успешно добавлен к задаче"
// @Failure 500 {object} map[string]string "Ошибка сервера при добавлении статуса к задаче"
// @Router /status/add-to-task [post]
func AddStatusToTask(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	var req request.AddStatusToTaskRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	taskID, err := uuid.Parse(req.TaskId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора задачи"})
	}
	statusID, err := uuid.Parse(req.StatusId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора статуса"})
	}

	del := false

	// ВАЖНО: StatusTask.Deleted — bool, CreatedAt/UpdatedAt — авто
	link := models.StatusTask{
		StatusID: statusID,
		TaskID:   taskID,
		Deleted:  &del,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		return tx.Session(&gorm.Session{}).
			Model(&models.StatusTask{}).
			Create(&link).Error
	}); txErr != nil {
		log.Printf("DB error (add status to task): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении статуса к задаче"})
	}

	return c.JSON(http.StatusCreated, response.AddStatusToTaskResponse{
		TaskId:   req.TaskId,
		StatusId: req.StatusId,
		Message:  "Статус успешно добавлен к задаче",
	})
}

// DeleteStatusFromTask godoc
// @Summary Удаление статуса из задачи
// @Description Удаление связи между статусом и задачей (мягкое удаление)
// @Tags Statuses
// @Produce json
// @Param task_id path string true "ID задачи"
// @Param status_id path string true "ID статуса"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге UUID"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Связь не найдена или уже удалена"
// @Success 200 {object} response.StatusUniversalResponse "Статус успешно удалён из задачи"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении статуса из задачи"
// @Router /status/delete-from-task/{task_id}/{status_id} [delete]
func DeleteStatusFromTask(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	taskIDStr := c.Param("task_id")
	statusIDStr := c.Param("status_id")

	taskID, err := uuid.Parse(taskIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора задачи"})
	}
	statusID, err := uuid.Parse(statusIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора статуса"})
	}

	now := time.Now()
	del := true
	updateData := map[string]any{
		"deleted":    &del,
		"updated_at": &now,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.StatusTask{}).Where("status_id = ? and task_id = ? and deleted = FALSE", statusID, taskID).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete status-task): %v", res.Error)
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Связь не найдена или уже удалена"})
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete status-task): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении статус у задачи"})
	}

	return c.JSON(http.StatusOK, response.StatusUniversalResponse{
		ID:      statusID.String(),
		Message: "Статус задачи успешно удалён",
	})
}

// DeleteStatusFromBoard godoc
// @Summary Удаление статуса с доски
// @Description Удаление связи между статусом и доской (мягкое удаление)
// @Tags Statuses
// @Produce json
// @Param board_id path string true "ID доски"
// @Param status_id path string true "ID статуса"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге UUID"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Связь не найдена или уже удалена"
// @Success 200 {object} response.StatusUniversalResponse "Статус успешно удалён с доски"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении статуса с доски"
// @Router /status/delete-from-board/{board_id}/{status_id} [delete]
func DeleteStatusFromBoard(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	boardIDStr := c.Param("board_id")
	statusIDStr := c.Param("status_id")

	boardID, err := uuid.Parse(boardIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора доски"})
	}
	statusID, err := uuid.Parse(statusIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора статуса"})
	}

	now := time.Now()
	del := true
	updateData := map[string]any{
		"deleted":    &del,
		"updated_at": &now,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.StatusBoard{}).Where("status_id = ? and board_id = ? and deleted = FALSE", statusID, boardID).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete status-board): %v", res.Error)
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Связь не найдена или уже удалена"})
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete status-board): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении статус у доски"})
	}

	return c.JSON(http.StatusOK, response.StatusUniversalResponse{
		ID:      statusID.String(),
		Message: "Статус доски успешно удалён",
	})
}

// AddStatusToBoard godoc
// @Summary Добавление статуса к доске
// @Description Создание связи между статусом и доской
// @Tags Statuses
// @Accept json
// @Produce json
// @Param body body request.AddStatusToBoardRequest true "Данные для добавления статуса к доске"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при привязке данных или парсинге UUID"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.AddStatusToBoardResponse "Статус успешно добавлен к доске"
// @Failure 500 {object} map[string]string "Ошибка сервера при добавлении статуса к доске"
// @Router /status/add-to-board [post]
func AddStatusToBoard(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	var req request.AddStatusToBoardRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	boardID, err := uuid.Parse(req.BoardId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора доски"})
	}
	statusID, err := uuid.Parse(req.StatusId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора статуса"})
	}

	del := false
	now := time.Now()
	link := models.StatusBoard{
		StatusID:  statusID,
		BoardID:   boardID,
		Deleted:   &del,
		CreatedAt: &now,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		return tx.Session(&gorm.Session{}).
			Model(&models.StatusBoard{}).
			Create(&link).Error
	}); txErr != nil {
		log.Printf("DB error (add status to board): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при добавлении статуса к доске"})
	}

	return c.JSON(http.StatusCreated, response.AddStatusToBoardResponse{
		BoardId:  req.BoardId,
		StatusId: req.StatusId,
		Message:  "Статус успешно добавлен к доске",
	})
}

func CreateStartStatuses() error {
	var rows []models.Status
	if err := DBConn.Session(&gorm.Session{}).Model(&models.Status{}).Where("deleted = FALSE AND (name = ? OR name = ?)", "Открыта", "Закрыта").Find(&rows).Error; err != nil {
		log.Printf("DB error (find statuses): %v", err)
		return err
	}

	existing := make(map[string]uuid.UUID)
	for _, v := range rows {
		if v.Name == nil {
			continue
		}
		existing[*v.Name] = v.ID
	}

	var statusesToCreate []models.Status

	if _, hasOpen := existing["Открыта"]; !hasOpen {
		statusesToCreate = append(statusesToCreate, newStatus("Открыта", "#008000", true))
	}
	if _, hasClosed := existing["Закрыта"]; !hasClosed {
		statusesToCreate = append(statusesToCreate, newStatus("Закрыта", "#FF0000", false))
	}

	if id, ok := existing["Открыта"]; ok {
		BaseStartStatus = id
	} else if len(statusesToCreate) > 0 {
		for _, s := range statusesToCreate {
			if s.Name != nil && *s.Name == "Открыта" {
				BaseStartStatus = s.ID
				break
			}
		}
	}

	if id, ok := existing["Закрыта"]; ok {
		BaseEndStatus = id
	} else if len(statusesToCreate) > 0 {
		for _, s := range statusesToCreate {
			if s.Name != nil && *s.Name == "Закрыта" {
				BaseEndStatus = s.ID
				break
			}
		}
	}

	if len(statusesToCreate) == 0 {
		return nil
	}

	if err := DBConn.Transaction(func(tx *gorm.DB) error {
		return tx.Session(&gorm.Session{}).Create(&statusesToCreate).Error
	}); err != nil {
		log.Printf("DB transaction error (create statuses): %v", err)
		return err
	}

	return nil
}

func newStatus(name, color string, isOpen bool) models.Status {
	id := uuid.New()
	now := time.Now()
	tr := true
	fal := false
	h := sha256.Sum256(id[:])
	key := hex.EncodeToString(h[:])[:8]

	return models.Status{
		ID:        id,
		Key:       &key,
		Name:      &name,
		Color:     &color,
		IsDefault: &tr,
		IsActive:  &tr,
		IsOpen:    &isOpen,
		CreatedAt: &now,
		Deleted:   &fal,
	}
}*/
