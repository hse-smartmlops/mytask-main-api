package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterProblemRoutes(e *echo.Echo){
	problemGroup := e.Group("/problem")
	{
		problemGroup.GET("", GetAllProblems)
		problemGroup.GET("/:id", GetProblemByID)
		problemGroup.GET("/user/:id", GetProblemsByUserId)
		problemGroup.POST("", CreateProblem)
		problemGroup.PUT("/:id", UpdateProblem)
		problemGroup.DELETE("/:id", DeleteProblem)
	}
}

func GetAllProblems(c echo.Context) error{
	var req request.ProblemsListRequest
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
	result := dbConn.Session(&gorm.Session{}).Model(models.Problem{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете проблем",
		})
	}

	var problems []models.Problem
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&problems).Error; err != nil{
		log.Printf("DB error (find projects): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проектов из базы данных",
		})
	}

	problemList := response.ProblemListResponse{
		Page: page,
		PageSize: pageSize,
		TotalCount: totalCount,
	}

	for _, problem := range problems{
		var problemID string
		if problem.ID != nil{
			problemID = problem.ID.String()
		}
		var description []string
		if problem.Description != nil{
			description = *problem.Description
		}
		var creatorID string
		if problem.CreatorID != nil{
			creatorID = problem.CreatorID.String()
		} 
		var name string
		if problem.Name != nil{
			name = *problem.Name
		}
		var createdAt time.Time
		if problem.CreatedAt != nil{
			createdAt = *problem.CreatedAt
		}
		var updatedAt time.Time
		if problem.UpdatedAt != nil{
			updatedAt = *problem.UpdatedAt
		}

		problemList.Problems = append(problemList.Problems, response.ProblemResponse{
			ID: problemID,
			Name: name,
			Description: description,
			CreatorId: creatorID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}
	return c.JSON(http.StatusOK, problemList)
}

func GetProblemsByUserId(c echo.Context) error{
	id := c.Param("id")
	var req request.ProblemListByUserId
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
	result := dbConn.Session(&gorm.Session{}).Model(models.Problem{}).Where("creator_id = ? and deleted = ?", id, false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете проблем",
		})
	}

	var problems []models.Problem
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&problems).Error; err != nil{
		log.Printf("DB error (find projects): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проектов из базы данных",
		})
	}

	problemList := response.ProblemsByUserId{
		UserID: id,
		Page: page,
		PageSize: pageSize,
		TotalCount: totalCount,
	}

	for _, problem := range problems{
		var problemID string
		if problem.ID != nil{
			problemID = problem.ID.String()
		}
		var description []string
		if problem.Description != nil{
			description = *problem.Description
		}
		var creatorID string
		if problem.CreatorID != nil{
			creatorID = problem.CreatorID.String()
		} 
		var name string
		if problem.Name != nil{
			name = *problem.Name
		}
		var createdAt time.Time
		if problem.CreatedAt != nil{
			createdAt = *problem.CreatedAt
		}
		var updatedAt time.Time
		if problem.UpdatedAt != nil{
			updatedAt = *problem.UpdatedAt
		}

		problemList.Problems = append(problemList.Problems, response.ProblemResponse{
			ID: problemID,
			Name: name,
			Description: description,
			CreatorId: creatorID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		})
	}
	return c.JSON(http.StatusOK, problemList)
}

func GetProblemByID(c echo.Context) error{
	id := c.Param("id")
	problemId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	var  problem models.Problem
	result := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).First(&problem, "id = ?", problemId)
	if result.Error != nil{
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Проблема не найдена",
			})
		}
		log.Printf("DB error (find problem by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проблемы из базы данных",
		})
	}

	if problem.ID == nil{
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Проблема не найдена",
		})
	}

	var description []string
	if problem.Description != nil{
		description = *problem.Description
	}
	var creatorId string
	if problem.CreatorID != nil{
		creatorId = problem.CreatorID.String()
	}
	var name string
	if problem.Name != nil{
		name = *problem.Name
	}

	var createdAt time.Time
	if problem.CreatedAt != nil{
		createdAt = *problem.CreatedAt
	}
	var updatedAt time.Time
	if problem.UpdatedAt != nil{
		updatedAt = *problem.UpdatedAt
	}

	problemResponse := response.ProblemResponse{
		ID: id,
		Description: description,
		Name: name,
		CreatorId: creatorId,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return c.JSON(http.StatusOK, problemResponse)
}

func CreateProblem(c echo.Context) error{
	var req request.ProblemCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	now := time.Now()

	creatorId, err := uuid.Parse(req.CreatorID)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор пользователя",
		})
	}

	problem := models.Problem{
		ID: &newUUID,
		Description: req.Description,
		CreatorID: &creatorId,
		Name: req.Name,
		CreatedAt: &now,
	}

	result := dbConn.Session(&gorm.Session{}).Create(&problem)
	if result.Error != nil{
		log.Printf("DB error (create problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании проблемы",
		})
	}

	createResponse := response.ProblemUniversalResponse{
		ID:   newUUID.String(),
		Message:  "Проблема создана",
	}

	return c.JSON(http.StatusCreated, createResponse)
}

func UpdateProblem(c echo.Context) error{
	id := c.Param("id")
	problemId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проблемы",
		})
	}

	var req request.ProblemUpdateRequest
	if err = c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.Description != nil{
		updateData["description"] = *req.Description
	}

	if len(updateData) == 0{
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	if err = dbConn.Session(&gorm.Session{}).Model(models.Problem{}).Where("id = ?", problemId).Updates(updateData).Error; err != nil{
		log.Printf("DB error (update problem): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении проблемы",
		})
	}

	updateReponse := response.ProblemUniversalResponse{
		ID: id,
		Message: "Проблема успешно создана",
	}

	return c.JSON(http.StatusOK, updateReponse)
}

func DeleteProblem(c echo.Context) error{
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	updateData["updated_at"] = time.Now()
	result := dbConn.Session(&gorm.Session{}).Model(models.Problem{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil{
		log.Printf("DB error (delete problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении проблемы",
		})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не удалено",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.ForumMessage{}).Where("problem_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении обсуждения данной проблемы",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.ReportProblem{}).Where("problem_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении связей отчет-проблема данной проблемы",
		})
	}

	delResponse := response.ProblemUniversalResponse{
		ID: id,
		Message: "Проблема успешно удалена",
	}

	return c.JSON(http.StatusOK, delResponse)
}