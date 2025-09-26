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
	projectGroup.Use(KeycloakAuthMiddleware)
	{
		projectGroup.GET("/all/:page/:pagesize", GetAllProjects)
		projectGroup.GET("/:id", GetProjectByID)
		projectGroup.POST("", CreateProject)
		projectGroup.PATCH("/:id", UpdateProject)
		projectGroup.DELETE("/:id", DeleteProject)
		projectGroup.GET("/user/:id", GetProjectsByUser)
		projectGroup.GET("/team/:team_id", GetTeamProjects)
	}
}

// GetAllProjects godoc
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
func GetAllProjects(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Ошибка при парсинге номера страницы"})
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Project{}).
		Where("deleted = FALSE").
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count projects): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error":"Ошибка при подсчете проектов"})
	}

	var projects []models.Project
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Project{}).
		Where("deleted = FALSE").
		Limit(pageSize).Offset(offset).
		Find(&projects).Error; err != nil {
		log.Printf("DB error (find projects): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error":"Ошибка при получении проектов из базы данных"})
	}

	out := response.ProjectListResponse{Page: page, PageSize: pageSize, TotalCount: totalCount}
	for _, p := range projects {
		out.Projects = append(out.Projects, response.ProjectResponse{
			ID:              p.ID.String(),
			Name:            utils.GetString(p.Name),
			Description:     utils.GetString(p.Description),
			GitlabProjectId: utils.GetInt(p.GitlabProjectID),
			CreatedBy:       utils.GetUUIDString(p.CreatedBy),
			Status:          utils.GetString(p.Status),
			GitlabUrl:       utils.GetString(p.GitlabURL),
			CreatedAt:       utils.GetTime(p.CreatedAt),
			UpdatedAt:       utils.GetTime(p.UpdatedAt),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// GetProjectByID godoc
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
func GetProjectByID(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Некорректный идентификатор проекта"})
	}

	var p models.Project
	res := DBConn.Session(&gorm.Session{}).
		Model(&models.Project{}).
		Where("id = ? AND deleted = FALSE", projectID).
		First(&p)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error":"Проект не найден"})
		}
		log.Printf("DB error (find project by id): %v", res.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error":"Ошибка при получении проекта из базы данных"})
	}

	return c.JSON(http.StatusOK, response.ProjectResponse{
		ID:              p.ID.String(),
		Name:            utils.GetString(p.Name),
		Description:     utils.GetString(p.Description),
		GitlabProjectId: utils.GetInt(p.GitlabProjectID),
		CreatedBy:       utils.GetUUIDString(p.CreatedBy),
		Status:          utils.GetString(p.Status),
		GitlabUrl:       utils.GetString(p.GitlabURL),
		CreatedAt:       utils.GetTime(p.CreatedAt),
		UpdatedAt:       utils.GetTime(p.UpdatedAt),
	})
}

// GetProjectsByUser godoc
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
func GetProjectsByUser(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Некорректный идентификатор пользователя"})
	}

	var projects []models.Project
	res := DBConn.Session(&gorm.Session{}).
		Model(&models.Project{}).
		Select("projects.*").
		Joins("JOIN project_teams pt ON pt.project_id = projects.id").
		Joins("JOIN teams t ON t.id = pt.team_id").
		Joins("JOIN team_members tm ON tm.team_id = t.id").
		Where(`
			tm.user_id = ? 
			AND projects.deleted = FALSE 
			AND tm.deleted = FALSE 
			AND pt.deleted = FALSE 
			AND t.deleted = FALSE
		`, userID).
		Find(&projects)
	if res.Error != nil {
		log.Printf("DB error (find projects by user): %v", res.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error":"Ошибка при получении проектов из базы данных"})
	}
	if len(projects) == 0 {
		return c.JSON(http.StatusOK, []response.ProjectResponse{})
	}

	var out []response.ProjectResponse
	for _, p := range projects {
		out = append(out, response.ProjectResponse{
			ID:              p.ID.String(),
			Name:            utils.GetString(p.Name),
			Description:     utils.GetString(p.Description),
			GitlabProjectId: utils.GetInt(p.GitlabProjectID),
			CreatedBy:       utils.GetUUIDString(p.CreatedBy),
			Status:          utils.GetString(p.Status),
			GitlabUrl:       utils.GetString(p.GitlabURL),
			CreatedAt:       utils.GetTime(p.CreatedAt),
			UpdatedAt:       utils.GetTime(p.UpdatedAt),
		})
	}
	return c.JSON(http.StatusOK, out)
}

// GetTeamProjects godoc
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
func GetTeamProjects(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	teamID, err := uuid.Parse(c.Param("team_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Некорректный team_id"})
	}

	var pts []models.ProjectTeam
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.ProjectTeam{}).
		Where("team_id = ? AND deleted = FALSE", teamID).
		Find(&pts).Error; err != nil {
		log.Printf("DB error (find projectTeams): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error":"Ошибка при получении связей команда-проект"})
	}
	if len(pts) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{"error":"Для данной команды проекты не найдены"})
	}

	projectIDs := make([]uuid.UUID, 0, len(pts))
	for _, pt := range pts { projectIDs = append(projectIDs, pt.ProjectID) }

	var projects []models.Project
	if err := DBConn.Session(&gorm.Session{}).
		Model(&models.Project{}).
		Where("id IN ? AND deleted = FALSE", projectIDs).
		Find(&projects).Error; err != nil {
		log.Printf("DB error (find projects): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error":"Ошибка при получении списка проектов"})
	}

	resp := response.ProjectByTeamResponse{Projects: make([]response.ProjectResponse, 0, len(projects))}
	for _, p := range projects {
		resp.Projects = append(resp.Projects, response.ProjectResponse{
			ID:              p.ID.String(),
			Name:            utils.GetString(p.Name),
			Description:     utils.GetString(p.Description),
			GitlabProjectId: utils.GetInt(p.GitlabProjectID),
			CreatedBy:       utils.GetUUIDString(p.CreatedBy),
			Status:          utils.GetString(p.Status),
			GitlabUrl:       utils.GetString(p.GitlabURL),
			CreatedAt:       utils.GetTime(p.CreatedAt),
			UpdatedAt:       utils.GetTime(p.UpdatedAt),
		})
	}
	return c.JSON(http.StatusOK, resp)
}

