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
)

func RegisterBoardRoutes(e *echo.Echo) {
	projectGroup := e.Group("/boards")
	projectGroup.Use(KeycloakAuthMiddleware)
	{
		projectGroup.GET("/all/:page/:pagesize", GetAllBoards)
		projectGroup.GET("/:id", GetBoardById)
		projectGroup.GET("/project/:projectId", GetBoardByProjectId)
		projectGroup.POST("", CreateBoard)
		projectGroup.PATCH("/:id", UpdateBoard)
		projectGroup.DELETE("/:id", DeleteBoard)
	}
}

// GetAllBoards godoc
// @Summary Получение списка всех досок
// @Description Получает список всех досок с учетом пагинации, исключая удаленные.
// @Tags Boards
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.BoardListResponse "Список досок успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении досок"
// @Router /boards/all/{page}/{pagesize} [get]
func GetAllBoards(c echo.Context) error {
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

	var total int64
	if err := DBConn.Session(&gorm.Session{}).Model(&models.Board{}).
		Where("deleted = ?", false).
		Count(&total).Error; err != nil {
		log.Printf("DB error (count boards): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте досок"})
	}

	var boards []models.Board
	if err := DBConn.Session(&gorm.Session{}).
		Where("deleted = ?", false).
		Order("created_at DESC NULLS LAST").
		Limit(pageSize).Offset(offset).
		Preload("Statuses", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false).Order("sort_order ASC")
		}).
		Preload("Statuses.Tasks", "deleted = ?", false).
		Find(&boards).Error; err != nil {
		log.Printf("DB error (find boards with preloads): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении досок"})
	}

	// Теперь формируем ответ
	boardList := response.BoardListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: int(total),
		Boards:     make([]response.BoardResponse, 0, len(boards)),
	}

	for _, board := range boards {
		statuses := make([]response.StatusResponse, 0, len(board.Statuses))
		for _, status := range board.Statuses {
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
					// Добавь другие поля, которые нужны в TaskResponse
				})
			}

			statuses = append(statuses, response.StatusResponse{
				ID:        status.ID.String(),
				Key:       utils.GetString(status.Key),
				Name:      utils.GetString(status.Name),
				Color:     utils.GetString(status.Color),
				IsDefault: utils.GetBool(status.IsDefault),
				IsActive:  utils.GetBool(status.IsActive),
				IsOpen:    utils.GetBool(status.IsOpen),
				CreatedAt: utils.GetTime(status.CreatedAt),
				UpdatedAt: utils.GetTime(status.UpdatedAt),
				Order:     utils.GetInt(status.SortOrder),
				Tasks:     tasks, // ← теперь задачи внутри статуса
			})
		}

		boardList.Boards = append(boardList.Boards, response.BoardResponse{
			Id:          board.ID.String(),
			ProjectId:   board.ProjectID.String(),
			Name:        utils.GetString(board.Name),
			Description: utils.GetString(board.Description),
			UpdatedAt:   utils.GetTime(board.UpdatedAt),
			CreatedAt:   utils.GetTime(board.CreatedAt),
			Statuses:    statuses,
		})
	}

	return c.JSON(http.StatusOK, boardList)
}

// GetBoardById godoc
// @Summary Получение доски по ID
// @Description Получает данные доски по её уникальному идентификатору
// @Tags Boards
// @Accept json
// @Produce json
// @Param id path string true "ID доски"
// @Security BearerAuth
// @Success 200 {object} response.BoardResponse "Доска успешно получена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Доска не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении доски"
// @Router /boards/{id} [get]
func GetBoardById(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор доски"})
	}

	var board models.Board
	if err := DBConn.Session(&gorm.Session{}).
		Where("id = ? AND deleted = ?", boardID, false).
		Preload("Statuses", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false).Order("sort_order ASC")
		}).
		Preload("Statuses.Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		First(&board).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Доска не найдена"})
		}
		log.Printf("DB error (find board by id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении доски"})
	}

	// Формируем ответ
	statuses := make([]response.StatusResponse, 0, len(board.Statuses))
	for _, status := range board.Statuses {
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

		statuses = append(statuses, response.StatusResponse{
			ID:        status.ID.String(),
			Key:       utils.GetString(status.Key),
			Name:      utils.GetString(status.Name),
			Color:     utils.GetString(status.Color),
			Order:     utils.GetInt(status.SortOrder), // ← ДОБАВЛЕНО
			IsDefault: utils.GetBool(status.IsDefault),
			IsActive:  utils.GetBool(status.IsActive),
			IsOpen:    utils.GetBool(status.IsOpen),
			CreatedAt: utils.GetTime(status.CreatedAt),
			UpdatedAt: utils.GetTime(status.UpdatedAt),
			Tasks:     tasks,
		})
	}

	boardResponse := response.BoardResponse{
		Id:          board.ID.String(),
		ProjectId:   board.ProjectID.String(),
		Name:        utils.GetString(board.Name),
		Description: utils.GetString(board.Description),
		UpdatedAt:   utils.GetTime(board.UpdatedAt),
		CreatedAt:   utils.GetTime(board.CreatedAt),
		Statuses:    statuses,
	}

	return c.JSON(http.StatusOK, boardResponse)
}

