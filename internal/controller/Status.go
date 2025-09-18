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
	statusGroup := e.Group("/status")
	statusGroup.Use(KeycloakAuthMiddleware)
	{
		statusGroup.GET("/all/:page/:pagesize", getAllStatuses)
		statusGroup.GET("/:id", getStatusByID)
		statusGroup.POST("", createStatus)	
		statusGroup.PATCH("/:id", updateStatus)
		statusGroup.DELETE("/:id", deleteStatus)
		statusGroup.GET("/project/:board_id", getStatusesByBoardId)
		statusGroup.GET("/task/:task_id", getStatusesByTaskId)
		statusGroup.POST("/add-to-task", addStatusToTask)
		statusGroup.POST("/add-to-board", addStatusToBoard)
	}
}

// getAllStatuses godoc
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
func getAllStatuses(c echo.Context) error{
	if err := authorize(c); err != nil{
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
	result := dbConn.Session(&gorm.Session{}).Model(models.Status{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count statuses): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете статусов",
		})
	}

	var statuses []models.Status
	if err := dbConn.Session(&gorm.Session{}).Model(models.Status{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&statuses).Error; err != nil {
			log.Printf("DB error (find statuses): %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении статусов из базы данных",
			})
	}

	statusesList := response.StatusListResponse{
		Page: page,
		PageSize: pageSize,
		TotalCount: totalCount,
	}

	for _, status := range statuses{
		statusesList.Statuses = append(statusesList.Statuses, response.StatusResponse{
			ID:        status.ID.String(),
			Key:       utils.GetString(status.Key),
			Name:      utils.GetString(status.Name),
			Color:     utils.GetString(status.Color),
			IsDefault: utils.GetBool(status.IsDefault),
			IsActive:  utils.GetBool(status.IsActive),
			IsOpen:    utils.GetBool(status.IsOpen),
			CreatedAt: utils.GetTime(status.CreatedAt),
			UpdatedAt: utils.GetTime(status.UpdatedAt),
		})
	}
	return c.JSON(http.StatusOK, statusesList)
}

// getStatusesByBoardId godoc
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
func getStatusesByBoardId(c echo.Context) error {
    // Authorization check
    if err := authorize(c); err != nil {
        return err
    }

    boardId := c.Param("board_id")
    boardUUID, err := uuid.Parse(boardId)
    if err != nil {
        log.Printf("failed to parse board uuid: %v", err)
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Ошибка при парсинге идентификатора доски",
        })
    }

    // Fetch all StatusBoard entries with preloaded Status in one query
    var statusBoards []models.StatusBoard
    if err := dbConn.Model(models.StatusBoard{}).
        Preload("Status",    "deleted = ?", false).Session(&gorm.Session{}).
        Where("status_boards.deleted = ? AND status_boards.board_id = ?", false, boardUUID).
        Find(&statusBoards).Error; err != nil {
        log.Printf("failed to get statuses for board %s: %v", boardId, err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Ошибка при получении статусов для доски",
        })
    }

    // Build response
    getResponse := response.StatusByBoardIdResponse{
        BoardId: boardId,
    }

    for _, sb := range statusBoards {
        if sb.Status != nil {
            getResponse.Statuses = append(getResponse.Statuses, response.StatusResponse{
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

    return c.JSON(http.StatusOK, getResponse)
}

// getStatusesByTaskId godoc
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
func getStatusesByTaskId(c echo.Context) error {
    // Authorization check
    if err := authorize(c); err != nil {
        return err
    }

    taskId := c.Param("task_id")
    taskUUID, err := uuid.Parse(taskId)
    if err != nil {
        log.Printf("failed to parse task uuid: %v", err)
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Ошибка при парсинге идентификатора задачи",
        })
    }

    // Fetch all StatusTask entries with preloaded Status in one query
    var statusTasks []models.StatusTask
    if err := dbConn.Session(&gorm.Session{}).Model(models.StatusTask{}).
        Preload("Status",    "deleted = ?", false).
        Where("deleted = ? AND task_id = ?", false, taskUUID).
        Find(&statusTasks).Error; err != nil {
        log.Printf("failed to get statuses for task %s: %v", taskId, err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Ошибка при получении статусов для задачи",
        })
    }

    // Build response
    getResponse := response.StatusByTaskIdResponse{
        TaskId: taskId,
    }

    for _, st := range statusTasks {
        if st.Status != nil {
            getResponse.Statuses = append(getResponse.Statuses, response.StatusResponse{
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

    return c.JSON(http.StatusOK, getResponse)
}

// getStatusByID godoc
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
func getStatusByID(c echo.Context) error{
	if err := authorize(c); err != nil{
		return err
	}
	id := c.Param("id")
	statusID, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор статуса",
		})
	}

	var status models.Status
	result := dbConn.Session(&gorm.Session{}).Model(models.Status{}).Where("deleted = ? and id = ?", false, statusID).First(&status)
	if result.Error != nil{
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Статус не найден",
			})
		}
		log.Printf("DB error (find project by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении статуса из базы данных",
		})
	}

	statusResponse := response.StatusResponse{
		ID:        status.ID.String(),
		Key:       utils.GetString(status.Key),
		Name:      utils.GetString(status.Name),
		Color:     utils.GetString(status.Color),
		IsDefault: utils.GetBool(status.IsDefault),
		IsActive:  utils.GetBool(status.IsActive),
		IsOpen:    utils.GetBool(status.IsOpen),
		CreatedAt: utils.GetTime(status.CreatedAt),
		UpdatedAt: utils.GetTime(status.UpdatedAt),
	}
	return c.JSON(http.StatusOK, statusResponse)
}