// CreateProject godoc
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
func CreateProject(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	var req request.CreateProjectRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Не удалось получить данные из запроса"})
	}
	if req.CreatedBy == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Отсутствует идентификатор создателя"})
	}
	creatorUUID, err := uuid.Parse(*req.CreatedBy)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Некорректный идентификатор создателя"})
	}

	now := time.Now()
	del := false

	project := models.Project{
		ID:              uuid.New(),
		Name:            req.Name,
		Description:     req.Description,
		CreatedAt:       &now,
		CreatedBy:       &creatorUUID,
		Status:          req.Status,
		GitlabProjectID: req.Gitlab_project_id,
		GitlabURL:       req.Gitlab_url,
		Deleted:         &del,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.Project{}).
			Omit(clause.Associations).
			Create(&project)
		return res.Error
	}); txErr != nil {
		log.Printf("DB transaction error (create project): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error":"Ошибка при создании проекта"})
	}

	return c.JSON(http.StatusCreated, response.ProjectUniversalResponse{
		ID:      project.ID.String(),
		Message: "Проект создан",
	})
}

// UpdateProject godoc
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
func UpdateProject(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Некорректный идентификатор проекта"})
	}

	var req request.UpdateProjectRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Не удалось получить данные из запроса"})
	}

	updateData := map[string]interface{}{}
	if req.Name != nil            { updateData["name"] = *req.Name }
	if req.Description != nil     { updateData["description"] = *req.Description }
	if req.GitlabProjectId != nil { updateData["gitlab_project_id"] = *req.GitlabProjectId }
	if req.GitlabUrl != nil       { updateData["gitlab_url"] = *req.GitlabUrl }
	if req.Status != nil          { updateData["status"] = *req.Status }
	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Не указаны поля для обновления"})
	}

	now := time.Now()
	updateData["updated_at"] = &now

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.Project{}).
			Where("id = ? AND deleted = FALSE", projectID).
			Updates(updateData)
		if res.Error != nil { return res.Error }
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message":"Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update project): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error":"Ошибка при обновлении проекта"})
	}

	return c.JSON(http.StatusOK, response.ProjectUniversalResponse{
		ID:      projectID.String(),
		Message: "Проект успешно обновлен",
	})
}

// DeleteProject godoc
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
func DeleteProject(c echo.Context) error {
	if err := Authorize(c); err != nil { return err }

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error":"Некорректный идентификатор проекта"})
	}

	delTrue := true
	now := time.Now()
	update := map[string]interface{}{"deleted": &delTrue, "updated_at": &now}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		// проект
		res := tx.Session(&gorm.Session{}).
			Model(&models.Project{}).
			Where("id = ?", projectID).
			Updates(update)
		if res.Error != nil { return res.Error }
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message":"Ничего не удалено"})
		}
		// каскад
		if res = tx.Session(&gorm.Session{}).
			Model(&models.Task{}).
			Where("project_id = ?", projectID).
			Updates(update); res.Error != nil { return res.Error }

		if res = tx.Session(&gorm.Session{}).
			Model(&models.Board{}).
			Where("project_id = ?", projectID).
			Updates(update); res.Error != nil { return res.Error }

		if res = tx.Session(&gorm.Session{}).
			Model(&models.ProjectTeam{}).
			Where("project_id = ?", projectID).
			Updates(update); res.Error != nil { return res.Error }

		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete project): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error":"Ошибка при удалении проекта"})
	}

	return c.JSON(http.StatusOK, response.ProjectUniversalResponse{
		ID:      projectID.String(),
		Message: "Проект удален",
	})
}
