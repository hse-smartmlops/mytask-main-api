package controller

import (
	"emplacc-api/internal/db"
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"emplacc-api/internal/utils"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func RegisterTeamRoutes(e *echo.Echo) {
	teamGroup := e.Group("/team")
	teamGroup.Use(KeycloakAuthMiddleware)
	teamGroup.GET("/all", GetTeams)
	teamGroup.GET("/:id", GetTeamByID)
	teamGroup.POST("", CreateTeam)
	teamGroup.PATCH("/:id", UpdateTeam)
	teamGroup.DELETE("/:id", DeleteTeam) // Только для admin, добавить проверку роли
	teamGroup.POST("/user", AddUserToTeam)
	teamGroup.DELETE("/user", DeleteUserFromTeam)
	teamGroup.POST("/project", AddProjectToTeam)
	teamGroup.DELETE("/project", DeleteProjectFromTeam)
	teamGroup.GET("/project/:project_id", GetProjectTeams)
}

var DBConn *gorm.DB = db.DB_conn

// GetTeams godoc
// @Summary Получение списка всех команд
// @Description Получает список всех команд с их участниками
// @Tags Teams
// @Accept json
// @Produce json
// @Success 200 {object} response.TeamsListResponse "Список команд успешно получен"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении команд"
// @Router /team/all [get]
func GetTeams(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	var teams []models.Team

	err := DBConn.Session(&gorm.Session{}).
		Where("deleted = ?", false).
		Preload("TeamMembers", "deleted = ?", false).
		Preload("TeamMembers.User", "deleted = ?", false).
		Find(&teams).Error

	if err != nil {
		log.Printf("DB error (find teams with preload): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении списка команд",
		})
	}

	teamListResponse := response.TeamsListResponse{Teams: make([]response.TeamResponse, 0, len(teams))}
	for _, team := range teams {
		members := make([]response.TeamMemberResponse, 0, len(team.TeamMembers))
		for _, tm := range team.TeamMembers {
			user := tm.User
			members = append(members, response.TeamMemberResponse{
				UserID:         user.ID.String(),
				Specialization: utils.GetString(tm.Specialization),
				FirstName:      user.FirstName,
				LastName:       user.LastName,
				Email:          user.Email,
			})
		}

		teamListResponse.Teams = append(teamListResponse.Teams, response.TeamResponse{
			ID:          team.ID.String(),
			Name:        utils.GetString(team.Name),
			Description: utils.GetString(team.Description),
			UpdatedAt:   utils.GetTime(team.UpdatedAt),
			CreatedAt:   utils.GetTime(team.CreatedAt),
			Members:     members,
		})
	}

	return c.JSON(http.StatusOK, teamListResponse)
}

// GetProjectTeams godoc
// @Summary Получение списка команд проекта
// @Description Получает список всех команд, связанных с проектом, по project_id
// @Tags Projects
// @Accept json
// @Produce json
// @Param project_id path string true "ID проекта"
// @Success 200 {object} response.TeamsListResponse "Список команд успешно получен"
// @Security BearerAuth
// @Failure 400 {object} map[string]string "Некорректный project_id"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Проект или команды не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении команд"
// @Router /team/project/{project_id} [get]
func GetProjectTeams(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный project_id",
		})
	}

	var projectTeams []models.ProjectTeam
	err = DBConn.Session(&gorm.Session{}).
		Where("project_id = ? AND deleted = ?", projectID, false).
		Preload("Team", "deleted = ?", false).
		Preload("Team.TeamMembers", "deleted = ?", false).
		Preload("Team.TeamMembers.User", "deleted = ?", false).
		Find(&projectTeams).Error

	if err != nil {
		log.Printf("DB error (find projectTeams with preload): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении связей проект-команда",
		})
	}

	if len(projectTeams) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Для данного проекта команды не найдены",
		})
	}

	seenTeams := make(map[uuid.UUID]models.Team)
	for _, pt := range projectTeams {
		if pt.Team != nil && !*pt.Team.Deleted {
			seenTeams[pt.Team.ID] = *pt.Team
		}
	}

	if len(seenTeams) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Нет активных команд для данного проекта",
		})
	}

	teamListResponse := response.TeamsListResponse{Teams: make([]response.TeamResponse, 0, len(seenTeams))}
	for _, team := range seenTeams {
		members := make([]response.TeamMemberResponse, 0, len(team.TeamMembers))
		for _, tm := range team.TeamMembers {
			if tm.User == nil {
				continue 
			}
			members = append(members, response.TeamMemberResponse{
				UserID:         tm.User.ID.String(),
				Specialization: utils.GetString(tm.Specialization),
				FirstName:      tm.User.FirstName,                 
				LastName:       tm.User.LastName,                  
				Email:          tm.User.Email,                     
			})
		}

		teamListResponse.Teams = append(teamListResponse.Teams, response.TeamResponse{
			ID:          team.ID.String(),
			Name:        utils.GetString(team.Name),
			Description: utils.GetString(team.Description),
			UpdatedAt:   utils.GetTime(team.UpdatedAt),
			CreatedAt:   utils.GetTime(team.CreatedAt),
			Members:     members,
		})
	}

	return c.JSON(http.StatusOK, teamListResponse)
}

