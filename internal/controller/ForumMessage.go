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
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func RegisterForumMessagesRoutes(e *echo.Echo) {
	forumMessageGroup := e.Group("/forum-messages")
	forumMessageGroup.Use(KeycloakAuthMiddleware)
	{
		forumMessageGroup.GET("/all/:page/:pagesize", getAllForumMessages)
		forumMessageGroup.GET("/problem/:id/:page/:pagesize", getForumMessagesByProblemId)
		forumMessageGroup.GET("/:id", getForumMessageById)
		forumMessageGroup.POST("", createForumMessage)
		forumMessageGroup.PATCH("/:id", updateForumMessage)
		forumMessageGroup.DELETE("/:id", deleteForumMessage)
	}
}

// getAllForumMessages godoc
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
func getAllForumMessages(c echo.Context) error {
	if err := authorize(c); err != nil {
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

	var totalCount int64
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.ForumMessage{}).
		Where("deleted = FALSE").
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count forum messages): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете сообщений форума"})
	}

	var forumMessages []models.ForumMessage
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.ForumMessage{}).
		Where("deleted = FALSE").
		Limit(pageSize).
		Offset(offset).
		Find(&forumMessages).Error; err != nil {
		log.Printf("DB error (find forum messages): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении сообщений форума из базы данных"})
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

// getForumMessagesByProblemId godoc
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
func getForumMessagesByProblemId(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

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
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.ForumMessage{}).
		Where("deleted = FALSE AND problem_id = ?", problemID).
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count forum messages): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете сообщений форума"})
	}

	var forumMessages []models.ForumMessage
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.ForumMessage{}).
		Where("deleted = FALSE AND problem_id = ?", problemID).
		Limit(pageSize).
		Offset(offset).
		Find(&forumMessages).Error; err != nil {
		log.Printf("DB error (find forum messages): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении сообщений форума из базы данных"})
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

// getForumMessageById godoc
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
func getForumMessageById(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат идентификатора сообщения"})
	}

	var m models.ForumMessage
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.ForumMessage{}).
		Where("id = ? AND deleted = FALSE", messageID).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Сообщение не найдено"})
		}
		log.Printf("DB error (find forum message): %v", err)
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

// createForumMessage godoc
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
func createForumMessage(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	var req request.CreateForumMessageRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	problemId, err := uuid.Parse(req.ProblemID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проблемы"})
	}

	var creatorId *uuid.UUID
	if req.CreatorID != "" {
		v, err := uuid.Parse(req.CreatorID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор создателя"})
		}
		creatorId = &v
	}

	now := time.Now()
	del := false

	var description pq.StringArray
	if req.Description != nil {
		description = pq.StringArray(*req.Description)
	} else {
		description = pq.StringArray{}
	}

	fm := models.ForumMessage{
		ID:          uuid.New(),
		ProblemID:   problemId,
		Description: description,
		CreatorID:   creatorId,
		CreatedAt:   &now,
		Deleted:     &del,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).
			Model(&models.ForumMessage{}).
			Create(&fm); res.Error != nil {
			log.Printf("DB error (create forum message): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create forum message): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании сообщения форума"})
	}

	return c.JSON(http.StatusCreated, response.ForumMessageUniversalResponse{
		ID:      fm.ID.String(),
		Message: "Сообщение форума создано",
	})
}

// updateForumMessage godoc
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
func updateForumMessage(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат идентификатора сообщения"})
	}

	var req request.UpdateForumMessageRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	updateData := map[string]interface{}{}
	if req.ProblemID != nil {
		pid, err := uuid.Parse(*req.ProblemID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проблемы"})
		}
		updateData["problem_id"] = pid
	}
	if req.Description != nil {
		updateData["description"] = pq.StringArray(*req.Description)
	}
	if req.CreatorID != nil {
		cid, err := uuid.Parse(*req.CreatorID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор создателя"})
		}
		updateData["creator_id"] = cid
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
	}

	now := time.Now()
	updateData["updated_at"] = &now

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.ForumMessage{}).
			Where("id = ? AND deleted = FALSE", messageID).
			Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (update forum message): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update forum message): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении сообщения форума"})
	}

	return c.JSON(http.StatusOK, response.ForumMessageUniversalResponse{
		ID:      messageID.String(),
		Message: "Сообщение форума обновлено",
	})
}

// deleteForumMessage godoc
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
func deleteForumMessage(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	messageID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат идентификатора сообщения"})
	}

	delTrue := true
	now := time.Now()
	update := map[string]interface{}{
		"deleted":    &delTrue,
		"updated_at": &now,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.ForumMessage{}).
			Where("id = ?", messageID).
			Updates(update)
		if res.Error != nil {
			log.Printf("DB error (delete forum message): %v", res.Error)
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
		log.Printf("DB transaction error (delete forum message): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении сообщения форума"})
	}

	return c.JSON(http.StatusOK, response.ForumMessageUniversalResponse{
		ID:      messageID.String(),
		Message: "Сообщение форума удалено",
	})
}
