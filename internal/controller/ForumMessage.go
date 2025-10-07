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

type ForumMessageController struct {
	forumMessageService service.ForumMessageService
}

func NewForumMessageController(forumMessageService service.ForumMessageService) *ForumMessageController {
	return &ForumMessageController{
		forumMessageService: forumMessageService,
	}
}

func RegisterForumMessagesRoutes(e *echo.Echo, forumMessageService service.ForumMessageService) {
	controller := NewForumMessageController(forumMessageService)
	forumMessageGroup := e.Group("/forum-messages")
	{
		forumMessageGroup.GET("/all/:page/:pagesize", controller.GetAllForumMessages)
		forumMessageGroup.GET("/problem/:id/:page/:pagesize", controller.GetForumMessagesByProblemId)
		forumMessageGroup.GET("/:id", controller.GetForumMessageById)
		forumMessageGroup.POST("", controller.CreateForumMessage)
		forumMessageGroup.PATCH("/:id", controller.UpdateForumMessage)
		forumMessageGroup.DELETE("/:id", controller.DeleteForumMessage)
	}
}

// GetAllForumMessages godoc
// @Summary Получение списка всех сообщений форума
// @Description Получает список всех сообщений форума с учетом пагинации, исключая удаленные
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ForumMessageListResponse "Список сообщений форума успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении сообщений форума"
// @Router /forum-messages/all/{page}/{pagesize} [get]
func (fmc *ForumMessageController) GetAllForumMessages(c echo.Context) error {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге размера страницы"})
	}

	forumMessages, totalCount, err := fmc.forumMessageService.GetAllForumMessages(page, pageSize)
	if err != nil {
		log.Printf("service error (get all forum messages): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете сообщений форума"})
	}

	out := response.ForumMessageListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}
	for _, m := range forumMessages {
		description := []string(m.Description)
		var creatorId string
		if m.CreatorID != nil {
			creatorId = m.CreatorID.String()
		}
		out.Messages = append(out.Messages, response.ForumMessageResponse{
			ID:          m.ID.String(),
			ProblemID:   m.ProblemID.String(),
			Description: description,
			CreatorID:   creatorId,
			CreatedAt:   utils.GetTime(m.CreatedAt),
			UpdatedAt:   utils.GetTime(m.UpdatedAt),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// GetForumMessagesByProblemId godoc
// @Summary Получение сообщений форума по ID проблемы
// @Description Получает список сообщений форума, связанных с указанной проблемой, с учетом пагинации
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param id path string true "ID проблемы"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ForumMessageListByProblemIdResponse "Список сообщений форума успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проблемы или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении сообщений форума"
// @Router /forum-messages/problem/{id}/{page}/{pagesize} [get]
func (fmc *ForumMessageController) GetForumMessagesByProblemId(c echo.Context) error {
	problemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат идентификатора проблемы"})
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге размера страницы"})
	}

	forumMessages, totalCount, err := fmc.forumMessageService.GetForumMessagesByProblemId(problemID, page, pageSize)
	if err != nil {
		log.Printf("service error (get forum messages by problem id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете сообщений форума"})
	}

	out := response.ForumMessageListByProblemIdResponse{
		ProblemId:  problemID.String(),
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}
	for _, m := range forumMessages {
		description := []string(m.Description)
		var creatorId string
		if m.CreatorID != nil {
			creatorId = m.CreatorID.String()
		}
		out.Messages = append(out.Messages, response.ForumMessageResponse{
			ID:          m.ID.String(),
			ProblemID:   m.ProblemID.String(),
			Description: description,
			CreatorID:   creatorId,
			CreatedAt:   utils.GetTime(m.CreatedAt),
			UpdatedAt:   utils.GetTime(m.UpdatedAt),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// GetForumMessageById godoc
// @Summary Получение сообщения форума по ID
// @Description Получает данные сообщения форума по его уникальному идентификатору
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param id path string true "ID сообщения форума"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ForumMessageResponse "Сообщение форума успешно получено"
// @Failure 400 {object} map[string]string "Некорректный идентификатор сообщения"
// @Failure 404 {object} map[string]string "Сообщение форума не найдено"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении сообщения форума"
// @Router /forum-messages/{id} [get]
func (fmc *ForumMessageController) GetForumMessageById(c echo.Context) error {
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат идентификатора сообщения"})
	}

	m, err := fmc.forumMessageService.GetForumMessageById(messageID)
	if err != nil {
		if err.Error() == "forum message not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Сообщение не найдено"})
		}
		log.Printf("service error (get forum message): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении сообщения форума из базы данных"})
	}

	description := []string(m.Description)
	var creatorId string
	if m.CreatorID != nil {
		creatorId = m.CreatorID.String()
	}

	return c.JSON(http.StatusOK, response.ForumMessageResponse{
		ID:          m.ID.String(),
		ProblemID:   m.ProblemID.String(),
		Description: description,
		CreatorID:   creatorId,
		CreatedAt:   utils.GetTime(m.CreatedAt),
		UpdatedAt:   utils.GetTime(m.UpdatedAt),
	})
}

