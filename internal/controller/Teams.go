package controller

import (
	"emplacc-api/internal/db"
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

func RegisterTeamRoutes(e *echo.Echo) {
	teamGroup := e.Group("/team")
	teamGroup.GET("/all", GetTeams)
	teamGroup.GET("/:id", GetTeamByID)
	teamGroup.POST("", CreateTeam)
	teamGroup.PATCH("/:id", UpdateTeam)
	teamGroup.DELETE("/:id", DeleteTeam) // Только для admin, добавить проверку роли
	teamGroup.POST("/user", AddUserToTeam)
	teamGroup.DELETE("/user", DeleteUserFromTeam)
	teamGroup.POST("/project", AddProjectToTeam)
	teamGroup.DELETE("/project", DeleteProjectFromTeam)
}

var dbConn *gorm.DB = db.DB_conn

// GetTeams godoc
// @Summary Получение списка всех команд
// @Description Получает список всех команд с их участниками
// @Tags Teams
// @Accept json
// @Produce json
// @Success 200 {object} response.TeamsListResponse "Список команд успешно получен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении команд"
// @Router /team/all [get]
func GetTeams(c echo.Context) error {
	var teams []models.Team
	result := dbConn.Session(&gorm.Session{}).Find(&teams)
	if result.Error != nil {
		log.Printf("DB error (find teams): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении списка команд",
		})
	}

	var teamMembers []models.TeamMember
	resultTM := dbConn.Session(&gorm.Session{}).Find(&teamMembers).Where("deleted = ?", false)
	if resultTM.Error != nil {
		log.Printf("DB error (find team members): %v", resultTM.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении участников команд",
		})
	}

	teamAndMembers := make(map[uuid.UUID][]response.TeamMemberResponse, len(teams))
	teamInfo := make(map[uuid.UUID]response.TeamResponse, len(teams))

	for _, team := range teams {
		var name string
		if team.Name != nil {
			name = *team.Name
		}

		var description string
		if team.Description != nil {
			description = *team.Description
		}

		var updatedAt time.Time
		if team.UpdatedAt != nil {
			updatedAt = *team.UpdatedAt
		}

		teamInfo[team.ID] = response.TeamResponse{
			ID:          team.ID.String(),
			Name:        name,
			Description: description,
			UpdatedAt:   updatedAt,
		}
	}

	for _, tm := range teamMembers {
		var member models.User
		userId := tm.UserID.String()
		err := dbConn.Session(&gorm.Session{}).Select("id, profession, first_name, last_name, email").
			Where("id = ? AND deleted = ?", userId, false).
			Find(&member).Error
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		profession := ""
		if member.Profession != nil {
			profession = *member.Profession
		}

		var memberFirstName string
		if member.FirstName != nil {
			memberFirstName = *member.FirstName
		}

		var memberLastName string
		if member.LastName != nil {
			memberLastName = *member.LastName
		}

		var memberEmail string
		if member.Email != nil {
			memberEmail = *member.Email
		}

		teamAndMembers[tm.TeamID] = append(teamAndMembers[tm.TeamID],
			response.TeamMemberResponse{
				UserID:         member.ID.String(),
				Specialization: profession,
				FirstName:      memberFirstName,
				LastName:       memberLastName,
				Email:          memberEmail,
			})
	}

	teamListResponse := response.TeamsListResponse{Teams: make([]response.TeamResponse, 0)}
	for k, v := range teamAndMembers {
		teamListResponse.Teams = append(teamListResponse.Teams, response.TeamResponse{
			ID:          k.String(),
			Name:        teamInfo[k].Name,
			Description: teamInfo[k].Description,
			Members:     v,
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
// @Success 200 {object} response.TeamResponse "Команда успешно получена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор команды"
// @Failure 404 {object} map[string]string "Команда не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении команды"
// @Router /team/{id} [get]
func GetTeamByID(c echo.Context) error {
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
			return c.JSON(http.StatusInternalServerError, map[string]string{
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
		member := memberMap[tm.UserID]
		profession := ""
		if member.Profession != nil {
			profession = *member.Profession
		}

		var memberFirstName string
		if member.FirstName != nil {
			memberFirstName = *member.FirstName
		}

		var memberLastName string
		if member.LastName != nil {
			memberLastName = *member.LastName
		}

		var memberEmail string
		if member.Email != nil {
			memberEmail = *member.Email
		}

		membersResp = append(membersResp, response.TeamMemberResponse{
			UserID:         member.ID.String(),
			Specialization: profession,
			FirstName:      memberFirstName,
			LastName:       memberLastName,
			Email:          memberEmail,
		})
	}

	var name string
	if team.Name != nil {
		name = *team.Name
	}

	var description string
	if team.Description != nil {
		description = *team.Description
	}

	var updatedAt time.Time
	if team.UpdatedAt != nil {
		updatedAt = *team.UpdatedAt
	}

	// Формируем итоговый ответ
	teamResponse := response.TeamResponse{
		ID:          team.ID.String(),
		Name:        name,
		Description: description,
		Members:     membersResp,
		UpdatedAt:   updatedAt,
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
// @Success 201 {object} response.TeamUniversalResponse "Команда успешно создана"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании команды"
// @Router /team [post]
func CreateTeam(c echo.Context) error {
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

	team := models.Team{ID: newUUID, Name: name, Description: description, Deleted: &del}
	result := dbConn.Session(&gorm.Session{}).Create(&team)
	if result.Error != nil {
		log.Printf("DB error (create team): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании команды",
		})
	}

	createRespose := response.TeamUniversalResponse{
		ID:      newUUID.String(),
		Message: "Команда создана",
	}

	return c.JSON(http.StatusCreated, createRespose)
}

// UpdateTeam godoc
// @Summary Обновление команды
// @Description Обновляет данные команды по её ID
// @Tags Teams
// @Accept json
// @Produce json
// @Param id path string true "ID команды"
// @Param team body request.TeamUpdateRequest true "Данные для обновления команды"
// @Success 200 {object} response.TeamUniversalResponse "Команда успешно обновлена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении команды"
// @Router /team/{id} [patch]
func UpdateTeam(c echo.Context) error {
	teamIDParam := c.Param("id")
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
	if err := dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", teamIDParam, false).First(&team).Error; err != nil {
		log.Printf("Team not found: %v", err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Команда не найдена",
		})
	}

	// Выполняем обновление только указанных полей
	if err := dbConn.Session(&gorm.Session{}).Model(&models.Team{}).Where("id = ?", teamIDParam).Updates(updateData).Error; err != nil {
		log.Printf("DB error (update team): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении команды",
		})
	}

	// Формируем ответ
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
// @Success 200 {object} response.TeamUniversalResponse "Команда успешно удалена"
// @Failure 404 {object} map[string]string "Команда не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении команды"
// @Router /team/{id} [delete]
func DeleteTeam(c echo.Context) error {
	teamIDParam := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	result := dbConn.Session(&gorm.Session{}).Model(models.Team{}).Where("id = ?", teamIDParam).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete team): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении команды",
		})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не удалено",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.TeamMember{}).Where("team_id = ?", teamIDParam).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete team_member connections): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении связей пользователь-команда",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.ProjectTeam{}).Where("team_id = ?", teamIDParam).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete project_team connections): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении связей проект-команда",
		})
	}

	deleteResponce := response.TeamUniversalResponse{
		ID:      teamIDParam,
		Message: "Команда удалена",
	}

	return c.JSON(http.StatusOK, deleteResponce)
}