// GetBoardByProjectId godoc
// @Summary Получение досок по ID проекта
// @Description Получает список досок, связанных с указанным проектом
// @Tags Boards
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Security BearerAuth
// @Success 200 {object} response.BoardForProjectResponse "Список досок успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проекта"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении досок"
// @Router /boards/project/{projectId} [get]
func GetBoardByProjectId(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	projectUUID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
	}

	var boards []models.Board
	if err := DBConn.Session(&gorm.Session{}).
		Where("project_id = ? AND deleted = ?", projectUUID, false).
		Order("created_at DESC NULLS LAST").
		Preload("Statuses", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false).Order("sort_order ASC")
		}).
		Preload("Statuses.Tasks", func(db *gorm.DB) *gorm.DB {
			return db.Session(&gorm.Session{}).Where("deleted = ?", false)
		}).
		Find(&boards).Error; err != nil {
		log.Printf("DB error (find boards by project id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении досок"})
	}

	projectResponse := response.BoardForProjectResponse{
		ProjectId: projectUUID.String(),
		Boards:    make([]response.BoardResponse, 0, len(boards)),
	}

	for _, board := range boards {
		statuses := make([]response.StatusResponse, 0, len(board.Statuses))
		for _, status := range board.Statuses {
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

			statuses = append(statuses, response.StatusResponse{
				ID:        status.ID.String(),
				Key:       utils.GetString(status.Key),
				Name:      utils.GetString(status.Name),
				Color:     utils.GetString(status.Color),
				Order:     utils.GetInt(status.SortOrder), // ← ДОБАВЛЕНО
				IsDefault: utils.GetBool(status.IsDefault),
				IsActive:  utils.GetBool(status.IsActive),
				IsOpen:    utils.GetBool(status.IsOpen),
				CreatedAt: utils.GetTime(status.CreatedAt),
				UpdatedAt: utils.GetTime(status.UpdatedAt),
				Tasks:     tasks,
			})
		}

		projectResponse.Boards = append(projectResponse.Boards, response.BoardResponse{
			Id:          board.ID.String(),
			ProjectId:   board.ProjectID.String(),
			Name:        utils.GetString(board.Name),
			Description: utils.GetString(board.Description),
			CreatedAt:   utils.GetTime(board.CreatedAt),
			UpdatedAt:   utils.GetTime(board.UpdatedAt),
			Statuses:    statuses,
		})
	}

	return c.JSON(http.StatusOK, projectResponse)
}

// CreateBoard godoc
// @Summary Создание новой доски
// @Description Создает новую доску с указанными параметрами
// @Tags Boards
// @Accept json
// @Produce json
// @Param board body request.BoardCreateRequest true "Данные для создания доски"
// @Security BearerAuth
// @Success 201 {object} response.BoardUniversalResponse "Доска успешно создана"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректный идентификатор проекта"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании доски"
// @Router /boards [post]
func CreateBoard(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	var req request.BoardCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	if req.ProjectID == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Поле project_id обязательно"})
	}

	projectID, err := uuid.Parse(*req.ProjectID)
	if err != nil {
		log.Printf("UUID parse error (project_id): %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
	}

	now := time.Now()
	deleted := false
	boardID := uuid.New()
	tr := true

	board := models.Board{
		ID:          boardID,
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		Deleted:     &deleted,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	// Вспомогательная функция для создания статуса с уникальным Key и порядком
	makeStatus := func(order int, name, color string, isOpen bool) models.Status {
		id := uuid.New()
		key := id.String()[:8] // сокращённый UUID (8 символов)
		return models.Status{
			ID:        id,
			BoardID:   boardID,
			SortOrder:     &order,      // <-- НОВОЕ ПОЛЕ
			Key:       &key,
			Name:      &name,
			Color:     &color,
			IsDefault: &tr, // false
			IsActive:  boolPtr(true),
			IsOpen:    &isOpen,
			Deleted:   &deleted,
			CreatedAt: &now,
			UpdatedAt: &now,
		}
	}

	defaultStatuses := []models.Status{
		makeStatus(0, "To Do", "#fc0000ff", true),
		makeStatus(1024, "Done", "#28A745", true),
	}

	if err := DBConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{}).Create(&board).Error; err != nil {
			return err
		}
		if err := tx.Session(&gorm.Session{}).Create(&defaultStatuses).Error; err != nil {
			return err
		}
		return nil
	}); err != nil {
		log.Printf("DB transaction error (create board): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании доски"})
	}

	return c.JSON(http.StatusCreated, response.BoardUniversalResponse{
		Id:      board.ID.String(),
		Message: "Доска успешно создана",
	})
}