// GetTeamByID godoc
// @Summary Получение команды по ID
// @Description Получает данные команды по её уникальному идентификатору, включая участников
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "ID команды"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TeamResponse "Команда успешно получена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор команды"
// @Failure 404 {object} map[string]string "Команда не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении команды"
// @Router /team/{id} [get]
func GetTeamByID(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}

	teamIDParam := c.Param("id")
	teamUUID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор команды",
		})
	}

	var team models.Team
	err = DBConn.Session(&gorm.Session{}).
		Where("id = ? AND deleted = ?", teamUUID, false).
		Preload("TeamMembers", "deleted = ?", false).
		Preload("TeamMembers.User", "deleted = ?", false).
		First(&team).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Команда не найдена",
			})
		}
		log.Printf("DB error (get team with preload): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении данных команды",
		})
	}

	membersResp := make([]response.TeamMemberResponse, 0, len(team.TeamMembers))
	for _, tm := range team.TeamMembers {
		if tm.User == nil {
			log.Printf("Warning: TeamMember.UserID=%s has no associated User", tm.UserID)
			continue
		}

		membersResp = append(membersResp, response.TeamMemberResponse{
			UserID:         tm.User.ID.String(),
			Specialization: utils.GetString(tm.Specialization), 
			FirstName:      tm.User.FirstName,                  
			LastName:       tm.User.LastName,                   
			Email:          tm.User.Email,                      
		})
	}

	teamResponse := response.TeamResponse{
		ID:          team.ID.String(),
		Name:        utils.GetString(team.Name),
		Description: utils.GetString(team.Description),
		UpdatedAt:   utils.GetTime(team.UpdatedAt),
		CreatedAt:   utils.GetTime(team.CreatedAt),
		Members:     membersResp,
	}

	return c.JSON(http.StatusOK, teamResponse)
}
// CreateTeam godoc
// @Summary Создание новой команды
// @Description Создает новую команду с указанными параметрами
// @Tags Teams
// @Accept json
// @Produce json
// @Param team body request.TeamCreateRequest true "Данные для создания команды"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.TeamUniversalResponse "Команда успешно создана"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании команды"
// @Router /team [post]
func CreateTeam(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}
	var req request.TeamCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}
	newUUID := uuid.New()

	var name *string
	if req.Name != "" {
		name = &req.Name
	}
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	del := false
	now := time.Now()

	team := models.Team{
		ID:          newUUID,
		Name:        name,
		Description: description,
		Deleted:     &del,
		CreatedAt:   &now,
	}
	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).Model(models.Team{}).Omit(clause.Associations).Create(&team); res.Error != nil {
			log.Printf("DB error (create team): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create team): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании команды"})
	}

	createResponse := response.TeamUniversalResponse{
		ID:      newUUID.String(),
		Message: "Команда создана",
	}

	return c.JSON(http.StatusCreated, createResponse)
}

