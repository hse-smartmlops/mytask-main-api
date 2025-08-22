package controller

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"
)

func RegisterUserRoutes(e *echo.Echo) {
	userGroup := e.Group("/user")
	userGroup.GET("", GetAllUsers)
	userGroup.GET("/:id", GetUserById)
	userGroup.POST("", CreateUser)
	userGroup.POST("/:id", UpdateUser)
	userGroup.DELETE("/:id", DeleteUser)
	userGroup.POST("/role", AddUserRole)
	userGroup.DELETE("/role", RemoveUserRole)
}

// GetAllUsers godoc
// @Summary Получение списка всех пользователей
// @Description Получает список всех пользователей с учетом пагинации, исключая удаленных
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Номер страницы" default(1)
// @Param pageSize query int false "Размер страницы" default(10)
// @Success 200 {object} response.GetAllUsersResponse "Список пользователей успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении пользователей"
// @Router /user [get]
func GetAllUsers(c echo.Context) error {
	var req request.GetAllUsersRequest
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
	result := dbConn.Session(&gorm.Session{}).Model(models.User{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count users): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете пользователей",
		})
	}

	var users []models.User
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		log.Printf("DB error (find projects): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проектов из базы данных",
		})
	}

	userList := response.GetAllUsersResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, user := range users {
		var userId string
		userId = user.ID.String()
		var email string
		if user.Email != nil {
			email = *user.Email
		}
		var isActive bool
		if user.IsActive != nil {
			isActive = *user.IsActive
		}
		var createdAt time.Time
		if user.CreatedAt != nil {
			createdAt = *user.CreatedAt
		}
		var tgId string
		if user.TgID != nil {
			tgId = *user.TgID
		}
		var tgUserId int64
		if user.TgUserID != nil {
			tgUserId = *user.TgUserID
		}
		var profession string
		if user.Profession != nil {
			profession = *user.Profession
		}
		var emailVerified bool
		if user.EmailVerified != nil {
			emailVerified = *user.EmailVerified
		}
		var firstName string
		if user.FirstName != nil {
			firstName = *user.FirstName
		}
		var lastName string
		if user.LastName != nil {
			lastName = *user.LastName
		}
		var lastLogin time.Time
		if user.LastLogin != nil {
			lastLogin = *user.LastLogin
		}
		var authProviderId string
		if user.AuthProviderID != nil {
			authProviderId = user.AuthProviderID.String()
		}
		userList.Users = append(userList.Users, response.GetUserResponse{
			ID:             userId,
			Email:          email,
			IsActive:       isActive,
			CreatedAt:      createdAt,
			TgId:           tgId,
			TgUserId:       tgUserId,
			Profession:     profession,
			EmailVerified:  emailVerified,
			FirstName:      firstName,
			LastName:       lastName,
			LastLogin:      lastLogin,
			AuthProviderId: authProviderId,
		})
	}
	return c.JSON(http.StatusOK, userList)
}

// GetUserById godoc
// @Summary Получение пользователя по ID
// @Description Получает данные пользователя по его уникальному идентификатору
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {object} response.GetUserResponse "Пользователь успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор пользователя"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении пользователя"
// @Router /user/{id} [get]
func GetUserById(c echo.Context) error {
	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор пользователя",
		})
	}

	var user models.User
	result := dbConn.Session(&gorm.Session{}).Where("id = ?", userId).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Пользователь не найден",
			})
		}
		log.Printf("DB error (find user by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении пользователя из базы данных",
		})
	}
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	var isActive bool
	if user.IsActive != nil {
		isActive = *user.IsActive
	}
	var createdAt time.Time
	if user.CreatedAt != nil {
		createdAt = *user.CreatedAt
	}
	var tgId string
	if user.TgID != nil {
		tgId = *user.TgID
	}
	var tgUserId int64
	if user.TgUserID != nil {
		tgUserId = *user.TgUserID
	}
	var profession string
	if user.Profession != nil {
		profession = *user.Profession
	}
	var emailVerified bool
	if user.EmailVerified != nil {
		emailVerified = *user.EmailVerified
	}
	var firstName string
	if user.FirstName != nil {
		firstName = *user.FirstName
	}
	var lastName string
	if user.LastName != nil {
		lastName = *user.LastName
	}
	var lastLogin time.Time
	if user.LastLogin != nil {
		lastLogin = *user.LastLogin
	}
	var authProviderId string
	if user.AuthProviderID != nil {
		authProviderId = user.AuthProviderID.String()
	}
	getUserResponse := response.GetUserResponse{
		ID:             id,
		Email:          email,
		IsActive:       isActive,
		CreatedAt:      createdAt,
		TgId:           tgId,
		TgUserId:       tgUserId,
		Profession:     profession,
		EmailVerified:  emailVerified,
		FirstName:      firstName,
		LastName:       lastName,
		LastLogin:      lastLogin,
		AuthProviderId: authProviderId,
	}
	return c.JSON(http.StatusOK, getUserResponse)
}

