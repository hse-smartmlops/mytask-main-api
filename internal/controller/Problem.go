package controller

import (
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterProblemRoutes(e *echo.Echo){
	problemGroup := e.Group("/problem")
	{
		problemGroup.GET("", GetAllProblems)
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

		})
	}
}