// createStatus godoc
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
func createStatus(c echo.Context) error{
	if err := authorize(c); err != nil{
		log.Printf("Failed to authorize: %v", err)
		return err
	}

	var req request.CreateStatusRequest
	if err := c.Bind(&req); err != nil{
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при привязке данных запроса",
		})
	}

	newUUID := uuid.New()
	now := time.Now()
	del := false
	h := sha256.Sum256(newUUID[:])
	key := hex.EncodeToString(h[:])[:8]

	status := models.Status{
		ID: newUUID,
		Key: &key,
		Name: &req.Name,
		Color: &req.Color,
		IsDefault: &req.IsDefault,
		IsActive: &req.IsActive,
		IsOpen: &req.IsOpen,
		CreatedAt: &now,
		Deleted: &del,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error{
		if res := tx.Session(&gorm.Session{}).Model(models.Status{}).Create(&status); res.Error != nil{
			log.Printf("DB error (create status): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil{
		log.Printf("DB error (create status): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании статуса в базе данных",
		})
	}

	createResponse := response.StatusUniversalResponse{
		ID: newUUID.String(),
		Message: "Статус успешно создан",
	}
	return c.JSON(http.StatusCreated, createResponse)
}

// updateStatus godoc
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
func updateStatus(c echo.Context) error{
	if err := authorize(c); err != nil{
		return err
	}
	id := c.Param("id")
	statusID, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор статуса",
		})
	}

	var req request.UpdateStatusRequest
	if err := c.Bind(&req); err != nil{
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при привязке данных запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.Name != nil{
		updateData["name"] = *req.Name
	}
	if req.Color != nil{
		updateData["color"] = *req.Color
	}
	if req.IsDefault != nil{
		updateData["is_default"] = *req.IsDefault
	}
	if req.IsActive != nil{
		updateData["is_active"] = *req.IsActive
	}
	if req.IsOpen != nil{
		updateData["is_open"] = *req.IsOpen
	}

	if len(updateData) == 0{
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Нет данных для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error{
		res := tx.Session(&gorm.Session{}).Model(&models.Status{}).Where("id = ? and deleted = ?", statusID, false).Updates(updateData)
		if res.Error != nil{
			log.Printf("DB error (update status): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0{
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{
				"error": "Ничего не найдено для обновленияЭ",
			})
		}
		return nil
	}); txErr != nil{
		log.Printf("DB transaction error (update status): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении статуса в базе данных",
		})
	}

	updateResponse := response.StatusUniversalResponse{
		ID: id,
		Message: "Статус успешно обновлен",
	}
	return c.JSON(http.StatusOK, updateResponse)
}

