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
	// apply Keycloak auth middleware to all board routes
	projectGroup.Use(KeycloakAuthMiddleware)
	{
		projectGroup.GET("/all/:page/:pagesize", getAllBoards)
		projectGroup.GET("/:id", getBoardById)
		projectGroup.GET("/project/:projectId", getBoardByProjectId)
		projectGroup.POST("", createBoard)
		projectGroup.PATCH("/:id", updateBoard)
		projectGroup.DELETE("/:id", deleteBoard)
	}
}

// getAllBoards godoc
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
func getAllBoards(c echo.Context) error {
	if err := authorize(c); err != nil {
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
	if err := dbConn.Model(&models.Board{}).Session(&gorm.Session{}).Where("deleted = ?", false).Count(&total).Error; err != nil {
		log.Printf("DB error (count boards): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте досок"})
	}

	var boards []models.Board
	if err := dbConn.Session(&gorm.Session{}).Model(models.Board{}).Where("deleted = ?", false).Limit(pageSize).Offset(offset).Find(&boards).Error; err != nil {
		log.Printf("DB error (find boards): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении досок"})
	}

	boardIDs := make([]uuid.UUID, 0, len(boards))
	for _, b := range boards {
		boardIDs = append(boardIDs, b.ID)
	}

	var statusBoards []models.StatusBoard
	if len(boardIDs) > 0 {
    if err := dbConn.Model(&models.StatusBoard{}).Session(&gorm.Session{}).
        Preload("Status", "deleted = ?", false).
        Where("board_id IN ?", boardIDs). 
        Find(&statusBoards).Error; err != nil {
        log.Printf("DB error (find statusBoards): %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов"})
    	}
	}

	statusMap := make(map[uuid.UUID][]response.StatusResponse)
	for _, sb := range statusBoards {
		if sb.Status != nil {
			boardID := sb.BoardID 
			statusMap[boardID] = append(statusMap[boardID], response.StatusResponse{
				ID:        sb.Status.ID.String(),
				Key:       utils.GetString(sb.Status.Key),
				Name:      utils.GetString(sb.Status.Name),
				Color:     utils.GetString(sb.Status.Color),
				IsDefault: utils.GetBool(sb.Status.IsDefault),
				IsActive:  utils.GetBool(sb.Status.IsActive),
				IsOpen:    utils.GetBool(sb.Status.IsOpen),
				CreatedAt: utils.GetTime(sb.Status.CreatedAt),
				UpdatedAt: utils.GetTime(sb.Status.UpdatedAt),
			})
		}
	}

	boardList := response.BoardListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: int(total),
	}

	for _, board := range boards {
		boardList.Boards = append(boardList.Boards, response.BoardResponse{
			Id:          board.ID.String(),
			ProjectId:   board.ProjectID.String(),
			Name:        utils.GetString(board.Name),
			Description: utils.GetString(board.Description),
			UpdatedAt:   utils.GetTime(board.UpdatedAt),
			CreatedAt:   utils.GetTime(board.CreatedAt),
			Statuses:    statusMap[board.ID],
		})
	}

	return c.JSON(http.StatusOK, boardList)
}

// getBoardById godoc
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
func getBoardById(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	id := c.Param("id")
	boardId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	var board models.Board
	result := dbConn.Session(&gorm.Session{}).Model(models.Board{}).First(&board, "id = ? and deleted = ?", boardId, false)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Доска не найден",
			})
		}
		log.Printf("DB error (find project by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении доски из базы данных",
		})
	}

	if result.RowsAffected == 0{
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Доска не найдена",
		})
	}

	boardResponse := response.BoardResponse{
		Id:          board.ID.String(),
		ProjectId:   board.ProjectID.String(),
		Name:        utils.GetString(board.Name),
		Description: utils.GetString(board.Description),
		UpdatedAt:   utils.GetTime(board.UpdatedAt),
		CreatedAt: utils.GetTime(board.CreatedAt),
	}

	// Fetch all StatusBoard entries with preloaded Status in one query
    var statusBoards []models.StatusBoard
    if err := dbConn.Model(models.StatusBoard{}).
        Preload("Status", "statuses.deleted = ?", false).Session(&gorm.Session{}).
        Where("status_boards.deleted = ? AND status_boards.board_id = ?", false, board.ID).
        Find(&statusBoards).Error; err != nil {
        log.Printf("failed to get statuses for board %s: %v", boardId, err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Ошибка при получении статусов для доски",
        })
    }

    for _, sb := range statusBoards {
        if sb.Status != nil {
            boardResponse.Statuses = append(boardResponse.Statuses, response.StatusResponse{
				ID:        sb.Status.ID.String(),
				Key:       utils.GetString(sb.Status.Key),
				Name:      utils.GetString(sb.Status.Name),
				Color:     utils.GetString(sb.Status.Color),
				IsDefault: utils.GetBool(sb.Status.IsDefault),
				IsActive:  utils.GetBool(sb.Status.IsActive),
				IsOpen:    utils.GetBool(sb.Status.IsOpen),
				CreatedAt: utils.GetTime(sb.Status.CreatedAt),
				UpdatedAt: utils.GetTime(sb.Status.UpdatedAt),
            })
        }
    }

	return c.JSON(http.StatusOK, boardResponse)
}

