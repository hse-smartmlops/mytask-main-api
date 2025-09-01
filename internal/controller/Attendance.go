package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterAttendanceRoutes(e *echo.Echo){
	attendanceGroup := e.Group("/attendance")
	attendanceGroup.Use(KeycloakAuthMiddleware)
	{
		attendanceGroup.GET("/all/:page/:pagesize", GetAllAttendances)
		attendanceGroup.GET("/:id", GetAllAttendances)
		attendanceGroup.GET("/user/:id", GetAttendacesByUserId)
		attendanceGroup.POST("", CreateAttendance)
		attendanceGroup.PATCH("/:id", UpdateAttendance)
		attendanceGroup.DELETE("/:id", DeleteAttendance)
	}
}

// GetAllAttendances godoc
// @Summary Получение всех посещений
// @Description Получение списка всех посещений с пагинацией (логически не удаленных)
// @Tags attendance
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
func GetAllAttendances(c echo.Context) error{
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
	result := dbConn.Session(&gorm.Session{}).Model(models.Attendance{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count attendancese): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете посещений",
		})
	}

	var attendances []models.Attendance
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&attendances).Error; err != nil{
			log.Printf("DB error (find attendancese): %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении посещений из базы данных",
			})
	}

	attendanceList := response.AttendancesListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, attendance := range 	attendances{
		attendanceId := attendance.ID.String()
		userId := attendance.ID.String()
		var date time.Time
		if attendance.Date != nil{
			date = *attendance.Date
		}
		var workdayHours int16
		if attendance.WorkdayHours != nil{
			workdayHours = *attendance.WorkdayHours
		}
		var plannedStart time.Time
		if attendance.PlannedStart != nil{
			plannedStart = *attendance.PlannedStart
		}
		var actualStart time.Time
		if attendance.ActualStart != nil{ 
			actualStart = *attendance.ActualStart
		}
		var status string
		if attendance.Status != nil{
			status = *attendance.Status
		}
		var commits int16
		if attendance.Commits != nil{
			commits = *attendance.Commits
		}
		var mergeRequests int16
		if attendance.MergeRequests != nil{
			mergeRequests = *attendance.MergeRequests
		}
		var codeReviews int16
		if attendance.CodeReviews != nil{
			codeReviews = *attendance.CodeReviews
		}
		var endWork time.Time
		if attendance.EndWork != nil{
			endWork = *attendance.EndWork
		}
		var updatedAt time.Time
		if attendance.UpdatedAt != nil{
			updatedAt = *attendance.UpdatedAt
		}
		attendanceList.Attendances = append(attendanceList.Attendances, response.AttendanceResponse{
			ID: attendanceId,
			UserID: userId,
			Date: date,
			WorkdayHours: workdayHours,
			PlannedStart: plannedStart,
			ActualStart: actualStart,
			Status: status,
			Commits: commits,
			MergeRequests: mergeRequests,
			CodeReviews: codeReviews,
			EndWork: endWork,
			UpdatedAt: updatedAt,
		})
	}
	return c.JSON(http.StatusOK, attendanceList)
}

// GetAttendacesByUserId godoc
// @Summary Получение посещений по ID пользователя
// @Description Получение списка посещений для конкретного пользователя (логически не удаленных)
// @Tags attendance
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Security BearerAuth
// @Success 200 {object} response.AttendancesByUserId "Список посещений пользователя успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор пользователя"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении посещений"
// @Router /attendance/user/{id} [get]
func GetAttendacesByUserId(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор посещения",
		})
	}

	var attendances []models.Attendance
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ? and user_id = ?", false, userId).Find(&attendances).Error; err != nil{
		log.Printf("DB error (find attendances): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении посещений из базы данных",
		})
	}

	attendanceList := response.AttendancesByUserId{
		UserID: id,
	}

	for _, attendance := range 	attendances{
		attendanceId := attendance.ID.String()
		userId := attendance.ID.String()
		var date time.Time
		if attendance.Date != nil{
			date = *attendance.Date
		}
		var workdayHours int16
		if attendance.WorkdayHours != nil{
			workdayHours = *attendance.WorkdayHours
		}
		var plannedStart time.Time
		if attendance.PlannedStart != nil{
			plannedStart = *attendance.PlannedStart
		}
		var actualStart time.Time
		if attendance.ActualStart != nil{ 
			actualStart = *attendance.ActualStart
		}
		var status string
		if attendance.Status != nil{
			status = *attendance.Status
		}
		var commits int16
		if attendance.Commits != nil{
			commits = *attendance.Commits
		}
		var mergeRequests int16
		if attendance.MergeRequests != nil{
			mergeRequests = *attendance.MergeRequests
		}
		var codeReviews int16
		if attendance.CodeReviews != nil{
			codeReviews = *attendance.CodeReviews
		}
		var endWork time.Time
		if attendance.EndWork != nil{
			endWork = *attendance.EndWork
		}
		var updatedAt time.Time
		if attendance.UpdatedAt != nil{
			updatedAt = *attendance.UpdatedAt
		}
		attendanceList.Attendances = append(attendanceList.Attendances, response.AttendanceResponse{
			ID: attendanceId,
			UserID: userId,
			Date: date,
			WorkdayHours: workdayHours,
			PlannedStart: plannedStart,
			ActualStart: actualStart,
			Status: status,
			Commits: commits,
			MergeRequests: mergeRequests,
			CodeReviews: codeReviews,
			EndWork: endWork,
			UpdatedAt: updatedAt,
		})
	}
	return c.JSON(http.StatusOK, attendanceList)
}

