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
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Board{}).
		Where("deleted = FALSE").
		Count(&total).Error; err != nil {
		log.Printf("DB error (count boards): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте досок"})
	}

	var boards []models.Board
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Board{}).
		Where("deleted = FALSE").
		Order("created_at DESC NULLS LAST").
		Limit(pageSize).Offset(offset).
		Find(&boards).Error; err != nil {
		log.Printf("DB error (find boards): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении досок"})
	}

	boardIDs := make([]uuid.UUID, 0, len(boards))
	for _, b := range boards {
		boardIDs = append(boardIDs, b.ID)
	}

	var statusBoards []models.StatusBoard
	if len(boardIDs) > 0 {
		if err := DBConn.Session(&gorm.Session{}).
			Model(&models.StatusBoard{}).
			Preload("Status", "deleted = FALSE").
			Where("deleted = FALSE AND board_id IN ?", boardIDs).
			Find(&statusBoards).Error; err != nil {
			log.Printf("DB error (find statusBoards): %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов"})
		}
	}

	statusMap := make(map[uuid.UUID][]response.StatusResponse)
	for _, sb := range statusBoards {
		if sb.Status != nil {
			statusMap[sb.BoardID] = append(statusMap[sb.BoardID], response.StatusResponse{
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
	res := DBConn.Session(&gorm.Session{}).
		Model(&models.Board{}).
		First(&board, "id = ? AND deleted = FALSE", boardID)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Доска не найдена"})
		}
		log.Printf("DB error (find board by id): %v", res.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении доски"})
	}
	log.Printf("Loaded board: Name=%v, Desc=%v", board.Name, board.Description)

	boardResponse := response.BoardResponse{
		Id:          board.ID.String(),
		ProjectId:   board.ProjectID.String(),
		Name:        utils.GetString(board.Name),
		Description: utils.GetString(board.Description),
		UpdatedAt:   utils.GetTime(board.UpdatedAt),
		CreatedAt:   utils.GetTime(board.CreatedAt),
	}

	var statusBoards []models.StatusBoard
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.StatusBoard{}).
		Preload("Status", "deleted = FALSE").
		Where("deleted = FALSE AND board_id = ?", board.ID).
		Find(&statusBoards).Error; err != nil {
		log.Printf("DB error (statuses for board %s): %v", boardID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов для доски"})
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
// @Failure 404 {object} map[string]string "Доска не найдена"
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
		Model(&models.Board{}).
		Preload("StatusBoards", "deleted = FALSE").
		Preload("StatusBoards.Status", "deleted = FALSE").
		Where("project_id = ? AND deleted = FALSE", projectUUID).
		Order("created_at DESC NULLS LAST").
		Find(&boards).Error; err != nil {
		log.Printf("DB error (find boards by project id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении досок"})
	}

	projectResponse := response.BoardForProjectResponse{ProjectId: projectUUID.String()}
	for _, board := range boards {
		br := response.BoardResponse{
			Id:          board.ID.String(),
			ProjectId:   board.ProjectID.String(),
			Name:        utils.GetString(board.Name),
			Description: utils.GetString(board.Description),
			CreatedAt:   utils.GetTime(board.CreatedAt),
			UpdatedAt:   utils.GetTime(board.UpdatedAt),
		}
		for _, sb := range board.StatusBoards {
			if sb.Status != nil {
				br.Statuses = append(br.Statuses, response.StatusResponse{
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
		projectResponse.Boards = append(projectResponse.Boards, br)
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

	var projectID uuid.UUID
	if req.ProjectID != nil {
		parsed, err := uuid.Parse(*req.ProjectID)
		if err != nil {
			log.Printf("UUID parse error (project_id): %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
		}
		projectID = parsed
	}

	now := time.Now()
	del := false
	boardId := uuid.New()

	board := models.Board{
		ID:          boardId,
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		Deleted:     &del,
		CreatedAt:   &now,
	}

	statusBoards := []models.StatusBoard{
		{StatusID: BaseStartStatus,
		BoardID: boardId,
		CreatedAt: &now,
		UpdatedAt: &now,
		Deleted: &del,},
		{StatusID: BaseEndStatus,
		BoardID: boardId,
		CreatedAt: &now,
		UpdatedAt: &now,
		Deleted: &del,},
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{}).
			Model(&models.Board{}).
			Create(&board).Error; err != nil{
				return err
			}
		if err := tx.Session(&gorm.Session{}).
			Model(&models.StatusBoard{}).
			Create(&statusBoards).Error; err != nil{
				return err
			}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create board): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании доски"})
	}

	return c.JSON(http.StatusCreated, response.BoardUniversalResponse{
		Id:      board.ID.String(),
		Message: "Доска успешно создана",
	})
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

	updateData := make(map[string]any)
	if req.Name != nil {
		updateData["name"] = req.Name
	}
	if req.Description != nil {
		updateData["description"] = req.Description
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
	}

	now := time.Now()
	updateData["updated_at"] = &now

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.Board{}).
			Where("id = ? AND deleted = FALSE", boardID).
			Updates(updateData)
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

	return c.JSON(http.StatusOK, response.BoardUniversalResponse{
		Id:      boardID.String(),
		Message: "Доска успешно обновлена",
	})
}

// DeleteBoard godoc
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
	del := true
	updateData := map[string]any{
		"deleted":    &del,
		"updated_at": &now,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		// помечаем саму доску
		res := tx.Session(&gorm.Session{}).
			Model(&models.Board{}).
			Where("id = ? AND deleted = FALSE", boardID).
			Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete board): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Доска не найдена или уже удалена"})
		}
		// помечаем связи статусов
		if res = tx.Session(&gorm.Session{}).
			Model(&models.StatusBoard{}).
			Where("board_id = ? AND deleted = FALSE", boardID).
			Updates(updateData); res.Error != nil {
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

	return c.JSON(http.StatusOK, response.BoardUniversalResponse{
		Id:      boardID.String(),
		Message: "Доска удалена",
	})
}
