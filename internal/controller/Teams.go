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
)

func RegisterTeamRoutes(e *echo.Echo) {
	teamGroup := e.Group("/team")
	teamGroup.Use(KeycloakAuthMiddleware)
	teamGroup.GET("/all", getTeams)
	teamGroup.GET("/:id", getTeamByID)
	teamGroup.POST("", createTeam)
	teamGroup.PATCH("/:id", updateTeam)
	teamGroup.DELETE("/:id", deleteTeam) // Только для admin, добавить проверку роли
	teamGroup.POST("/user", addUserToTeam)
	teamGroup.DELETE("/user", deleteUserFromTeam)
	teamGroup.POST("/project", addProjectToTeam)
	teamGroup.DELETE("/project", deleteProjectFromTeam)
	teamGroup.GET("/project/:project_id", getProjectTeams)
}

var dbConn *gorm.DB = db.DB_conn

// getTeams godoc
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
func getTeams(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	// Загружаем команды вместе с участниками и пользователями
	var teams []models.Team
	if err := dbConn.Session(&gorm.Session{}).
		Preload("TeamMembers.User").
		Where("deleted = ?", false).
		Find(&teams).Error; err != nil {
		log.Printf("DB error (find teams with preload): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении списка команд",
		})
	}

	teamListResponse := response.TeamsListResponse{Teams: make([]response.TeamResponse, 0, len(teams))}
	for _, team := range teams {
		members := make([]response.TeamMemberResponse, 0, len(team.TeamMembers))
		for _, tm := range team.TeamMembers {
			if tm.Deleted != nil && *tm.Deleted {
				continue
			}
			if tm.User == nil || (tm.User.Deleted != nil && *tm.User.Deleted) {
				continue
			}

			user := tm.User

			members = append(members, response.TeamMemberResponse{
				UserID:         user.ID.String(),
				Specialization: utils.GetString(user.Profession),
				FirstName:      utils.GetString(user.FirstName),
				LastName:       utils.GetString(user.LastName),
				Email:          utils.GetString(user.Email),
			})
		}

		teamListResponse.Teams = append(teamListResponse.Teams, response.TeamResponse{
			ID:          team.ID.String(),
			Name:        utils.GetString(team.Name),
			Description: utils.GetString(team.Description),
			UpdatedAt:   utils.GetTime(team.UpdatedAt),
			CreatedAt: utils.GetTime(team.CreatedAt),
			Members:     members,
		})
	}

	return c.JSON(http.StatusOK, teamListResponse)
}