// CreateUser godoc
// @Summary Создание нового пользователя
// @Description Создает нового пользователя с указанными параметрами
// @Tags Users
// @Accept json
// @Produce json
// @Param user body request.UserCreateRequest true "Данные для создания пользователя"
// @Success 201 {object} response.UserUniversalResponse "Пользователь успешно создан"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании пользователя"
// @Router /user [post]
func CreateUser(c echo.Context) error {
	var req request.UserCreateRequest
	err := c.Bind(&req)
	if err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	now := time.Now()

	authProviderId, err := uuid.Parse(*req.AuthProviderId)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор провайдера авторизации",
		})
	}

	del := false

	user := models.User{
		ID:             newUUID,
		Email:          req.Email,
		IsActive:       req.IsActive,
		CreatedAt:      &now,
		UpdatedAt:      &now,
		TgID:           req.TgId,
		TgUserID:       req.TgUserId,
		Profession:     req.Profession,
		EmailVerified:  req.EmailVerified,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		LastLogin:      req.LastLogin,
		AuthProviderID: &authProviderId,
		Deleted: &del,
	}

	result := dbConn.Session(&gorm.Session{}).Create(&user)
	if result.Error != nil {
		log.Printf("DB error (create user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании пользователя",
		})
	}

	createResponse := response.UserUniversalResponse{
		ID:      newUUID.String(),
		Message: "Пользователь успешно создан",
	}
	return c.JSON(http.StatusCreated, createResponse)
}

// UpdateUser godoc
// @Summary Обновление пользователя
// @Description Обновляет данные пользователя по его ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param user body request.UpdateUserRequest true "Данные для обновления пользователя"
// @Success 200 {object} response.UserUniversalResponse "Пользователь успешно обновлен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении пользователя"
// @Router /user/{id} [post]
func UpdateUser(c echo.Context) error {
	id := c.Param("id")
	userId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор пользователя",
		})
	}

	var req request.UpdateUserRequest
	if err = c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := make(map[string]interface{})
	if req.Email != nil {
		updateData["email"] = *req.Email
	}
	if req.IsActive != nil {
		updateData["is_active"] = *req.IsActive
	}
	if req.CreatedAt != nil {
		updateData["created_at"] = *req.CreatedAt
	}
	if req.TgId != nil {
		updateData["tg_id"] = *req.TgId
	}
	if req.TgUserId != nil {
		updateData["tg_user_id"] = *req.TgUserId
	}
	if req.Profession != nil {
		updateData["profession"] = *req.Profession
	}
	if req.EmailVerified != nil {
		updateData["email_verified"] = *req.EmailVerified
	}
	if req.FirstName != nil {
		updateData["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		updateData["last_name"] = *req.LastName
	}
	if req.LastLogin != nil {
		updateData["last_login"] = *req.LastLogin
	}
	if req.AuthProviderId != nil {
		updateData["auth_provider_id"] = *req.AuthProviderId
	}

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	if err = dbConn.Session(&gorm.Session{}).Model(models.Project{}).Where("id = ?", userId).Updates(updateData).Error; err != nil {
		log.Printf("DB error (update user): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обновлении пользователя",
		})
	}

	updateResponse := response.UserUniversalResponse{
		ID:      userId.String(),
		Message: "Пользователь с ID " + id + " обновлен",
	}
	return c.JSON(http.StatusOK, updateResponse)
}

