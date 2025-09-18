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
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func RegisterProjectRoutes(e *echo.Echo) {
	projectGroup := e.Group("/project")
	// apply Keycloak auth middleware to all project routes
	projectGroup.Use(KeycloakAuthMiddleware)
	{
		projectGroup.GET("/all/:page/:pagesize", getAllProjects)
		projectGroup.GET("/:id", getProjectByID)
		projectGroup.POST("", createProject)
		projectGroup.PATCH("/:id", updateProject)
		projectGroup.DELETE("/:id", deleteProject)
		projectGroup.GET("/user/:id", getProjectsByUser)
		projectGroup.GET("/team/:team_id", getTeamProjects)
	}
}

// getAllProjects godoc
// @Summary Получение списка всех проектов
// @Description Получает список всех проектов с учетом пагинации, исключая удаленные
// @Tags Projects
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ProjectListResponse "Список проектов успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении проектов"
// @Router /project/all/{page}/{pagesize} [get]
func getAllProjects(c echo.Context) error {
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
	result := dbConn.Session(&gorm.Session{}).Model(models.Project{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count projects): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете проектов",
		})
	}

	// Получаем список проектов с пагинацией
	var projects []models.Project
	if err := dbConn.Session(&gorm.Session{}).Model(models.Project{}).Where("deleted = ?", false).
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
		projectList.Projects = append(projectList.Projects, response.ProjectResponse{
			ID:              project.ID.String(),
			Name:            utils.GetString(project.Name),
			Description:     utils.GetString(project.Description), // исправлено
			GitlabProjectId: utils.GetInt(project.GitlabProjectID),
			CreatedBy: 		 utils.GetUUIDString(project.CreatedBy),
			Status: 		 utils.GetString(project.Status),
			GitlabUrl:       utils.GetString(project.GitlabURL),
			CreatedAt:       utils.GetTime(project.CreatedAt),
			UpdatedAt:       utils.GetTime(project.UpdatedAt),
		})
	}
	return c.JSON(http.StatusOK, projectList)
}

// getProjectByID godoc
// @Summary Получение проекта по ID
// @Description Получает данные проекта по его уникальному идентификатору
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "ID проекта"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ProjectResponse "Проект успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проекта"
// @Failure 404 {object} map[string]string "Проект не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении проекта"
// @Router /project/{id} [get]
func getProjectByID(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	projectId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	var project models.Project
	result := dbConn.Session(&gorm.Session{}).Model(models.Project{}).Where("deleted = ?", false).First(&project, "id = ?", projectId)
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

	projectResponse := response.ProjectResponse{
		ID:              project.ID.String(),
		Name:            utils.GetString(project.Name),
		Description:     utils.GetString(project.Description), // исправлено
		GitlabProjectId: utils.GetInt(project.GitlabProjectID),
		CreatedBy: 		 utils.GetUUIDString(project.CreatedBy),
		Status: 		 utils.GetString(project.Status),
		GitlabUrl:       utils.GetString(project.GitlabURL),
		CreatedAt:       utils.GetTime(project.CreatedAt),
		UpdatedAt:       utils.GetTime(project.UpdatedAt),
	}
	return c.JSON(http.StatusOK, projectResponse)
}

// getProjectsByUser godoc
// @Summary Получение проектов пользователя
// @Description Получает список проектов, в которых участвует пользователь через команды
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "Идентификатор пользователя (UUID)"
// @Security BearerAuth
// @Success 200 {array} response.ProjectResponse "Список проектов успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор пользователя"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Проекты не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении проектов"
// @Router /project/user/{id} [get]
func getProjectsByUser(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор пользователя",
		})
	}

	var projects []models.Project

	// Логика: User -> TeamMembers -> ProjectTeams -> Projects
	result := dbConn.Session(&gorm.Session{}).
		Model(&models.Project{}).
		Select("projects.*").
		Joins("JOIN project_teams pt ON pt.project_id = projects.id").
		Joins("JOIN teams t ON t.id = pt.team_id").
		Joins("JOIN team_members tm ON tm.team_id = t.id").
		Where("tm.user_id = ? AND projects.deleted = ? AND (tm.deleted = ? OR tm.deleted IS NULL) AND (pt.deleted = ? OR pt.deleted IS NULL) AND (t.deleted = ? OR t.deleted IS NULL)", 
			userId, false, false, false, false).
		Find(&projects)

	if result.Error != nil {
		log.Printf("DB error (find projects by user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проектов из базы данных",
		})
	}

	if len(projects) == 0 {
		return c.JSON(http.StatusOK, []response.ProjectResponse{})
	}

	// Маппинг в DTO
	var projectResponses []response.ProjectResponse
	for _, project := range projects {
		projectResponses = append(projectResponses, response.ProjectResponse{
			ID:              project.ID.String(),
			Name:            utils.GetString(project.Name),
			Description:     utils.GetString(project.Description),
			GitlabProjectId: utils.GetInt(project.GitlabProjectID),
			CreatedBy:       utils.GetUUIDString(project.CreatedBy),
			Status:          utils.GetString(project.Status),
			GitlabUrl:       utils.GetString(project.GitlabURL),
			CreatedAt:       utils.GetTime(project.CreatedAt),
			UpdatedAt:       utils.GetTime(project.UpdatedAt),
		})
	}

	return c.JSON(http.StatusOK, projectResponses)
}

