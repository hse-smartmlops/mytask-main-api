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
	"gorm.io/gorm"
)

func RegisterProjectRoutes(e *echo.Echo) {
	projectGroup := e.Group("/project")
	{
		projectGroup.GET("", GetAllProjects)
		projectGroup.GET("/:id", GetProjectByID)
		projectGroup.POST("", CreateProject)
		projectGroup.PATCH("/:id", UpdateProject)
		projectGroup.DELETE("/:id", DeleteProject)
	}
}

// GetAllProjects godoc
// @Summary Получение списка всех проектов
// @Description Получает список всех проектов с учетом пагинации, исключая удаленные
// @Tags Projects
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param pageSize query int false "Размер страницы" default(10)
// @Success 200 {object} response.ProjectListResponse "Список проектов успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении проектов"
// @Router /project [get]
func GetAllProjects(c echo.Context) error {
	var req request.ProjectListRequest
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
	result := dbConn.Session(&gorm.Session{}).Model(models.Project{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count projects): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете проектов",
		})
	}

	// Получаем список проектов с пагинацией
	var projects []models.Project
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&projects).Error; err != nil {
		log.Printf("DB error (find projects): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проектов из базы данных",
		})
	}

	// Формируем ответ
	projectList := response.ProjectListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, project := range projects {
		gitLabId := ""
		if project.GitlabProjectID != nil {
			gitLabId = strconv.Itoa(*project.GitlabProjectID)
		}
		gitLabUrl := ""
		if project.GitlabURL != nil {
			gitLabUrl = *project.GitlabURL
		}
		var priority int16
		if project.Priority != nil {
			priority = int16(*project.Priority)
		}

		var name string
		if project.Name != nil {
			name = *project.Name
		}

		var description string
		if project.Description != nil {
			description = *project.Description
		}
		var status string
		if project.Status != nil {
			status = *project.Status
		}

		var createdAt time.Time
		if project.CreatedAt != nil {
			createdAt = *project.CreatedAt
		}

		var updatedAt time.Time
		if project.UpdatedAt != nil {
			updatedAt = *project.UpdatedAt
		}

		projectList.Projects = append(projectList.Projects, response.ProjectResponse{
			ID:              project.ID.String(),
			Name:            name,
			Description:     description, // исправлено
			Status:          status,
			GitlabProjectId: gitLabId,
			GitlabUrl:       gitLabUrl,
			CreatedAt:       createdAt,
			Priority:        priority,
			UpdatedAt:       updatedAt,
		})
	}
	return c.JSON(http.StatusOK, projectList)
}

// GetProjectByID godoc
// @Summary Получение проекта по ID
// @Description Получает данные проекта по его уникальному идентификатору
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "ID проекта"
// @Success 200 {object} response.ProjectResponse "Проект успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проекта"
// @Failure 404 {object} map[string]string "Проект не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении проекта"
// @Router /project/{id} [get]
func GetProjectByID(c echo.Context) error {
	id := c.Param("id")
	projectId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	var project models.Project
	result := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).First(&project, "id = ?", projectId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Проект не найден",
			})
		}
		log.Printf("DB error (find project by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проекта из базы данных",
		})
	}

	if project.ID == nil {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Проект не найден",
		})
	}

	var description string
	if project.Description != nil {
		description = *project.Description
	}
	var gitLabId string
	if project.GitlabProjectID != nil {
		gitLabId = strconv.Itoa(*project.GitlabProjectID)
	}
	var gitLabUrl string
	if project.GitlabURL != nil {
		gitLabUrl = *project.GitlabURL
	}
	var priority int16
	if project.Priority != nil {
		priority = int16(*project.Priority)
	}
	var name string
	if project.Name != nil {
		name = *project.Name
	}
	var status string
	if project.Status != nil {
		status = *project.Status
	}
	var createdAt time.Time
	if project.CreatedAt != nil {
		createdAt = *project.CreatedAt
	}
	var updatedAt time.Time
	if project.UpdatedAt != nil {
		updatedAt = *project.UpdatedAt
	}

	projectResponse := response.ProjectResponse{
		ID:              id,
		Name:            name,
		Description:     description,
		Status:          status,
		GitlabProjectId: gitLabId,
		GitlabUrl:       gitLabUrl,
		CreatedAt:       createdAt,
		Priority:        priority,
		UpdatedAt:       updatedAt,
	}
	return c.JSON(http.StatusOK, projectResponse)
}