// deleteStatus godoc
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
func deleteStatus(c echo.Context) error{
	if err := authorize(c); err != nil{
		return err
	}

	id := c.Param("id")
	statusID, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор статуса",
		})
	}
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error{
		res := tx.Session(&gorm.Session{}).Model(&models.Status{}).Where("id = ?", statusID).Updates(updateData)
		if res.Error != nil{
			log.Printf("DB error (delete status): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0{
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{
				"error": "Ничего не найдено для удаления",
			})
		}
		if res = tx.Session(&gorm.Session{}).Model(models.StatusBoard{}).Where("status_id = ?", statusID).Updates(updateData); res.Error != nil{
			log.Printf("DB error (delete status_board): %v", res.Error)
			return res.Error
		}
		if res = tx.Session(&gorm.Session{}).Model(models.StatusTask{}).Where("status_id = ?", statusID).Updates(updateData); res.Error != nil{
			log.Printf("DB error (delete status_task): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil{
		if he, ok := txErr.(*echo.HTTPError); ok{
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete status): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении статуса из базы данных",
		})
	}

	delResponse := response.StatusUniversalResponse{
		ID: id,
		Message: "Статус успешно удален",
	}

	return c.JSON(http.StatusOK, delResponse)
}

// addStatusToTask godoc
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
func addStatusToTask(c echo.Context) error{
	if err := authorize(c); err != nil{
		return err
	}	
	var req request.AddStatusToTaskRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}
	
	del := false

	taskID, err := uuid.Parse(req.TaskId)
	if err != nil{
		log.Printf("failed to parse task uuid: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге идентификатора задачи",
		})
	}

	statusID, err := uuid.Parse(req.StatusId)
	if err != nil{
		log.Printf("failed to parse status uuid: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге идентификатора статуса",
		})
	}

	now := time.Now()

	statusTask := models.StatusTask{
		TaskID: taskID,
		StatusID: statusID,
		Deleted: &del,
		CreatedAt: &now,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error{
		if res := tx.Session(&gorm.Session{}).Model(models.StatusTask{}).Create(&statusTask); res.Error != nil{
			log.Printf("DB error (add status to task): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil{
		log.Printf("DB error (add status to task): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при добавлении статуса к задаче в базе данных",
		})
	}

	addResponse := response.AddStatusToTaskResponse{
		TaskId: req.TaskId,
		StatusId: req.StatusId,
		Message: "Статус успешно добавлен к задаче",
	}
	return c.JSON(http.StatusCreated, addResponse)
}

// addStatusToBoard godoc
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
func addStatusToBoard(c echo.Context) error{
	if err := authorize(c); err != nil{
		return err
	}	
	var req request.AddStatusToBoardRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}
	
	del := false

	boardID, err := uuid.Parse(req.BoardId)
	if err != nil{
		log.Printf("failed to parse task uuid: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге идентификатора доски",
		})
	}

	statusID, err := uuid.Parse(req.StatusId)
	if err != nil{
		log.Printf("failed to parse status uuid: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге идентификатора статуса",
		})
	}

	now := time.Now()

	statusBoard := models.StatusBoard{
		BoardID: boardID,
		StatusID: statusID,
		Deleted: &del,
		CreatedAt: &now,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error{
		if res := tx.Session(&gorm.Session{}).Model(models.StatusBoard{}).Create(&statusBoard); res.Error != nil{
			log.Printf("DB error (add status to board): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil{
		log.Printf("DB error (add status to board): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при добавлении статуса к доске в базе данных",
		})
	}

	addResponse := response.AddStatusToBoardResponse{
		BoardId: req.BoardId,
		StatusId: req.StatusId,
		Message: "Статус успешно добавлен к доске",
	}

	return c.JSON(http.StatusCreated, addResponse)
}
