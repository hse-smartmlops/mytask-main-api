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

func RegisterProblemRoutes(e *echo.Echo) {
	problemGroup := e.Group("/problem")
	problemGroup.Use(KeycloakAuthMiddleware)
	{
		problemGroup.GET("/all/:page/:pagesize", getAllProblems)
		problemGroup.GET("/:id", getProblemByID)
		problemGroup.GET("/user/:id/:page/:pagesize", getProblemsByUserId)
		problemGroup.POST("", createProblem)
		problemGroup.PATCH("/:id", updateProblem)
		problemGroup.DELETE("/:id", deleteProblem)
	}
}

// getAllProblems godoc
// @Summary Получение списка всех проблем
// @Description Получает список всех проблем с учетом пагинации, исключая удаленные
// @Tags Problems
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ProblemListResponse "Список проблем успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении проблем"
// @Router /problem/all/{page}/{pagesize} [get]
func getAllProblems(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге номера страницы"})
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.Problem{}).
		Where("deleted = FALSE").
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count problem): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете проблем"})
	}

	var problems []models.Problem
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.Problem{}).
		Where("deleted = FALSE").
		Limit(pageSize).
		Offset(offset).
		Find(&problems).Error; err != nil {
		log.Printf("DB error (find problems): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении проблем из базы данных"})
	}

	out := response.ProblemListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, p := range problems {
		desc := []string(p.Description)
		var creatorId string
		if p.CreatorID != nil {
			creatorId = p.CreatorID.String()
		}
		out.Problems = append(out.Problems, response.ProblemResponse{
			ID:          p.ID.String(),
			Name:        utils.GetString(p.Name),
			Description: desc,
			CreatorId:   creatorId,
			CreatedAt:   utils.GetTime(p.CreatedAt),
			UpdatedAt:   utils.GetTime(p.UpdatedAt),
		})
	}

	return c.JSON(http.StatusOK, out)
}

// getProblemsByUserId godoc
// @Summary Получение проблем по ID пользователя
// @Description Получает список проблем, созданных указанным пользователем, с учетом пагинации
// @Tags Problems
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ProblemsByUserId "Список проблем успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении проблем"
// @Router /problem/user/{id}/{page}/{pagesize} [get]
func getProblemsByUserId(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	creatorUUID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор пользователя"})
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге номера страницы"})
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.Problem{}).
		Where("creator_id = ? AND deleted = FALSE", creatorUUID).
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count problem): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете проблем"})
	}

	var problems []models.Problem
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.Problem{}).
		Where("creator_id = ? AND deleted = FALSE", creatorUUID).
		Limit(pageSize).
		Offset(offset).
		Find(&problems).Error; err != nil {
		log.Printf("DB error (find problem): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении проблем из базы данных"})
	}

	out := response.ProblemsByUserId{
		UserID:     creatorUUID.String(),
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}
	for _, p := range problems {
		desc := []string(p.Description)
		var creatorId string
		if p.CreatorID != nil {
			creatorId = p.CreatorID.String()
		}
		out.Problems = append(out.Problems, response.ProblemResponse{
			ID:          p.ID.String(),
			Name:        utils.GetString(p.Name),
			Description: desc,
			CreatorId:   creatorId,
			CreatedAt:   utils.GetTime(p.CreatedAt),
			UpdatedAt:   utils.GetTime(p.UpdatedAt),
		})
	}

	return c.JSON(http.StatusOK, out)
}

// getProblemByID godoc
// @Summary Получение проблемы по ID
// @Description Получает данные проблемы по её уникальному идентификатору
// @Tags Problems
// @Accept json
// @Produce json
// @Param id path string true "ID проблемы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ProblemResponse "Проблема успешно получена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проблемы"
// @Failure 404 {object} map[string]string "Проблема не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении проблемы"
// @Router /problem/{id} [get]
func getProblemByID(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	problemId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проблемы"})
	}

	var p models.Problem
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.Problem{}).
		Where("id = ? AND deleted = FALSE", problemId).
		First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Проблема не найдена"})
		}
		log.Printf("DB error (find problem by id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении проблемы из базы данных"})
	}

	desc := []string(p.Description)
	var creatorId string
	if p.CreatorID != nil {
		creatorId = p.CreatorID.String()
	}

	return c.JSON(http.StatusOK, response.ProblemResponse{
		ID:          p.ID.String(),
		Name:        utils.GetString(p.Name),
		Description: desc,
		CreatorId:   creatorId,
		CreatedAt:   utils.GetTime(p.CreatedAt),
		UpdatedAt:   utils.GetTime(p.UpdatedAt),
	})
}