// getTeamProjects godoc
// @Summary Получение списка проектов команды
// @Description Получает список всех проектов, связанных с командой, по team_id
// @Tags Teams
// @Accept json
// @Produce json
// @Param team_id path string true "ID команды"
// @Success 200 {object} response.ProjectByTeamResponse "Список проектов успешно получен"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Некорректный team_id"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Команда или проекты не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении проектов"
// @Router /project/team/{team_id} [get]
func getTeamProjects(c echo.Context) error {
	if err := authorize(c); err != nil {
        return err
    }

    teamIDStr := c.Param("team_id")
    teamID, err := uuid.Parse(teamIDStr)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Некорректный team_id",
        })
    }

	// Получаем связи команда-проект
	var projectTeams []models.ProjectTeam
	if err := dbConn.
		Session(&gorm.Session{}).Model(models.ProjectTeam{}).
		Where("team_id = ? AND deleted = ?", teamID, false).
		Find(&projectTeams).Error; err != nil {
		log.Printf("DB error (find projectTeams): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении связей команда-проект",
		})
	}

	if len(projectTeams) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Для данной команды проекты не найдены",
		})
	}

	// Собираем projectIDs
	projectIDs := make([]uuid.UUID, 0, len(projectTeams))
	for _, pt := range projectTeams {
		projectIDs = append(projectIDs, pt.ProjectID)
	}

	// Получаем проекты
	var projects []models.Project
	if err := dbConn.Model(models.Project{}).
		Session(&gorm.Session{}).
		Where("id IN ? AND deleted = ?", projectIDs, false).
		Find(&projects).Error; err != nil {
		log.Printf("DB error (find projects): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении списка проектов",
		})
	}

	// Формируем ответ
	projectListResponse := response.ProjectByTeamResponse{Projects: make([]response.ProjectResponse, 0, len(projects))}
	for _, project := range projects {
		projectListResponse.Projects = append(projectListResponse.Projects, response.ProjectResponse{
			ID:              project.ID.String(),
			Name:            utils.GetString(project.Name),
			Description:     utils.GetString(project.Description),
			GitlabProjectId: utils.GetInt(project.GitlabProjectID),
			CreatedBy:       utils.GetUUIDString(project.CreatedBy),
			Status:          utils.GetString(project.Status),
			GitlabUrl:       utils.GetString(project.GitlabURL),
			CreatedAt:       utils.GetTime(project.CreatedAt),
			UpdatedAt:       utils.GetTime(project.UpdatedAt),
		})
	}

	return c.JSON(http.StatusOK, projectListResponse)
}

// createProject godoc
// @Summary Создание нового проекта
// @Description Создает новый проект с указанными параметрами
// @Tags Projects
// @Accept json
// @Produce json
// @Param project body request.CreateProjectRequest true "Данные для создания проекта"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.ProjectUniversalResponse "Проект успешно создан"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании проекта"
// @Router /project [post]
func createProject(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
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

	if req.CreatedBy == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Отсутствует идентификатор создателя",
		})
	}
	creatorUUID, err := uuid.Parse(*req.CreatedBy)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор создателя",
		})
	}

	var creatorID *uuid.UUID = &creatorUUID


	project := models.Project{
		ID:              newUUID,
		Name:            req.Name,
		Description:     req.Description,
		CreatedAt:       &now,
		CreatedBy:       creatorID,
		Status: 		 req.Status,
		GitlabProjectID: req.Gitlab_project_id,
		GitlabURL:       req.Gitlab_url,
		Deleted: &del,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).Model(models.Project{}).Omit(clause.Associations).Create(&project); res.Error != nil {
			log.Printf("DB error (create project): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create project): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании проекта"})
	}

	createResponse := response.ProjectUniversalResponse{
		ID:      project.ID.String(),
		Message: "Проект создан",
	}

	return c.JSON(http.StatusCreated, createResponse)
}

// updateProject godoc
// @Summary Обновление проекта
// @Description Обновляет данные проекта по его ID
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "ID проекта"
// @Param project body request.UpdateProjectRequest true "Данные для обновления проекта"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ProjectUniversalResponse "Проект успешно обновлен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении проекта"
// @Router /project/{id} [patch]
func updateProject(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
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
	if req.GitlabProjectId != nil {
		updateData["gitlab_project_id"] = *req.GitlabProjectId
	}
	if req.GitlabUrl != nil {
		updateData["gitlab_url"] = *req.GitlabUrl
	}
	if req.Status != nil{
		updateData["status"] = *req.Status
	}

	// Проверяем, есть ли поля для обновления
	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	// Выполняем обновление только указанных полей
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.Project{}).Where("id = ? AND deleted = ?", projectId, false).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (update project): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update project): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении проекта"})
	}

	// Формируем ответ
	updateResponse := response.ProjectUniversalResponse{
		ID:      id,
		Message: "Проект с ID " + id + " обновлен",
	}

	return c.JSON(http.StatusOK, updateResponse)
}

// deleteProject godoc
// @Summary Удаление проекта
// @Description Логическое удаление проекта по ID, включая связанные данные (поле deleted = true)
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "ID проекта"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.ProjectUniversalResponse "Проект успешно удален"
// @Failure 404 {object} map[string]string "Проект не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении проекта"
// @Router /project/{id} [delete]
func deleteProject(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	id := c.Param("id")
	projectId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.Project{}).Where("id = ?", projectId).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete project): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}

		if res = tx.Session(&gorm.Session{}).Model(models.Task{}).Where("project_id = ?", projectId).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete project - tasks): %v", res.Error)
			return res.Error
		}

		if res = tx.Session(&gorm.Session{}).Model(models.Board{}).Where("project_id = ?", projectId).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete project - boards): %v", res.Error)
			return res.Error
		}

		if res = tx.Session(&gorm.Session{}).Model(models.ProjectTeam{}).Where("project_id = ?", projectId).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete project - project_teams): %v", res.Error)
			return res.Error
		}

		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete project): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении проекта"})
	}

	delResponse := response.ProjectUniversalResponse{ID: id, Message: "Проект с ID " + id + " удален"}

	return c.JSON(http.StatusOK, delResponse)
}