// UpdateTeam godoc
// @Summary Обновление команды
// @Description Обновляет данные команды по её ID
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "ID команды"
// @Param team body request.TeamUpdateRequest true "Данные для обновления команды"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TeamUniversalResponse "Команда успешно обновлена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении команды"
// @Router /team/{id} [patch]
func UpdateTeam(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}
	teamIDParam := c.Param("id")
	teamUUID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор команды"})
	}

	var req request.TeamUpdateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.Name != nil {
		updateData["name"] = *req.Name
	}
	if req.Description != nil {
		updateData["description"] = *req.Description
	}
	if len(updateData) == 0 {
		log.Printf("No fields provided for update")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}
	updateData["updated_at"] = time.Now()

	// существует и не удалена
	var team models.Team
	if err := DBConn.Session(&gorm.Session{}).Model(models.Team{}).
		Where("id = ? AND deleted = FALSE", teamUUID).
		First(&team).Error; err != nil {
		log.Printf("Team not found: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Команда не найдена",
		})
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.Team{}).Where("id = ?", teamUUID).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (update team): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update team): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении команды"})
	}

	updateResponse := response.TeamUniversalResponse{
		ID:      teamIDParam,
		Message: "Команда обновлена",
	}

	return c.JSON(http.StatusOK, updateResponse)
}

// DeleteTeam godoc
// @Summary Удаление команды
// @Description Логическое удаление команды по ID (поле deleted = true)
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "ID команды"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TeamUniversalResponse "Команда успешно удалена"
// @Failure 404 {object} map[string]string "Команда не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении команды"
// @Router /team/{id} [delete]
func DeleteTeam(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}
	teamIDParam := c.Param("id")
	updateData := map[string]interface{}{"deleted": true}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.Team{}).Where("id = ?", teamIDParam).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete team): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}

		if res = tx.Session(&gorm.Session{}).Model(models.TeamMember{}).Where("team_id = ?", teamIDParam).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete team_member connections): %v", res.Error)
			return res.Error
		}
		if res = tx.Session(&gorm.Session{}).Model(models.ProjectTeam{}).Where("team_id = ?", teamIDParam).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete project_team connections): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete team): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении команды"})
	}

	deleteResponse := response.TeamUniversalResponse{
		ID:      teamIDParam,
		Message: "Команда удалена",
	}

	return c.JSON(http.StatusOK, deleteResponse)
}

// AddUserToTeam godoc
// @Summary Добавление пользователя в команду
// @Description Добавляет пользователя в указанную команду
// @Tags Teams
// @Accept json
// @Produce json
// @Param addUser body request.TeamAddUserRequest true "Данные для добавления пользователя в команду"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TeamUniversalUserResponse "Пользователь успешно добавлен в команду"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 500 {object} map[string]string "Ошибка сервера при добавлении пользователя в команду"
// @Router /team/user [post]
func AddUserToTeam(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}
	var req request.TeamAddUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	var user models.User
	if err := DBConn.Session(&gorm.Session{}).Model(models.User{}).
		Select("id, profession").
		Where("id = ? AND deleted = FALSE", req.UserID).
		First(&user).Error; err != nil {
		log.Printf("DB error (select profession): %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при получении данных пользователя",
		})
	}

	var team models.Team
	if err := DBConn.Session(&gorm.Session{}).Model(models.Team{}).
		Where("id = ? AND deleted = FALSE", req.TeamID).
		First(&team).Error; err != nil {
		log.Printf("DB error (select team): %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при получении команды",
		})
	}

	del := false

	teamMember := models.TeamMember{
		UserID:         user.ID,
		TeamID:         team.ID,
		Specialization: &user.Profession,
		Deleted:        &del,
	}
	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).Model(models.TeamMember{}).Create(&teamMember); res.Error != nil {
			log.Printf("DB error (create team member): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (add user to team): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Не удалось добавить пользователя в команду"})
	}

	addResponse := response.TeamUniversalUserResponse{
		TeamID:  team.ID.String(),
		UserID:  req.UserID,
		Message: "Пользователь добавлен в команду",
	}

	return c.JSON(http.StatusOK, addResponse)
}

