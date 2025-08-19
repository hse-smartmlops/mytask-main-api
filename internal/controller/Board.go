package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

func RegisterBoardRoutes(e *echo.Echo) {
	projectGroup := e.Group("/boards")
	{
		projectGroup.GET("", GetAllBoards)
		projectGroup.GET("/:id", GetBoardById)
		projectGroup.GET("/project/:projectId", GetBoardByProjectId)
		projectGroup.POST("", CreateBoard)
		projectGroup.PATCH("/:id", UpdateBoard)
		projectGroup.DELETE("/:id", DeleteBoard)
	}
}

func GetAllBoards(c echo.Context) error {
	var req request.BoardListRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	// Значения по умолчанию
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := dbConn.Session(&gorm.Session{}).Model(&models.Board{}).Count(&total).Error; err != nil {
		log.Printf("DB error (count boards): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчёте досок",
		})
	}

	var boards []models.Board
	if err := dbConn.Session(&gorm.Session{}).
		Limit(pageSize).
		Offset(offset).
		Find(&boards).Error; err != nil {
		log.Printf("DB error (find boards): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении досок из базы данных",
		})
	}

	boardList := response.BoardListResponse{
		Page:     page,
		PageSize: pageSize,
	}

	for _, board := range boards {
		if board.Deleted != nil {
			if !*board.Deleted {
				var id string
				if board.ID != nil {
					id = board.ID.String()
				}
				var projectId string
				if board.ProjectID != nil {
					projectId = board.ProjectID.String()
				}
				var name string
				if board.Name != nil {
					name = *board.Name
				}
				var description string
				if board.Description != nil {
					description = *board.Description
				}
				var filter string
				if board.Filter != nil {
					filter = *board.Filter
				}
				var updatedAt time.Time
				if board.UpdatedAt != nil {
					updatedAt = *board.UpdatedAt
				}

				boardList.Boards = append(boardList.Boards, response.BoardResponse{
					Id:          id,
					ProjectId:   projectId,
					Name:        name,
					Description: description,
					Filter:      filter,
					UpdatedAt:   updatedAt,
				})
			}
		}
	}
	boardList.TotalCount = len(boardList.Boards)
	return c.JSON(http.StatusOK, boardList)
}

func GetBoardById(c echo.Context) error {
	id := c.Param("id")
	boardId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	var board models.Board
	result := dbConn.Session(&gorm.Session{}).First(&board, "id = ?", boardId)
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

	if board.Deleted != nil {
		if *board.Deleted {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Доска не найдена",
			})
		}
	}

	var projectId string
	if board.ProjectID != nil {
		projectId = board.ProjectID.String()
	}

	var name string
	if board.Name != nil {
		name = *board.Name
	}

	var description string
	if board.Description != nil {
		description = *board.Description
	}

	var filter string
	if board.Filter != nil {
		filter = *board.Filter
	}

	var updatedAt time.Time
	if board.UpdatedAt != nil {
		updatedAt = *board.UpdatedAt
	}

	boardResponse := response.BoardResponse{
		Id:          id,
		ProjectId:   projectId,
		Name:        name,
		Description: description,
		Filter:      filter,
		UpdatedAt:   updatedAt,
	}
	return c.JSON(http.StatusOK, boardResponse)
}

func GetBoardByProjectId(c echo.Context) error {
	projectID := c.Param("projectId")
	projectId, err := uuid.Parse(projectID)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	var boards []models.Board
	result := dbConn.Session(&gorm.Session{}).Find(&boards, "project_id = ?", projectId)
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

	projectResponse := response.BoardForProjectResponse{
		ProjectId: projectID,
	}

	for _, board := range boards {
		if board.Deleted != nil {
			if *board.Deleted {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Доска не найдена",
				})
			}
		}

		var name string
		if board.Name != nil {
			name = *board.Name
		}

		var description string
		if board.Description != nil {
			description = *board.Description
		}

		var filter string
		if board.Filter != nil {
			filter = *board.Filter
		}

		var updatedAt time.Time
		if board.UpdatedAt != nil {
			updatedAt = *board.UpdatedAt
		}

		projectResponse.Boards = append(projectResponse.Boards, response.BoardResponse{
			Id:          board.ID.String(),
			ProjectId:   projectId.String(),
			Name:        name,
			Description: description,
			Filter:      filter,
			UpdatedAt:   updatedAt,
		})
	}

	return c.JSON(http.StatusOK, projectResponse)
}

func CreateBoard(c echo.Context) error {
	var req request.BoardCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	var projectId *uuid.UUID
	if req.ProjectID != nil {
		projectID, err := uuid.Parse(*req.ProjectID)
		projectId = &projectID
		if err != nil {
			log.Printf("UUID parse error: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Некорректный идентификатор проекта",
			})
		}
	}

	board := models.Board{
		ID:          &newUUID,
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   projectId,
		Filter:      req.Filter,
	}

	result := dbConn.Session(&gorm.Session{}).Create(&board)
	if result.Error != nil {
		log.Printf("DB error (create board): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании доски",
		})
	}

	createResponse := response.BoardUniversalResponse{
		Id:      newUUID.String(),
		Message: "Доска успешно создана",
	}
	return c.JSON(http.StatusCreated, createResponse)
}

func UpdateBoard(c echo.Context) error {
	id := c.Param("id")
	var req request.BoardUpdateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	updateData["name"] = req.Name
	updateData["description"] = req.Description
	updateData["filter"] = req.Filter

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	if err := dbConn.Session(&gorm.Session{}).Model(models.Board{}).Where("id = ?", id).Updates(updateData).Error; err != nil {
		log.Printf("DB error (update board): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении доски",
		})
	}

	updateReponse := response.BoardUniversalResponse{
		Id:      id,
		Message: "Доска успешно обновлена",
	}
	return c.JSON(http.StatusOK, updateReponse)
}

func DeleteBoard(c echo.Context) error {
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	updateData["updated_at"] = time.Now()
	result := dbConn.Session(&gorm.Session{}).Model(models.Board{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete board): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении доски",
		})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не удалено",
		})
	}

	deleteResponse := response.BoardUniversalResponse{
		Id:      id,
		Message: "Доска с ID " + id + " удален",
	}

	return c.JSON(http.StatusOK, deleteResponse)
}