// getProjectTeams godoc
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
func getProjectTeams(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}

	projectIDStr := c.Param("project_id")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный project_id",
		})
	}

	// Получаем связи проект-команда
	var projectTeams []models.ProjectTeam
	if err := dbConn.
		Session(&gorm.Session{}).
		Where("project_id = ? AND deleted = ?", projectID, false).
		Find(&projectTeams).Error; err != nil {
		log.Printf("DB error (find projectTeams): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении связей проект-команда",
		})
	}

	if len(projectTeams) == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Для данного проекта команды не найдены",
		})
	}

	// Собираем teamIDs
	teamIDs := make([]uuid.UUID, 0, len(projectTeams))
	for _, pt := range projectTeams {
		teamIDs = append(teamIDs, pt.TeamID)
	}

	// Получаем команды
	var teams []models.Team
	if err := dbConn.
		Session(&gorm.Session{}).
		Where("id IN ? AND deleted = ?", teamIDs, false).
		Find(&teams).Error; err != nil {
		log.Printf("DB error (find teams): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении списка команд",
		})
	}

	// Получаем участников команд
	var teamMembers []models.TeamMember
	if err := dbConn.
		Session(&gorm.Session{}).
		Where("team_id IN ? AND deleted = ?", teamIDs, false).
		Find(&teamMembers).Error; err != nil {
		log.Printf("DB error (find team members): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении участников команд",
		})
	}

	// Собираем userIDs
	userIDs := make([]uuid.UUID, 0, len(teamMembers))
	for _, tm := range teamMembers {
		userIDs = append(userIDs, tm.UserID)
	}

	// Загружаем пользователей
	var users []models.User
	if len(userIDs) > 0 {
		if err := dbConn.
			Session(&gorm.Session{}).
			Select("id, profession, first_name, last_name, email").
			Where("id IN ? AND deleted = ?", userIDs, false).
			Find(&users).Error; err != nil {
			log.Printf("DB error (find users): %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при получении пользователей",
			})
		}
	}

	userMap := make(map[uuid.UUID]models.User, len(users))
	for _, u := range users {
		userMap[u.ID] = u
	}

	// Формируем ответ
	teamInfo := make(map[uuid.UUID]response.TeamResponse, len(teams))
	for _, team := range teams {
		teamInfo[team.ID] = response.TeamResponse{
			ID:          team.ID.String(),
			Name:        utils.GetString(team.Name),
			Description: utils.GetString(team.Description),
			UpdatedAt:   utils.GetTime(team.UpdatedAt),
			CreatedAt: utils.GetTime(team.CreatedAt),
		}
	}

	teamAndMembers := make(map[uuid.UUID][]response.TeamMemberResponse, len(teams))
	for _, tm := range teamMembers {
		if user, ok := userMap[tm.UserID]; ok {
			teamAndMembers[tm.TeamID] = append(teamAndMembers[tm.TeamID],
				response.TeamMemberResponse{
					UserID:         user.ID.String(),
					Specialization: utils.GetString(user.Profession),
					FirstName:      utils.GetString(user.FirstName),
					LastName:       utils.GetString(user.LastName),
					Email:          utils.GetString(user.Email),
				})
		}
	}

	// Итоговый список
	teamListResponse := response.TeamsListResponse{Teams: make([]response.TeamResponse, 0, len(teams))}
	for id, info := range teamInfo {
		info.Members = teamAndMembers[id]
		teamListResponse.Teams = append(teamListResponse.Teams, info)
	}

	return c.JSON(http.StatusOK, teamListResponse)
}



// getTeamByID godoc
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
func getTeamByID(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	teamIDParam := c.Param("id")
	teamUUID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор команды"})
	}

	var team models.Team
	result := dbConn.Session(&gorm.Session{}).First(&team, "id = ?", teamUUID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Команда не найдена"})
		}
		log.Printf("DB error (get team): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении данных команды",
		})
	}

	if team.Deleted != nil {
		if *team.Deleted {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Команда не найдена",
			})
		}
	}

	var teamMembers []models.TeamMember
	resultTM := dbConn.Session(&gorm.Session{}).Where("team_id = ? AND deleted = ?", teamUUID, false).Find(&teamMembers)
	if resultTM.Error != nil {
		log.Printf("DB error (get team members): %v", resultTM.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении участников команды",
		})
	}

	// Получаем пользователей одним запросом
	userIDs := make([]uuid.UUID, 0, len(teamMembers))
	for _, tm := range teamMembers {
		userIDs = append(userIDs, tm.UserID)
	}

	var members []models.User
	if len(userIDs) > 0 {
		err = dbConn.Session(&gorm.Session{}).Select("id, profession, first_name, last_name, email").
			Where("id IN (?) AND deleted = ?", userIDs, false).
			Find(&members).Error
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	// Преобразуем в map для быстрого доступа
	memberMap := make(map[uuid.UUID]models.User, len(members))
	for _, m := range members {
		memberMap[m.ID] = m
	}

	// Формируем список участников ответа
	membersResp := make([]response.TeamMemberResponse, 0, len(teamMembers))
	for _, tm := range teamMembers {
		user := memberMap[tm.UserID]
		membersResp = append(membersResp, response.TeamMemberResponse{
			UserID:         user.ID.String(),
			Specialization: utils.GetString(user.Profession),
			FirstName:      utils.GetString(user.FirstName),
			LastName:       utils.GetString(user.LastName),
			Email:          utils.GetString(user.Email),
		})
	}

	// Формируем итоговый ответ
	teamResponse := response.TeamResponse{
		ID:          team.ID.String(),
		Name:        utils.GetString(team.Name),
		Description: utils.GetString(team.Description),
		UpdatedAt:   utils.GetTime(team.UpdatedAt),
		CreatedAt: utils.GetTime(team.CreatedAt),
		Members: membersResp,
	}

	return c.JSON(http.StatusOK, teamResponse)
}