func boolPtr(b bool) *bool {
	return &b
}

// UpdateBoard godoc
// @Summary Обновление доски
// @Description Обновляет данные доски по её ID
// @Tags Boards
// @Accept json
// @Produce json
// @Param id path string true "ID доски"
// @Param board body request.BoardUpdateRequest true "Данные для обновления доски"
// @Security BearerAuth
// @Success 200 {object} response.BoardUniversalResponse "Доска успешно обновлена"
// @Failure 400 {object} map[string]string "Ошибка в запросе или нет полей для обновления"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Доска не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении доски"
// @Router /boards/{id} [patch]
func UpdateBoard(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор доски"})
	}

	var req request.BoardUpdateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	updateData := make(map[string]interface{})
	if req.Name != nil {
		updateData["name"] = *req.Name // ← разыменовываем указатель
	}
	if req.Description != nil {
		updateData["description"] = *req.Description // ← разыменовываем указатель
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
	}

	updateData["updated_at"] = time.Now()

	// Пытаемся обновить неудалённую доску
	result := DBConn.Session(&gorm.Session{}).Model(&models.Board{}).
		Where("id = ? AND deleted = ?", boardID, false).
		Updates(updateData)

	if result.Error != nil {
		log.Printf("DB error (update board): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении доски"})
	}

	if result.RowsAffected == 0 {
		// Проверяем, существует ли вообще доска с таким ID (даже удалённая)
		var count int64
		if err := DBConn.Session(&gorm.Session{}).Model(&models.Board{}).
			Where("id = ?", boardID).
			Count(&count).Error; err != nil {
			log.Printf("DB error (check existence): %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при проверке существования доски"})
		}

		if count == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Доска не найдена"})
		}
		// Если count > 0, но RowsAffected == 0 — значит, доска существует, но удалена (deleted = true)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Доска не найдена"})
	}

	return c.JSON(http.StatusOK, response.BoardUniversalResponse{
		Id:      boardID.String(),
		Message: "Доска успешно обновлена",
	})
}

// DeleteBoard godoc
// @Summary Удаление доски
// @Description Логическое удаление доски по ID (поле deleted = true), а также всех её статусов и задач
// @Tags Boards
// @Accept json
// @Produce json
// @Param id path string true "ID доски"
// @Security BearerAuth
// @Success 200 {object} response.BoardUniversalResponse "Доска успешно удалена"
// @Failure 404 {object} map[string]string "Доска не найдена"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении доски"
// @Router /boards/{id} [delete]
func DeleteBoard(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор доски"})
	}

	now := time.Now()
	deleted := true
	updateData := map[string]interface{}{
		"deleted":    &deleted,
		"updated_at": now,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		// 1. Проверяем, существует ли неудалённая доска
		var count int64
		if err := tx.Session(&gorm.Session{}).
			Model(&models.Board{}).
			Where("id = ? AND deleted = ?", boardID, false).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"error": "Доска не найдена или уже удалена"})
		}

		// 2. Удаляем доску
		if err := tx.Session(&gorm.Session{}).
			Model(&models.Board{}).
			Where("id = ?", boardID).
			Updates(updateData).Error; err != nil {
			return err
		}

		// 3. Получаем ID всех статусов этой доски (до их удаления!)
		var statusIDs []uuid.UUID
		if err := tx.Session(&gorm.Session{}).
			Model(&models.Status{}).
			Where("board_id = ? AND deleted = ?", boardID, false).
			Pluck("id", &statusIDs).Error; err != nil {
			return err
		}

		// 4. Удаляем задачи, привязанные к этим статусам
		if len(statusIDs) > 0 {
			if err := tx.Session(&gorm.Session{}).
				Model(&models.Task{}).
				Where("status_id IN ?", statusIDs).
				Updates(updateData).Error; err != nil {
				return err
			}
		}

		// 5. Удаляем статусы доски
		if err := tx.Session(&gorm.Session{}).
			Model(&models.Status{}).
			Where("board_id = ?", boardID).
			Updates(updateData).Error; err != nil {
			return err
		}

		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete board): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении доски"})
	}

	return c.JSON(http.StatusOK, response.BoardUniversalResponse{
		Id:      boardID.String(),
		Message: "Доска удалена",
	})
}