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

type BoardController struct {
	boardService service.BoardService
}

func NewBoardController(boardService service.BoardService) *BoardController {
	return &BoardController{
		boardService: boardService,
	}
}

func RegisterBoardRoutes(e *echo.Echo, boardService service.BoardService) {
	controller := NewBoardController(boardService)
	projectGroup := e.Group("/boards")
	{
		projectGroup.GET("/all/:page/:pagesize", controller.GetAllBoards)
		projectGroup.GET("/:id", controller.GetBoardById)
		projectGroup.GET("/project/:projectId", controller.GetBoardByProjectId)
		projectGroup.POST("", controller.CreateBoard)
		projectGroup.PATCH("/:id", controller.UpdateBoard)
		projectGroup.DELETE("/:id", controller.DeleteBoard)
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
func (bc *BoardController) GetAllBoards(c echo.Context) error {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	boards, total, err := bc.boardService.GetAllBoards(page, pageSize)
	if err != nil {
		log.Printf("service error (get all boards): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении досок"})
	}

	// Формируем ответ
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
				Tasks:     tasks,
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
func (bc *BoardController) GetBoardById(c echo.Context) error {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор доски"})
	}

	board, err := bc.boardService.GetBoardById(boardID)
	if err != nil {
		if err.Error() == "board not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Доска не найдена"})
		}
		log.Printf("service error (get board by id): %v", err)
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
			Order:     utils.GetInt(status.SortOrder),
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
func (bc *BoardController) GetBoardByProjectId(c echo.Context) error {
	projectUUID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
	}

	boards, err := bc.boardService.GetBoardByProjectId(projectUUID)
	if err != nil {
		log.Printf("service error (get board by project id): %v", err)
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
				Order:     utils.GetInt(status.SortOrder),
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
func (bc *BoardController) CreateBoard(c echo.Context) error {
	var req request.BoardCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	boardID, err := bc.boardService.CreateBoard(req)
	if err != nil {
		if err.Error() == "invalid project id" {
			log.Printf("UUID parse error (project_id): %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
		}
		log.Printf("service error (create board): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании доски"})
	}

	return c.JSON(http.StatusCreated, response.BoardUniversalResponse{
		Id:      boardID.String(),
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
// @Failure 404 {object} map[string]string "Доска не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении доски"
// @Router /boards/{id} [patch]
func (bc *BoardController) UpdateBoard(c echo.Context) error {
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

	err = bc.boardService.UpdateBoard(boardID, req)
	if err != nil {
		if err.Error() == "no fields to update" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
		}
		if err.Error() == "board not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Доска не найдена"})
		}
		log.Printf("service error (update board): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении доски"})
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
func (bc *BoardController) DeleteBoard(c echo.Context) error {
	boardID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор доски"})
	}

	err = bc.boardService.DeleteBoard(boardID)
	if err != nil {
		if err.Error() == "board not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Доска не найдена"})
		}
		log.Printf("service error (delete board): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении доски"})
	}

	return c.JSON(http.StatusOK, response.BoardUniversalResponse{
		Id:      boardID.String(),
		Message: "Доска удалена",
	})
}