// createTeam godoc
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
func createTeam(c echo.Context) error {
	if err := authorize(c); err != nil {
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
		ID: newUUID, 
		Name: name, 
		Description: description, 
		Deleted: &del, CreatedAt: &now,
	}
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Create(&team); res.Error != nil {
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

// updateTeam godoc
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
func updateTeam(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	teamIDParam := c.Param("id")
	teamUUID, err := uuid.Parse(teamIDParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор команды"})
	}
	// Привязка данных запроса
	var req request.TeamUpdateRequest
	if err := c.Bind(&req); err != nil {
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

	// Проверяем, есть ли поля для обновления
	if len(updateData) == 0 {
		log.Printf("No fields provided for update")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	// Проверка существования команды
	var team models.Team
	if err := dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", teamUUID, false).First(&team).Error; err != nil {
		log.Printf("Team not found: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Команда не найдена",
		})
	}

	// Выполняем обновление только указанных полей
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Team{}).Where("id = ?", teamUUID).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (update team): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0{
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update team): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении команды"})
	}

	// Формируем ответ
	updateResponse := response.TeamUniversalResponse{
		ID:      teamIDParam,
		Message: "Команда обновлена",
	}

	return c.JSON(http.StatusOK, updateResponse)
}

// deleteTeam godoc
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
func deleteTeam(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	teamIDParam := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(models.Team{}).Where("id = ?", teamIDParam).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete team): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}

		if res = tx.Model(models.TeamMember{}).Where("team_id = ?", teamIDParam).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete team_member connections): %v", res.Error)
			return res.Error
		}

		if res = tx.Model(models.ProjectTeam{}).Where("team_id = ?", teamIDParam).Updates(updateData); res.Error != nil {
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

// addUserToTeam godoc
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
func addUserToTeam(c echo.Context) error {
	if err := authorize(c); err != nil {
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
	if err := dbConn.Session(&gorm.Session{}).Select("id, profession").Where("id = ? AND deleted = ?", req.UserID, false).First(&user).Error; err != nil {
		log.Printf("DB error (select profession): %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при получении данных пользователя",
		})
	}

	var team models.Team
	if err := dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", req.TeamID, false).First(&team).Error; err != nil {
		log.Printf("DB error (select team): %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при получении команды",
		})
	}

	del := false

	teamMember := models.TeamMember{
		UserID:         user.ID,
		TeamID:         team.ID,
		Specialization: user.Profession,
		Deleted: &del,
	}
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Create(&teamMember); res.Error != nil {
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

// deleteUserFromTeam godoc
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
func deleteUserFromTeam(c echo.Context) error {
	if err := authorize(c); err != nil {
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
	if err := dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", req.TeamID, false).First(&team).Error; err != nil {
		log.Printf("DB error (select team): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении команды",
		})
	}

	var user models.User
	if err := dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", req.UserID, false).First(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении пользователя",
		})
	}

	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(models.TeamMember{}).Where("user_id = ? AND team_id = ?", user.ID, team.ID).Updates(updateData)
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

// addProjectToTeam godoc
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
func addProjectToTeam(c echo.Context) error {
	if err := authorize(c); err != nil {
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
	if err := dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", req.TeamID, false).First(&team).Error; err != nil {
		log.Printf("DB error (select team): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении команды",
		})
	}

	var project models.Project
	if err := dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", req.ProjectID, false).First(&project).Error; err != nil {
		log.Printf("DB error (select project): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проекта",
		})
	}

	del := false

	projectTeam := models.ProjectTeam{
		ProjectID: project.ID,
		TeamID:    team.ID,
		Deleted: &del,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Create(&projectTeam); res.Error != nil {
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

// deleteProjectFromTeam godoc
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
func deleteProjectFromTeam(c echo.Context) error {
	if err := authorize(c); err != nil {
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

	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	updateData["updated_at"] = time.Now()

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(models.ProjectTeam{}).Where("team_id = ? AND project_id = ?", teamID, projectID).Updates(updateData)
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
