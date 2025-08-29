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
	}
}

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