// DeleteUserFromTeam godoc
// @Summary Удаление пользователя из команды
// @Description Логически удаляет пользователя из команды (поле deleted = true)
// @Tags Teams
// @Accept json
// @Produce json
// @Param deleteUser body request.TeamDeleteUserRequest true "Данные для удаления пользователя из команды"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TeamUniversalUserResponse "Пользователь успешно удален из команды"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 404 {object} map[string]string "Ничего не удалено"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении пользователя из команды"
// @Router /team/user [delete]
func DeleteUserFromTeam(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}
	var req request.TeamDeleteUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	var team models.Team
	if err := DBConn.Session(&gorm.Session{}).Model(models.Team{}).
		Where("id = ? AND deleted = FALSE", req.TeamID).
		First(&team).Error; err != nil {
		log.Printf("DB error (select team): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении команды",
		})
	}

	var user models.User
	if err := DBConn.Session(&gorm.Session{}).Model(models.User{}).
		Where("id = ? AND deleted = FALSE", req.UserID).
		First(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении пользователя",
		})
	}

	updateData := map[string]interface{}{"deleted": true}
	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.TeamMember{}).
			Where("user_id = ? AND team_id = ?", user.ID, team.ID).
			Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete team member): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete team member): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении пользователя из команды"})
	}

	deleteResponse := response.TeamUniversalUserResponse{
		TeamID:  req.TeamID,
		UserID:  req.UserID,
		Message: "Пользователь успешно удален из команды",
	}

	return c.JSON(http.StatusOK, deleteResponse)
}

// AddProjectToTeam godoc
// @Summary Добавление проекта в команду
// @Description Привязывает проект к указанной команде
// @Tags Teams
// @Accept json
// @Produce json
// @Param addProject body request.TeamAddProjectRequest true "Данные для добавления проекта в команду"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TeamUniversalProjectResponse "Проект успешно привязан к команде"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 500 {object} map[string]string "Ошибка сервера при добавлении проекта в команду"
// @Router /team/project [post]
func AddProjectToTeam(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}
	var req request.TeamAddProjectRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	var team models.Team
	if err := DBConn.Session(&gorm.Session{}).Model(models.Team{}).
		Where("id = ? AND deleted = FALSE", req.TeamID).
		First(&team).Error; err != nil {
		log.Printf("DB error (select team): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении команды",
		})
	}

	var project models.Project
	if err := DBConn.Session(&gorm.Session{}).Model(models.Project{}).
		Where("id = ? AND deleted = FALSE", req.ProjectID).
		First(&project).Error; err != nil {
		log.Printf("DB error (select project): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проекта",
		})
	}

	del := false
	projectTeam := models.ProjectTeam{
		ProjectID: project.ID,
		TeamID:    team.ID,
		Deleted:   &del,
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Session(&gorm.Session{}).Model(models.ProjectTeam{}).Create(&projectTeam); res.Error != nil {
			log.Printf("DB error (add project to team): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (add project to team): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при привязке проекта к команде"})
	}

	addResponse := response.TeamUniversalProjectResponse{
		TeamID:    team.ID.String(),
		ProjectID: project.ID.String(),
		Message:   "Проект успешно привязан к команде",
	}
	return c.JSON(http.StatusOK, addResponse)
}

// DeleteProjectFromTeam godoc
// @Summary Удаление проекта из команды
// @Description Логически удаляет привязку проекта к команде (поле deleted = true)
// @Tags Teams
// @Accept json
// @Produce json
// @Param deleteProject body request.TeamDeleteProjectRequest true "Данные для удаления проекта из команды"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TeamUniversalProjectResponse "Проект успешно отвязан от команды"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 404 {object} map[string]string "Ничего не удалено"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении проекта из команды"
// @Router /team/project [delete]
func DeleteProjectFromTeam(c echo.Context) error {
	if err := Authorize(c); err != nil {
		return err
	}
	var req request.TeamDeleteProjectRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	teamID, err := uuid.Parse(req.TeamID)
	if err != nil {
		log.Printf("UUID parse error (teamID): %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор команды",
		})
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		log.Printf("UUID parse error (projectID): %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор проекта",
		})
	}

	updateData := map[string]interface{}{
		"deleted":    true,
		"updated_at": time.Now(),
	}

	if txErr := DBConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.ProjectTeam{}).
			Where("team_id = ? AND project_id = ?", teamID, projectID).
			Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete project from team): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete project from team): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении проекта из команды"})
	}

	deleteResponse := response.TeamUniversalProjectResponse{
		TeamID:    teamID.String(),
		ProjectID: projectID.String(),
		Message:   "Проект успешно отвязан от команды",
	}

	return c.JSON(http.StatusOK, deleteResponse)
}