// DeleteUser godoc
// @Summary Удаление пользователя
// @Description Логическое удаление пользователя по ID, включая связанные данные (поле deleted = true)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Success 200 {object} response.UserUniversalResponse "Пользователь успешно удален"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении пользователя"
// @Router /user/{id} [delete]
func DeleteUser(c echo.Context) error {
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	updateData["updated_at"] = time.Now()
	result := dbConn.Session(&gorm.Session{}).Model(&models.User{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении пользователя",
		})
	}
	if result.RowsAffected == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"message": "Ничего не удалено",
		})
	}

	var task models.Task
	result = dbConn.Session(&gorm.Session{}).Where("created_by = ? OR assigned_to = ?", id, id).First(&task)
	if result.Error != nil {
		log.Printf("DB error (find task): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при поиске пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.HelpRequest{}).Where("task_id = ?", task.ID).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete help request): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении запроса на помощь",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.Task{}).Where("created_by = ? OR assigned_to = ?", id, id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete task): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении задач данного пользователя",
		})
	}

	var dailyReport models.DailyReport
	result = dbConn.Session(&gorm.Session{}).Where("user_id = ? and deleted = ?", id, false).First(&dailyReport)
	if result.Error != nil {
		log.Printf("DB error (find daily report): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при поиске отчета",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(models.HelpRequest{}).Where("report_id = ?", dailyReport.ID).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete help request): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении запроса на помощь",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.DailyReport{}).Where("user_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete report): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении отчетов данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.UserRole{}).Where("assigned_by = ? OR user_id = ?", id, id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении связи пользователь-роль данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.TeamMember{}).Where("user_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении связи пользователь-команда данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.Attendance{}).Where("user_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении посещений данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.AuthEvent{}).Where("user_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении авторизаций данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.DailyReport{}).Where("user_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении отчетов данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.Session{}).Where("user_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении сессий данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.AuthProvider{}).Where("id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении аутентификационных провайдеров данного пользователя",
		})
	}

	var problem models.Problem
	result = dbConn.Session(&gorm.Session{}).Where("deleted = ? and creator_id = ?", false, id).First(&problem)
	if result.Error != nil {
		log.Printf("DB error (get problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении проблем данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.ForumMessage{}).Where("problem_id = ?", problem.ID).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (update forum): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении форумов данного пользователя",
		})
	}

	var reportProblems []models.ReportProblem
	result = dbConn.Session(&gorm.Session{}).Model(&models.ReportProblem{}).Where("problem_id = ?", problem.ID).Find(&reportProblems)
	if result.Error != nil {
		log.Printf("DB error (get report-problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении связей отчет-проблема данного пользователя",
		})
	}

	for _, reportProblem := range reportProblems{
		result = dbConn.Session(&gorm.Session{}).Model(&models.Problem{}).Where("id = ?", reportProblem.ProblemID).Updates(updateData)
		if result.Error != nil {
			log.Printf("DB error (delete user): %v", result.Error)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Ошибка при удалении проблем данного пользователя",
			})
		}
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.ReportProblem{}).Where("problem_id = ?", problem.ID).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete report-problem): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении связей отчет-проблема данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.Problem{}).Where("creator_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении проблем данного пользователя",
		})
	}

	result = dbConn.Session(&gorm.Session{}).Model(&models.ForumMessage{}).Where("creator_id = ?", id).Updates(updateData)
	if result.Error != nil {
		log.Printf("DB error (delete forum): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении форумов данного пользователя",
		})
	}

	deleteResponse := response.UserUniversalResponse{
		ID:      id,
		Message: "Пользователь с ID " + id + " удален",
	}

	return c.JSON(http.StatusOK, deleteResponse)
}

// AddUserRole godoc
// @Summary Добавление роли пользователю
// @Description Добавляет роль указанному пользователю
// @Tags Users
// @Accept json
// @Produce json
// @Param addRole body request.AddRoleUserRequest true "Данные для добавления роли пользователю"
// @Success 200 {object} response.AddRoleUserResponse "Роль успешно добавлена"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 500 {object} map[string]string "Ошибка сервера при добавлении роли"
// @Router /user/role [post]
func AddUserRole(c echo.Context) error {
	var req request.AddRoleUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	var user models.User
	result := dbConn.Session(&gorm.Session{}).Select("id, profession").Where("id = ? AND deleted = ?", req.UserId, false).First(&user)
	if result.Error != nil {
		log.Printf("DB error (select user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении данных пользователя",
		})
	}

	var role models.Role
	result = dbConn.Session(&gorm.Session{}).Select("id").Where("id = ? AND deleted = ?", req.RoleId, false).First(&role)
	if result.Error != nil {
		log.Printf("DB error (select role): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении данных роли",
		})
	}

	del := false

	userRole := models.UserRole{
		UserID: user.ID,
		RoleID: role.ID,
		Deleted: &del,
	}

	if err := dbConn.Session(&gorm.Session{}).Create(&userRole).Error; err != nil {
		log.Printf("DB error (create userRole): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Не удалось добавить роль пользователям",
		})
	}

	addResponse := response.AddRoleUserResponse{
		UserId:  user.ID.String(),
		RoleId:  role.ID.String(),
		Message: "Роль успешно добавлена",
	}

	return c.JSON(http.StatusOK, addResponse)
}

// RemoveUserRole godoc
// @Summary Удаление роли у пользователя
// @Description Удаляет роль у указанного пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Param removeRole body request.RemoveRoleUserRequest true "Данные для удаления роли у пользователя"
// @Success 200 {object} response.RemoveRoleUserResponse "Роль успешно удалена"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 404 {object} map[string]string "Ничего не удалено"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении роли"
// @Router /user/role [delete]
func RemoveUserRole(c echo.Context) error {
	var req request.RemoveRoleUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	var user models.User
	result := dbConn.Session(&gorm.Session{}).Select("id, profession").Where("id = ? AND deleted = ?", req.UserId, false).First(&user)
	if result.Error != nil {
		log.Printf("DB error (select user): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении данных пользователя",
		})
	}

	var role models.Role
	result = dbConn.Session(&gorm.Session{}).Select("id").Where("id = ? AND deleted = ?", req.RoleId, false).First(&role)
	if result.Error != nil {
		log.Printf("DB error (select role): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении данных роли",
		})
	}

	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	updateData["updated_at"] = time.Now()

	result = dbConn.Session(&gorm.Session{}).Model(models.UserRole{}).Where("user_id = ? AND role_id = ?", user.ID, role.ID).Updates(updateData)
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

	deleteResponse := response.RemoveRoleUserResponse{
		RoleId:  role.ID.String(),
		UserId:  user.ID.String(),
		Message: "Роль успешно удалена",
	}

	return c.JSON(http.StatusOK, deleteResponse)
}