// createProblem godoc
// @Summary Создание новой проблемы
// @Description Создает новую проблему с указанными параметрами
// @Tags Problems
// @Accept json
// @Produce json
// @Param problem body request.ProblemCreateRequest true "Данные для создания проблемы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.ProblemUniversalResponse "Проблема успешно создана"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректный идентификатор пользователя"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании проблемы"
// @Router /problem [post]
func createProblem(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	var req request.ProblemCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	creatorId, err := uuid.Parse(req.CreatorID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор пользователя"})
	}

	now := time.Now()
	del := false

	var description pq.StringArray
	if req.Description != nil {
		description = pq.StringArray(*req.Description)
	} else {
		description = pq.StringArray{}
	}

	p := models.Problem{
		ID:          uuid.New(),
		Description: description,
		CreatorID:   &creatorId,
		Name:        req.Name,
		CreatedAt:   &now,
		Deleted:     &del,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).
			Model(&models.Problem{}).
			Create(&p); res.Error != nil {
			log.Printf("DB error (create problem): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create problem): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании проблемы"})
	}

	return c.JSON(http.StatusCreated, response.ProblemUniversalResponse{
		ID:      p.ID.String(),
		Message: "Проблема создана",
	})
}

// updateProblem godoc
// @Summary Обновление проблемы
// @Description Обновляет данные проблемы по её ID
// @Tags Problems
// @Accept json
// @Produce json
// @Param id path string true "ID проблемы"
// @Param problem body request.ProblemUpdateRequest true "Данные для обновления проблемы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ProblemUniversalResponse "Проблема успешно обновлена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении проблемы"
// @Router /problem/{id} [patch]
func updateProblem(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	problemId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проблемы"})
	}

	var req request.ProblemUpdateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	updateData := map[string]interface{}{}
	if req.Description != nil {
		updateData["description"] = pq.StringArray(*req.Description)
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
	}

	now := time.Now()
	updateData["updated_at"] = &now

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.Problem{}).
			Where("id = ? AND deleted = FALSE", problemId).
			Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (update problem): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update problem): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении проблемы"})
	}

	return c.JSON(http.StatusOK, response.ProblemUniversalResponse{
		ID:      problemId.String(),
		Message: "Проблема успешно обновлена",
	})
}

// deleteProblem godoc
// @Summary Удаление проблемы
// @Description Логическое удаление проблемы по ID, включая связанные данные (поле deleted = true)
// @Tags Problems
// @Accept json
// @Produce json
// @Param id path string true "ID проблемы"
// @Security BearerAuth
// @Success 200 {object} response.ProblemUniversalResponse "Проблема успешно удалена"
// @Failure 404 {object} map[string]string "Проблема не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении проблемы"
// @Router /problem/{id} [delete]
func deleteProblem(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	problemId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проблемы"})
	}

	delTrue := true
	now := time.Now()
	update := map[string]interface{}{
		"deleted":    &delTrue,
		"updated_at": &now,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		// помечаем проблему удалённой
		res := tx.Session(&gorm.Session{}).
			Model(&models.Problem{}).
			Where("id = ?", problemId).
			Updates(update)
		if res.Error != nil {
			log.Printf("DB error (delete problem): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}

		// каскадно помечаем связанные сущности
		if res = tx.Session(&gorm.Session{}).
			Model(&models.ForumMessage{}).
			Where("problem_id = ?", problemId).
			Updates(update); res.Error != nil {
			log.Printf("DB error (delete problem - forum messages): %v", res.Error)
			return res.Error
		}

		if res = tx.Session(&gorm.Session{}).
			Model(&models.ReportProblem{}).
			Where("problem_id = ?", problemId).
			Updates(update); res.Error != nil {
			log.Printf("DB error (delete problem - report_problem): %v", res.Error)
			return res.Error
		}

		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete problem): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении проблемы"})
	}

	return c.JSON(http.StatusOK, response.ProblemUniversalResponse{
		ID:      problemId.String(),
		Message: "Проблема успешно удалена",
	})
}