// AddUserToTeam godoc
// @Summary Добавление пользователя в команду
// @Description Добавляет пользователя в указанную команду
// @Tags Teams
// @Accept json
// @Produce json
// @Param addUser body request.TeamAddUserRequest true "Данные для добавления пользователя в команду"
// @Success 200 {object} response.TeamUniversalUserResponse "Пользователь успешно добавлен в команду"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 500 {object} map[string]string "Ошибка сервера при добавлении пользователя в команду"
// @Router /team/user [post]
func AddUserToTeam(c echo.Context) error {
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
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении данных пользователя",
		})
	}

	var team models.Team
	if err := dbConn.Session(&gorm.Session{}).Where("id = ? AND deleted = ?", req.TeamID, false).First(&team).Error; err != nil {
		log.Printf("DB error (select team): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
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
	if err := dbConn.Session(&gorm.Session{}).Create(&teamMember).Error; err != nil {
		log.Printf("DB error (create team member): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Не удалось добавить пользователя в команду",
		})
	}

	addResponce := response.TeamUniversalUserResponse{
		TeamID:  team.ID.String(),
		UserID:  req.UserID,
		Message: "Пользователь добавлен в команду",
	}

	return c.JSON(http.StatusOK, addResponce)
}

// DeleteUserFromTeam godoc
// @Summary Удаление пользователя из команды
// @Description Логически удаляет пользователя из команды (поле deleted = true)
// @Tags Teams
// @Accept json
// @Produce json
// @Param deleteUser body request.TeamDeleteUserRequest true "Данные для удаления пользователя из команды"
// @Success 200 {object} response.TeamUniversalUserResponse "Пользователь успешно удален из команды"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 404 {object} map[string]string "Ничего не удалено"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении пользователя из команды"
// @Router /team/user [delete]
func DeleteUserFromTeam(c echo.Context) error {
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
	result := dbConn.Session(&gorm.Session{}).Model(models.TeamMember{}).Where("user_id = ? AND team_id = ?", user.ID, team.ID).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete team member): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении пользователя из команды",
		})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не удалено",
		})
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
// @Success 200 {object} response.TeamUniversalProjectResponse "Проект успешно привязан к команде"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 500 {object} map[string]string "Ошибка сервера при добавлении проекта в команду"
// @Router /team/project [post]
func AddProjectToTeam(c echo.Context) error {
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

	if err := dbConn.Session(&gorm.Session{}).Create(&projectTeam).Error; err != nil {
		log.Printf("DB error (add project to team): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при привязке проекта к команде",
		})
	}

	addResponce := response.TeamUniversalProjectResponse{
		TeamID:    team.ID.String(),
		ProjectID: project.ID.String(),
		Message:   "Проект успешно привязан к команде",
	}
	return c.JSON(http.StatusOK, addResponce)
}

// DeleteProjectFromTeam godoc
// @Summary Удаление проекта из команды
// @Description Логически удаляет привязку проекта к команде (поле deleted = true)
// @Tags Teams
// @Accept json
// @Produce json
// @Param deleteProject body request.TeamDeleteProjectRequest true "Данные для удаления проекта из команды"
// @Success 200 {object} response.TeamUniversalProjectResponse "Проект успешно отвязан от команды"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 404 {object} map[string]string "Ничего не удалено"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении проекта из команды"
// @Router /team/project [delete]
func DeleteProjectFromTeam(c echo.Context) error {
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

	result := dbConn.Session(&gorm.Session{}).Model(models.ProjectTeam{}).Where("team_id = ? AND project_id = ?", teamID, projectID).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete project from team): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении проекта из команды",
		})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusNotFound, map[string]string{
			"message": "Ничего не удалено",
		})
	}

	deleteResponce := response.TeamUniversalProjectResponse{
		TeamID:    teamID.String(),
		ProjectID: projectID.String(),
		Message:   "Проект успешно отвязан от команды",
	}

	return c.JSON(http.StatusOK, deleteResponce)
}