// CreateAttendance godoc
// @Summary Создание посещения
// @Description Создание нового посещения
// @Tags attendance
// @Accept json
// @Produce json
// @Param body body request.AttendanceCreateRequest true "Данные для создания посещения"
// @Security BearerAuth
// @Success 201 {object} response.AttendanceUniversalResponse "Посещение успешно создано"
// @Failure 400 {object} map[string]string "Некорректные данные запроса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании посещения"
// @Router /attendance [post]
func CreateAttendance(c echo.Context) error{
	if err := authorize(c); err != nil{
		return err
	}
	var req request.AttendanceCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	del := false

	userId, err := uuid.Parse(req.UserId)
	if err != nil{
		log.Printf("error in parsing user id: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "failed to parse user id",
		})
	}

	attendance := models.Attendance{
		ID: newUUID,
		UserID: userId,
		Date: req.Date,
		WorkdayHours: req.WorkdayHours,
		PlannedStart: req.PlannedStart,
		ActualStart: req.ActualStart,
		Status: req.Status,
		Commits: req.Commits,
		MergeRequests: req.MergeRequests,
		CodeReviews: req.CodeReviews,
		EndWork: req.EndWork,
		Deleted: &del,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Create(&attendance); res.Error != nil {
			log.Printf("DB error (create attendance): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create attendance): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании attendance"})
	}

	createResponse := response.AttendanceUniversalResponse{
		ID: newUUID.String(),
		Message: "Attendance created",
	}

	return c.JSON(http.StatusCreated, createResponse)
}

// UpdateAttendance godoc
// @Summary Обновление посещения
// @Description Обновление полей посещения по ID
// @Tags attendance
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
func UpdateAttendance(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	attendanceId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор посещения",
		})
	}

	var req request.AttendanceUpdateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.Date != nil{
		updateData["date"] = req.Date
	}
	if req.WorkdayHours != nil{
		updateData["workday_hours"] = req.WorkdayHours
	}
	if req.PlannedStart != nil{
		updateData["planned_start"] = req.PlannedStart
	}
	if req.ActualStart != nil{
		updateData["actual_start"] = req.ActualStart
	}
	if req.Status != nil{
		updateData["status"] = req.Status
	}
	if req.Commits != nil{
		updateData["commits"] = req.Commits
	}
	if req.MergeRequests != nil{
		updateData["merge_requests"] = req.MergeRequests
	}
	if req.CodeReviews != nil{
		updateData["code_reviews"] = req.CodeReviews
	}
	if req.EndWork != nil{
		updateData["end_work"] = req.EndWork
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Attendance{}).Where("id = ? AND deleted = ?", attendanceId, false).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (update attendance): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}	
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update attendance): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении attendance"})
	}

	updateResponse := response.AttendanceUniversalResponse{
		ID: id,
		Message: "Посещение обновлено",
	}

	return c.JSON(http.StatusOK, updateResponse)
}

// DeleteAttendance godoc
// @Summary Удаление посещения
// @Description Логическое удаление посещения по ID (поле deleted = true)
// @Tags attendance
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
func DeleteAttendance(c echo.Context) error{
	if err := authorize(c); err != nil {
		return err
	}

	id := c.Param("id")
	attendanceId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор посещения",
		})
	}

	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(models.Attendance{}).Where("id = ?", attendanceId).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete attendance): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "Посещение не найдено")
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (delete attendance): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении посещения"})
	}

	deleteResponse := response.AttendanceUniversalResponse{
		ID: id,
		Message: "Посещение удалено",
	}

	return c.JSON(http.StatusOK, deleteResponse)
}