// getBoardByProjectId godoc
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
// @Failure 404 {object} map[string]string "Доска не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении досок"
// @Router /boards/project/{projectId} [get]
func getBoardByProjectId(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	projectID := c.Param("projectId")
	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	// Загружаем доски проекта с предзагрузкой статусов
	var boards []models.Board
	if err := dbConn.Session(&gorm.Session{}).Model(models.Board{}).
		Preload("StatusBoards", "deleted = ?", false).
		Preload("StatusBoards.Status", "deleted = ?", false).
		Where("project_id = ? AND deleted = ?", projectUUID, false).
		Find(&boards).Error; err != nil {
		log.Printf("DB error (find boards by project id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении досок из базы данных",
		})
	}

	projectResponse := response.BoardForProjectResponse{
		ProjectId: projectID,
	}

	for _, board := range boards {
		boardResponse := response.BoardResponse{
			Id:          board.ID.String(),
			ProjectId:   board.ProjectID.String(),
			Name:        utils.GetString(board.Name),
			Description: utils.GetString(board.Description),
			CreatedAt:   utils.GetTime(board.CreatedAt),
			UpdatedAt:   utils.GetTime(board.UpdatedAt),
		}

		// Добавляем статусы из предзагрузки
		for _, sb := range board.StatusBoards {
			if sb.Status != nil {
				boardResponse.Statuses = append(boardResponse.Statuses, response.StatusResponse{
					ID:        sb.Status.ID.String(),
					Key:       utils.GetString(sb.Status.Key),
					Name:      utils.GetString(sb.Status.Name),
					Color:     utils.GetString(sb.Status.Color),
					IsDefault: utils.GetBool(sb.Status.IsDefault),
					IsActive:  utils.GetBool(sb.Status.IsActive),
					IsOpen:    utils.GetBool(sb.Status.IsOpen),
					CreatedAt: utils.GetTime(sb.Status.CreatedAt),
					UpdatedAt: utils.GetTime(sb.Status.UpdatedAt),
				})
			}
		}

		projectResponse.Boards = append(projectResponse.Boards, boardResponse)
	}

	return c.JSON(http.StatusOK, projectResponse)
}

// createBoard godoc
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
func createBoard(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	var req request.BoardCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	var projectId uuid.UUID
	if req.ProjectID != nil {
		projectID, err := uuid.Parse(*req.ProjectID)
		projectId = projectID
		if err != nil {
			log.Printf("UUID parse error: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Некорректный идентификатор проекта",
			})
		}
	}

	del := false

	now := time.Now()

	board := models.Board{
		ID:          newUUID,
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   projectId,
		Deleted: 		&del,
		CreatedAt: &now,
	}

	// create inside a transaction
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.Board{}).Create(&board)
		if res.Error != nil {
			log.Printf("DB error (create board): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create board): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании доски"})
	}

	createResponse := response.BoardUniversalResponse{
		Id:      newUUID.String(),
		Message: "Доска успешно создана",
	}
	return c.JSON(http.StatusCreated, createResponse)
}

// updateBoard godoc
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
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении доски"
// @Router /boards/{id} [patch]
func updateBoard(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	id := c.Param("id")
	boardId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор доски",
		})
	}

	var req request.BoardUpdateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.Name != nil {
    	updateData["name"] = req.Name
	}
	if req.Description != nil{
		updateData["description"] = req.Description
	}
	if req.Filter != nil{
		updateData["filter"] = req.Filter
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	// update inside a transaction
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.Board{}).Where("id = ? AND deleted = ?", boardId, false).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (update board): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update board): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении доски"})
	}

	updateReponse := response.BoardUniversalResponse{
		Id:      id,
		Message: "Доска успешно обновлена",
	}
	return c.JSON(http.StatusOK, updateReponse)
}

// deleteBoard godoc
// @Summary Удаление доски
// @Description Логическое удаление доски по ID (поле deleted = true)
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
func deleteBoard(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	updateData["updated_at"] = time.Now()
	// logical delete inside a transaction
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.Board{}).Where("id = ?", id).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete board): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		if res = tx.Session(&gorm.Session{}).Model(models.StatusBoard{}).Where("board_id = ?", id).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete status_board): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete board): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении доски"})
	}

	deleteResponse := response.BoardUniversalResponse{
		Id:      id,
		Message: "Доска с ID " + id + " удален",
	}

	return c.JSON(http.StatusOK, deleteResponse)
}