// CreateForumMessage godoc
// @Summary Создание нового сообщения форума
// @Description Создает новое сообщение форума с указанными параметрами
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param forumMessage body request.CreateForumMessageRequest true "Данные для создания сообщения форума"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.ForumMessageUniversalResponse "Сообщение форума успешно создано"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании сообщения форума"
// @Router /forum-messages [post]
func (fmc *ForumMessageController) CreateForumMessage(c echo.Context) error {
	var req request.CreateForumMessageRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	messageID, err := fmc.forumMessageService.CreateForumMessage(req)
	if err != nil {
		if err.Error() == "invalid problem id" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проблемы"})
		}
		if err.Error() == "invalid creator id" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор создателя"})
		}
		log.Printf("service error (create forum message): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании сообщения форума"})
	}

	return c.JSON(http.StatusCreated, response.ForumMessageUniversalResponse{
		ID:      messageID.String(),
		Message: "Сообщение форума создано",
	})
}

// UpdateForumMessage godoc
// @Summary Обновление сообщения форума
// @Description Обновляет данные сообщения форума по его ID
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param id path string true "ID сообщения форума"
// @Param forumMessage body request.UpdateForumMessageRequest true "Данные для обновления сообщения форума"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ForumMessageUniversalResponse "Сообщение форума успешно обновлено"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении сообщения форума"
// @Router /forum-messages/{id} [patch]
func (fmc *ForumMessageController) UpdateForumMessage(c echo.Context) error {
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат идентификатора сообщения"})
	}

	var req request.UpdateForumMessageRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	err = fmc.forumMessageService.UpdateForumMessage(messageID, req)
	if err != nil {
		if err.Error() == "no fields to update" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
		}
		if err.Error() == "forum message not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		if err.Error() == "invalid problem id" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проблемы"})
		}
		if err.Error() == "invalid creator id" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор создателя"})
		}
		log.Printf("service error (update forum message): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении сообщения форума"})
	}

	return c.JSON(http.StatusOK, response.ForumMessageUniversalResponse{
		ID:      messageID.String(),
		Message: "Сообщение форума обновлено",
	})
}

// DeleteForumMessage godoc
// @Summary Удаление сообщения форума
// @Description Логическое удаление сообщения форума по ID (поле deleted = true)
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param id path string true "ID сообщения форума"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ForumMessageUniversalResponse "Сообщение форума успешно удалено"
// @Failure 404 {object} map[string]string "Сообщение форума не найдено"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении сообщения форума"
// @Router /forum-messages/{id} [delete]
func (fmc *ForumMessageController) DeleteForumMessage(c echo.Context) error {
	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат идентификатора сообщения"})
	}

	err = fmc.forumMessageService.DeleteForumMessage(messageID)
	if err != nil {
		if err.Error() == "forum message not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		log.Printf("service error (delete forum message): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении сообщения форума"})
	}

	return c.JSON(http.StatusOK, response.ForumMessageUniversalResponse{
		ID:      messageID.String(),
		Message: "Сообщение форума удалено",
	})
}