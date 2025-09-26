package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	utils "emplacc-api/internal/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterAttendanceRoutes(e *echo.Echo) {
	attendanceGroup := e.Group("/attendance")
	attendanceGroup.Use(KeycloakAuthMiddleware)
	{
		attendanceGroup.GET("/all/:page/:pagesize", GetAllAttendances)
		attendanceGroup.GET("/user/:id", GetAttendancesByUserId)
		attendanceGroup.POST("", CreateAttendance)
		attendanceGroup.PATCH("/:id", UpdateAttendance)
		attendanceGroup.DELETE("/:id", DeleteAttendance)
	}
}

// getAllAttendances godoc
// @Summary Получение всех посещений
// @Description Получение списка всех посещений с пагинацией (логически не удаленных)
// @Tags Attendance
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.AttendancesListResponse "Список посещений успешно получен"
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров пагинации"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении посещений"
// @Router /attendance/all/{page}/{pagesize} [get]
func GetAllAttendances(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		log.Printf("failed to parse page: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		log.Printf("failed to parse pagesize: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге размера страницы"})
	}
	offset := (page - 1) * pageSize

	// Базовый запрос
	dbq := DBConn.Session(&gorm.Session{}).
		Model(&models.Attendance{}).
		Where("deleted = FALSE")

	// Count
	var totalCount int64
	if err := dbq.Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count attendances): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете посещений"})
	}

	// Выборка
	var attendances []models.Attendance
	if err := dbq.
		Order("created_at DESC NULLS LAST").
		Limit(pageSize).
		Offset(offset).
		Find(&attendances).Error; err != nil {
		log.Printf("DB error (find attendances): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении посещений из базы данных"})
	}

	resp := response.AttendancesListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, a := range attendances {
		resp.Attendances = append(resp.Attendances, response.AttendanceResponse{
			ID:            a.ID.String(),
			UserID:        a.UserID.String(),
			Date:          utils.GetTime(a.Date),
			WorkdayHours:  utils.GetInt16(a.WorkdayHours),
			PlannedStart:  utils.GetTime(a.PlannedStart),
			ActualStart:   utils.GetTime(a.ActualStart),
			Commits:       utils.GetInt16(a.Commits),
			MergeRequests: utils.GetInt16(a.MergeRequests),
			CodeReviews:   utils.GetInt16(a.CodeReviews),
			EndWork:       utils.GetTime(a.EndWork),
			UpdatedAt:     utils.GetTime(a.UpdatedAt),
			CreatedAt:     utils.GetTime(a.CreatedAt),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// getAttendancesByUserId godoc
// @Summary Получение посещений по ID пользователя
// @Description Получение списка посещений для конкретного пользователя (логически не удаленных)
// @Tags Attendance
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Security BearerAuth
// @Success 200 {object} response.AttendancesByUserId "Список посещений пользователя успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор пользователя"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении посещений"
// @Router /attendance/user/{id} [get]
func GetAttendancesByUserId(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор пользователя"})
	}

	dbq := DBConn.Session(&gorm.Session{}).
		Model(&models.Attendance{}).
		Where("deleted = FALSE AND user_id = ?", userID)

	var attendances []models.Attendance
	if err := dbq.
		Order("date DESC NULLS LAST, created_at DESC NULLS LAST").
		Find(&attendances).Error; err != nil {
		log.Printf("DB error (find attendances by user): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении посещений из базы данных"})
	}

	resp := response.AttendancesByUserId{UserID: userID.String()}
	for _, a := range attendances {
		resp.Attendances = append(resp.Attendances, response.AttendanceResponse{
			ID:            a.ID.String(),
			UserID:        a.UserID.String(),
			Date:          utils.GetTime(a.Date),
			WorkdayHours:  utils.GetInt16(a.WorkdayHours),
			PlannedStart:  utils.GetTime(a.PlannedStart),
			ActualStart:   utils.GetTime(a.ActualStart),
			Commits:       utils.GetInt16(a.Commits),
			MergeRequests: utils.GetInt16(a.MergeRequests),
			CodeReviews:   utils.GetInt16(a.CodeReviews),
			EndWork:       utils.GetTime(a.EndWork),
			UpdatedAt:     utils.GetTime(a.UpdatedAt),
			CreatedAt:     utils.GetTime(a.CreatedAt),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// createAttendance godoc
// @Summary Создание посещения
// @Description Создание нового посещения
// @Tags Attendance
// @Accept json
// @Produce json
// @Param body body request.AttendanceCreateRequest true "Данные для создания посещения"
// @Security BearerAuth
// @Success 201 {object} response.AttendanceUniversalResponse "Посещение успешно создано"
// @Failure 400 {object} map[string]string "Некорректные данные запроса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании посещения"
// @Router /attendance [post]
func CreateAttendance(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	var req request.AttendanceCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		log.Printf("parse user id error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор пользователя"})
	}

	now := time.Now()
	del := false

	attendance := models.Attendance{
		ID:            uuid.New(),
		UserID:        userID,
		Date:          req.Date,
		WorkdayHours:  req.WorkdayHours,
		PlannedStart:  req.PlannedStart,
		ActualStart:   req.ActualStart,
		Commits:       req.Commits,
		MergeRequests: req.MergeRequests,
		CodeReviews:   req.CodeReviews,
		EndWork:       req.EndWork,
		Deleted:       &del,
		CreatedAt:     &now,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Session(&gorm.Session{}).
			Model(&models.Attendance{}).
			Create(&attendance).Error; err != nil {
			log.Printf("DB error (create attendance): %v", err)
			return err
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create attendance): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании посещения"})
	}

	return c.JSON(http.StatusCreated, response.AttendanceUniversalResponse{
		ID:      attendance.ID.String(),
		Message: "Attendance created",
	})
}

// updateAttendance godoc
// @Summary Обновление посещения
// @Description Обновление полей посещения по ID
// @Tags Attendance
// @Accept json
// @Produce json
// @Param id path string true "ID посещения"
// @Param body body request.AttendanceUpdateRequest true "Данные для обновления посещения"
// @Security BearerAuth
// @Success 200 {object} response.AttendanceUniversalResponse "Посещение успешно обновлено"
// @Failure 400 {object} map[string]string "Некорректные данные запроса или ID"
// @Failure 404 {object} map[string]string "Посещение не найдено"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении посещения"
// @Router /attendance/{id} [patch]
func UpdateAttendance(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	attendanceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор посещения"})
	}

	var req request.AttendanceUpdateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	updateData := make(map[string]any)
	if req.Date != nil {
		updateData["date"] = req.Date
	}
	if req.WorkdayHours != nil {
		updateData["workday_hours"] = req.WorkdayHours
	}
	if req.PlannedStart != nil {
		updateData["planned_start"] = req.PlannedStart
	}
	if req.ActualStart != nil {
		updateData["actual_start"] = req.ActualStart
	}
	// ВНИМАНИЕ: поля Status в модели Attendance нет — не обновляем его
	if req.Commits != nil {
		updateData["commits"] = req.Commits
	}
	if req.MergeRequests != nil {
		updateData["merge_requests"] = req.MergeRequests
	}
	if req.CodeReviews != nil {
		updateData["code_reviews"] = req.CodeReviews
	}
	if req.EndWork != nil {
		updateData["end_work"] = req.EndWork
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
	}
	now := time.Now()
	updateData["updated_at"] = &now

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.Attendance{}).
			Where("id = ? AND deleted = FALSE", attendanceID).
			Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (update attendance): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Посещение не найдено или уже удалено"})
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (update attendance): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении посещения"})
	}

	return c.JSON(http.StatusOK, response.AttendanceUniversalResponse{
		ID:      attendanceID.String(),
		Message: "Посещение обновлено",
	})
}

// deleteAttendance godoc
// @Summary Удаление посещения
// @Description Логическое удаление посещения по ID (поле deleted = true)
// @Tags Attendance
// @Accept json
// @Produce json
// @Param id path string true "ID посещения"
// @Security BearerAuth
// @Success 200 {object} response.AttendanceUniversalResponse "Посещение успешно удалено"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 404 {object} map[string]string "Посещение не найдено"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении посещения"
// @Router /attendance/{id} [delete]
func DeleteAttendance(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	attendanceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор посещения"})
	}

	now := time.Now()
	del := true
	updateData := map[string]any{
		"deleted":    &del,
		"updated_at": &now,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.Attendance{}).
			Where("id = ? AND deleted = FALSE", attendanceID).
			Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete attendance): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "Посещение не найдено или уже удалено")
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete attendance): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении посещения"})
	}

	return c.JSON(http.StatusOK, response.AttendanceUniversalResponse{
		ID:      attendanceID.String(),
		Message: "Посещение удалено",
	})
}
