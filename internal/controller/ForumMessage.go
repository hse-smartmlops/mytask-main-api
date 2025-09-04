package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
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

func RegisterForumMessagesRoutes(e *echo.Echo){
	forumMessageGroup := e.Group("/forum-messages")
	forumMessageGroup.Use(KeycloakAuthMiddleware)
	{
		forumMessageGroup.GET("/all/:page/:pagesize", getAllForumMessages)
		forumMessageGroup.GET("/problem/:id", getForumMessagesByProblemId)
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

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.ForumMessage{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count forum messages): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете сообщений форума",
		})
	}

	// Получаем список проектов с пагинацией
	var forumMessages []models.ForumMessage
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&forumMessages).Error; err != nil {
		log.Printf("DB error (find forum messages): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении сообщений форума из базы данных",
		})
	}

	forumMessageList := response.ForumMessageListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, message := range forumMessages {
		var messageId = message.ID.String()
		var problemId = message.ProblemID.String()
		var description []string = []string(message.Description)
		var creatorId string
		if message.CreatorID != nil {
			creatorId = message.CreatorID.String()
		}
		var createdAt time.Time
		if message.CreatedAt != nil {
			createdAt = *message.CreatedAt
		}
		var updatedAt time.Time
		if message.UpdatedAt != nil {
			updatedAt = *message.UpdatedAt
		}
		forumMessageList.Messages = append(forumMessageList.Messages, response.ForumMessageResponse{
			ID:          messageId,
			ProblemID:   problemId,
			Description: description,
			CreatorID:   creatorId,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}
	return c.JSON(http.StatusOK, forumMessageList)
}

// getForumMessagesByProblemId godoc
// @Summary Получение сообщений форума по ID проблемы
// @Description Получает список сообщений форума, связанных с указанной проблемой, с учетом пагинации
// @Tags ForumMessages
// @Accept json
// @Produce json
// @Param id path string true "ID проблемы"
// @Param page query int false "Номер страницы" default(1)
// @Param pageSize query int false "Размер страницы" default(10)
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ForumMessageListByProblemIdResponse "Список сообщений форума успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проблемы или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении сообщений форума"
// @Router /forum-messages/problem/{id} [get]
func getForumMessagesByProblemId(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	problemID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid UUID format: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Неверный формат идентификатора проблемы",
		})
	}

	var req request.ForumMessageListByProblemIdRequest
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

	var totalCount int64
	result := dbConn.Session(&gorm.Session{}).Model(models.ForumMessage{}).Where("deleted = ? and problem_id = ?", false, problemID).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count forum messages): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете сообщений форума",
		})
	}

	// Получаем список проектов с пагинацией
	var forumMessages []models.ForumMessage
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ? and problem_id = ?", false, problemID).
		Limit(pageSize).
		Offset(offset).
		Find(&forumMessages).Error; err != nil {
		log.Printf("DB error (find forum messages): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении сообщений форума из базы данных",
		})
	}

	forumMessageList := response.ForumMessageListByProblemIdResponse{
		ProblemId:  id,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, message := range forumMessages {
		var messageId = message.ID.String()
		var problemId = message.ProblemID.String()
		var description []string = []string(message.Description)

		var creatorId string
		if message.CreatorID != nil {
			creatorId = message.CreatorID.String()
		}
		var createdAt time.Time
		if message.CreatedAt != nil {
			createdAt = *message.CreatedAt
		}
		var updatedAt time.Time
		if message.UpdatedAt != nil {
			updatedAt = *message.UpdatedAt
		}
		forumMessageList.Messages = append(forumMessageList.Messages, response.ForumMessageResponse{
			ID:          messageId,
			ProblemID:   problemId,
			Description: description,
			CreatorID:   creatorId,
			CreatedAt:   createdAt,
			UpdatedAt:   updatedAt,
		})
	}
	return c.JSON(http.StatusOK, forumMessageList)
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
	id := c.Param("id")
	messageID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid UUID format: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Неверный формат идентификатора сообщения",
		})
	}

	var forumMessage models.ForumMessage
	if err := dbConn.Session(&gorm.Session{}).Where("id = ? and deleted = ?", messageID, false).First(&forumMessage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Сообщение не найдено",
			})
		}
		log.Printf("DB error (find forum message): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении сообщения форума из базы данных",
		})
	}
	
	var problemId = forumMessage.ProblemID.String()
	var description []string = []string(forumMessage.Description)
	var creatorId string
	if forumMessage.CreatorID != nil{
		creatorId = forumMessage.CreatorID.String()
	}
	var createdAt time.Time
	if forumMessage.CreatedAt != nil{
		createdAt = *forumMessage.CreatedAt
	}
	var updatedAt time.Time
	if forumMessage.UpdatedAt != nil{
		updatedAt = *forumMessage.UpdatedAt
	}

	forumMessageResponse := response.ForumMessageResponse{
		ID: forumMessage.ID.String(),
		ProblemID: problemId,
		Description: description,
		CreatorID: creatorId,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return c.JSON(http.StatusOK, forumMessageResponse)
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
func createForumMessage(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	var req request.CreateForumMessageRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	now := time.Now()

	problemId, err := uuid.Parse(req.ProblemID)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проблемы",
		})
	}

	creatorId, err := uuid.Parse(req.CreatorID)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор создателя",
		})
	}

	del := false

	var description pq.StringArray
	if req.Description != nil{
		description = pq.StringArray(*req.Description)
	}else{
		description = pq.StringArray{}
	}

	forumMessage := models.ForumMessage{
		ID: newUUID,
		ProblemID: problemId,
		Description: description,
		CreatorID: &creatorId,
		CreatedAt: &now,	
		Deleted: &del,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Create(&forumMessage); res.Error != nil {
			log.Printf("DB error (create forum message): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create forum message): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании сообщения форума"})
	}

	createResponse := response.ForumMessageUniversalResponse{
		ID: newUUID.String(),
		Message: "Сообщение форума создано",
	}

	return c.JSON(http.StatusCreated, createResponse)
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
func updateForumMessage(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	messageID, err := uuid.Parse(id)
	if err != nil {
		log.Printf("Invalid UUID format: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Неверный формат идентификатора сообщения",
		})
	}

	var req request.UpdateForumMessageRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.ProblemID != nil {
		problemId, err := uuid.Parse(*req.ProblemID)
		if err != nil {
			log.Printf("UUID parse error: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Некорректный идентификатор проблемы",
			})
		}
		updateData["problem_id"] = problemId
	}	

	if req.Description != nil{
		updateData["description"] = req.Description
	}

	if req.CreatorID != nil{
		creatorId, err := uuid.Parse(*req.CreatorID)
		if err != nil{
			log.Printf("UUID parse error: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Некорректный идентификатор создателя",
			})
		}
		updateData["creator_id"] = creatorId
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(models.ForumMessage{}).Where("id = ?", messageID).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (update forum message): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0{
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update forum message): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении сообщения форума"})
	}

	updateResponse := response.ForumMessageUniversalResponse{
		ID: id,
		Message: "Сообщение форума обновлено",
	}

	return c.JSON(http.StatusOK, updateResponse)
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
func deleteForumMessage(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(models.ForumMessage{}).Where("id = ?", id).Updates(updateData)
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

	delResponse := response.ForumMessageUniversalResponse{
		ID: id,
		Message: "Сообщение форума удалено",
	}

	return c.JSON(http.StatusOK, delResponse)
}