// CreateProject godoc
// @Summary Создание нового проекта
// @Description Создает новый проект с указанными параметрами
// @Tags Projects
// @Accept json
// @Produce json
// @Param project body request.CreateProjectRequest true "Данные для создания проекта"
// @Success 201 {object} response.ProjectUniversalResponse "Проект успешно создан"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании проекта"
// @Router /project [post]
func CreateProject(c echo.Context) error {
	var req request.CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	now := time.Now()

	del := false

	project := models.Project{
		ID:              &newUUID,
		Name:            req.Name,
		Description:     req.Description,
		CreatedAt:       &now,
		Status:          req.Status,
		GitlabProjectID: req.Gitlab_project_id,
		GitlabURL:       req.Gitlab_url,
		Priority:        req.Priority,
		Deleted: &del,
	}

	result := dbConn.Session(&gorm.Session{}).Create(&project)
	if result.Error != nil {
		log.Printf("DB error (create project): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании проекта",
		})
	}

	createResponse := response.ProjectUniversalResponse{
		ID:      project.ID.String(),
		Message: "Проект создан",
	}

	return c.JSON(http.StatusCreated, createResponse)
}

// UpdateProject godoc
// @Summary Обновление проекта
// @Description Обновляет данные проекта по его ID
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "ID проекта"
// @Param project body request.UpdateProjectRequest true "Данные для обновления проекта"
// @Success 200 {object} response.ProjectUniversalResponse "Проект успешно обновлен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении проекта"
// @Router /project/{id} [patch]
func UpdateProject(c echo.Context) error {
	// Парсинг ID проекта
	id := c.Param("id")
	projectId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	// Привязка данных запроса
	var req request.UpdateProjectRequest
	if err = c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	// Формируем карту только для указанных полей
	updateData := make(map[string]interface{})
	if req.Name != nil {
		updateData["name"] = *req.Name
	}
	if req.Description != nil {
		updateData["description"] = *req.Description
	}
	if req.Status != nil {
		updateData["status"] = *req.Status
	}
	if req.GitlabProjectId != nil {
		updateData["gitlab_project_id"] = *req.GitlabProjectId
	}
	if req.GitlabUrl != nil {
		updateData["gitlab_url"] = *req.GitlabUrl
	}
	if req.Priority != nil {
		updateData["priority"] = *req.Priority
	}

	// Проверяем, есть ли поля для обновления
	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	// Выполняем обновление только указанных полей
	if err = dbConn.Session(&gorm.Session{}).Model(models.Project{}).Where("id = ?", projectId).Updates(updateData).Error; err != nil {
		log.Printf("DB error (update project): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении проекта",
		})
	}

	// Формируем ответ
	updateResponse := response.ProjectUniversalResponse{
		ID:      id,
		Message: "Проект с ID " + id + " обновлен",
	}

	return c.JSON(http.StatusOK, updateResponse)
}

// DeleteProject godoc
// @Summary Удаление проекта
// @Description Логическое удаление проекта по ID, включая связанные данные (поле deleted = true)
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "ID проекта"
// @Success 200 {object} response.ProjectUniversalResponse "Проект успешно удален"
// @Failure 404 {object} map[string]string "Проект не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении проекта"
// @Router /project/{id} [delete]
func DeleteProject(c echo.Context) error {
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	updateData["updated_at"] = time.Now()
	result := dbConn.Session(&gorm.Session{}).Model(models.Project{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete project): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении проекта",
		})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не удалено",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.Task{}).Where("project_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete project): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении задач данного проекта",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.Board{}).Where("project_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete boards from project): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении досок данного проекта",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.ProjectTeam{}).Where("project_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete project from teams): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении данного проекта из команд",
		})
	}

	delResponse := response.ProjectUniversalResponse{ID: id, Message: "Проект с ID " + id + " удален"}

	return c.JSON(http.StatusOK, delResponse)
}
