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
		g.GET("/project/:board_id", GetStatusesByBoardId)
		g.GET("/task/:task_id", GetStatusesByTaskId)
		g.POST("/add-to-task", AddStatusToTask)
		g.POST("/add-to-board", AddStatusToBoard)
		g.DELETE("/delete-from-task/:task_id/:status_id", DeleteStatusFromTask)
		g.DELETE("/delete-from-board/:task_id/:status_id", DeleteStatusFromBoard)
	}
}

var BaseStartStatus uuid.UUID
var BaseEndStatus uuid.UUID

// GetAllStatuses godoc
// @Summary Получение всех статусов
// @Description Получение списка всех статусов с пагинацией (только неудаленные)
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
	if err := Authorize(c); err != nil { return err }

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге номера страницы"})
	}
	offset := (page - 1) * pageSize

	var total int64
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Where("deleted = FALSE").
		Count(&total).Error; err != nil {
		log.Printf("DB error (count statuses): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте статусов"})
	}

	var rows []models.Status
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Where("deleted = FALSE").
		Limit(pageSize).Offset(offset).
		Find(&rows).Error; err != nil {
		log.Printf("DB error (find statuses): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов"})
	}

	out := response.StatusListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
	}
	for _, s := range rows {
		out.Statuses = append(out.Statuses, response.StatusResponse{
			ID:        s.ID.String(),
			Key:       utils.GetString(s.Key),
			Name:      utils.GetString(s.Name),
			Color:     utils.GetString(s.Color),
			IsDefault: utils.GetBool(s.IsDefault),
			IsActive:  utils.GetBool(s.IsActive),
			IsOpen:    utils.GetBool(s.IsOpen),
			CreatedAt: utils.GetTime(s.CreatedAt),
			UpdatedAt: utils.GetTime(s.UpdatedAt),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// GetStatusesByBoardId godoc
// @Summary Получение статусов по ID доски
// @Description Получение списка статусов, связанных с доской по ее ID
// @Tags Statuses
// @Produce json
// @Param board_id path string true "ID доски"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге ID доски"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.StatusByBoardIdResponse "Список статусов для доски"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении статусов"
// @Router /status/project/{board_id} [get]
func GetStatusesByBoardId(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	boardUUID, err := uuid.Parse(c.Param("board_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора доски"})
	}

	var links []models.StatusBoard
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.StatusBoard{}).
		Preload("Status", "deleted = FALSE").
		Where("status_boards.deleted = FALSE AND status_boards.board_id = ?", boardUUID).
		Find(&links).Error; err != nil {
		log.Printf("DB error (statuses by board): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов для доски"})
	}

	resp := response.StatusByBoardIdResponse{BoardId: c.Param("board_id")}
	for _, sb := range links {
		if sb.Status != nil {
			resp.Statuses = append(resp.Statuses, response.StatusResponse{
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
	return c.JSON(http.StatusOK, resp)
}

// GetStatusesByTaskId godoc
// @Summary Получение статусов по ID задачи
// @Description Получение списка статусов, связанных с задачей по ее ID
// @Tags Statuses
// @Produce json
// @Param task_id path string true "ID задачи"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Ошибка при парсинге ID задачи"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.StatusByTaskIdResponse "Список статусов для задачи"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении статусов"
// @Router /status/task/{task_id} [get]
func GetStatusesByTaskId(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	taskUUID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге идентификатора задачи"})
	}

	var links []models.StatusTask
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.StatusTask{}).
		Preload("Status", "deleted = FALSE").
		Where("deleted = FALSE AND task_id = ?", taskUUID).
		Find(&links).Error; err != nil {
		log.Printf("DB error (statuses by task): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов для задачи"})
	}

	resp := response.StatusByTaskIdResponse{TaskId: c.Param("task_id")}
	for _, st := range links {
		if st.Status.ID != uuid.Nil { // Status preloaded
			resp.Statuses = append(resp.Statuses, response.StatusResponse{
				ID:        st.Status.ID.String(),
				Key:       utils.GetString(st.Status.Key),
				Name:      utils.GetString(st.Status.Name),
				Color:     utils.GetString(st.Status.Color),
				IsDefault: utils.GetBool(st.Status.IsDefault),
				IsActive:  utils.GetBool(st.Status.IsActive),
				IsOpen:    utils.GetBool(st.Status.IsOpen),
				CreatedAt: utils.GetTime(st.Status.CreatedAt),
				UpdatedAt: utils.GetTime(st.Status.UpdatedAt),
			})
		}
	}
	return c.JSON(http.StatusOK, resp)
}

// GetStatusByID godoc
// @Summary Получение статуса по ID
// @Description Получение информации о статусе по его ID
// @Tags Statuses
// @Produce json
// @Param id path string true "ID статуса"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Некорректный идентификатор статуса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.StatusResponse "Информация о статусе"
// @Failure 404 {object} map[string]string "Статус не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении статуса"
// @Router /status/{id} [get]
func GetStatusByID(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	statusID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор статуса"})
	}

	var s models.Status
	res := DBConn.Session(&gorm.Session{}).
		Model(&models.Status{}).
		Where("deleted = FALSE AND id = ?", statusID).
		First(&s)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Статус не найден"})
		}
		log.Printf("DB error (find status by id): %v", res.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статуса"})
	}

	return c.JSON(http.StatusOK, response.StatusResponse{
		ID:        s.ID.String(),
		Key:       utils.GetString(s.Key),
		Name:      utils.GetString(s.Name),
		Color:     utils.GetString(s.Color),
		IsDefault: utils.GetBool(s.IsDefault),
		IsActive:  utils.GetBool(s.IsActive),
		IsOpen:    utils.GetBool(s.IsOpen),
		CreatedAt: utils.GetTime(s.CreatedAt),
		UpdatedAt: utils.GetTime(s.UpdatedAt),
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
// @Description Логическое удаление статуса по ID, включая связанные данные (поле deleted = true)
// @Tags Statuses
// @Accept json
// @Produce json
// @Param id path string true "ID статуса"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Некорректный идентификатор статуса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.StatusUniversalResponse "Статус успешно удален"
// @Failure 404 {object} map[string]string "Статус не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении статуса"
// @Router /status/{id} [delete]
func DeleteStatus(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	statusID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор статуса"})
	}

	delTrue := true
	now := time.Now()
	update := map[string]interface{}{"deleted": &delTrue, "updated_at": &now}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		// сам статус
		res := tx.Session(&gorm.Session{}).
			Model(&models.Status{}).
			Where("id = ? AND deleted = FALSE", statusID).
			Updates(update)
		if res.Error != nil { return res.Error }
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		// связи
		if res = tx.Session(&gorm.Session{}).
			Model(&models.StatusBoard{}).
			Where("status_id = ?", statusID).
			Updates(update); res.Error != nil {
			return res.Error
		}
		if res = tx.Session(&gorm.Session{}).
			Model(&models.StatusTask{}).
			Where("status_id = ?", statusID).
			Updates(map[string]interface{}{"deleted": true}); res.Error != nil {
			return res.Error
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
}
