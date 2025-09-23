package controller

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"emplacc-api/internal/utils"
	"log"
	"net/http"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func RegisterUserRoutes(e *echo.Echo) {
	userGroup := e.Group("/user")
	userGroup.Use(KeycloakAuthMiddleware)
	userGroup.GET("/all/:page/:pagesize", getAllUsers)
	userGroup.GET("/:id", getUserById)
	userGroup.POST("", createUser)
	userGroup.POST("/:id", updateUser)
	userGroup.DELETE("/:id", banUser)
	userGroup.POST("/restore", restoreUser)
	userGroup.POST("/role", addUserRole)
	userGroup.DELETE("/role", removeUserRole)
	userGroup.DELETE("/full-delete/:id", deleteUser)
}

// getAllUsers godoc
// @Summary Получение списка всех пользователей
// @Description Получает список всех пользователей с учетом пагинации, исключая удаленных
// @Tags Users
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.GetAllUsersResponse "Список пользователей успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении пользователей"
// @Router /user/all/{page}/{pagesize} [get]
func getAllUsers(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	pageReq := c.Param("page")
	pageSizeReq := c.Param("pagesize")

	page, err := strconv.Atoi(pageReq)
	if err != nil {
		log.Printf("failed to parse page: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Ошибка при парсинге страницы",
		})
	}
	if page <= 0 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeReq)
	if err != nil {
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
	result := dbConn.Session(&gorm.Session{}).Model(models.User{}).Where("deleted = FALSE").Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count users): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете пользователей",
		})
	}

	var users []models.User
	if err := dbConn.Session(&gorm.Session{}).Model(models.User{}).
		Where("deleted = FALSE").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error; err != nil {
		log.Printf("DB error (find users): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении пользователей из базы данных",
		})
	}

	userList := response.GetAllUsersResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
	}

	for _, user := range users {
		userList.Users = append(userList.Users, response.GetUserResponse{
			ID:            user.ID.String(),
			Email:         user.Email,
			IsActive:      user.IsActive,
			CreatedAt:     user.CreatedAt,
			TgId:          user.TgID,
			TgUserId:      user.TgUserID,
			Profession:    user.Profession,
			EmailVerified: user.EmailVerified,
			FirstName:     user.FirstName,
			LastName:      user.LastName,
			LastLogin:     user.LastLogin,
		})
	}
	return c.JSON(http.StatusOK, userList)
}

// getUserById godoc
// @Summary Получение пользователя по ID
// @Description Получает данные пользователя по его уникальному идентификатору
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.GetUserResponse "Пользователь успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор пользователя"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении пользователя"
// @Router /user/{id} [get]
func getUserById(c echo.Context) error {
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

	var user models.User
	result := dbConn.Session(&gorm.Session{}).Model(models.User{}).
		Where("id = ? AND deleted = FALSE", userId).
		First(&user)
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
	getUserResponse := response.GetUserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		IsActive:      user.IsActive,
		CreatedAt:     user.CreatedAt,
		TgId:          user.TgID,
		TgUserId:      user.TgUserID,
		Profession:    user.Profession,
		EmailVerified: user.EmailVerified,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		LastLogin:     user.LastLogin,
	}
	return c.JSON(http.StatusOK, getUserResponse)
}

// createUser godoc
// @Summary Создание нового пользователя
// @Description Создаёт пользователя с автоматически сгенерированным UUID
// @Tags Users
// @Accept json
// @Produce json
// @Param body body request.UserCreateRequest true "Данные для создания пользователя"
// @Security BearerAuth
// @Success 201 {object} map[string]string "Пользователь успешно создан"
// @Failure 400 {object} map[string]string "Ошибка при привязке данных"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка при создании пользователя"
// @Router /user/create [post]
func createUser(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err // authorize возвращает echo.HTTPError — это корректно
	}

	var req request.UserCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	if err := CreateUserWithIdFunc(req, c, newUUID); err != nil {
		log.Printf("Failed to create user: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании пользователя",
		})
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Пользователь успешно создан",
		"id":      newUUID.String(),
	})
}

/*
func CreateUserFunc(req request.UserCreateRequest, c echo.Context) error {
	newUUID := uuid.New()
	now := time.Now()
	del := false

	user := models.User{
		ID:            newUUID,
		Email:         req.Email,
		IsActive:      req.IsActive,
		CreatedAt:     &now,
		UpdatedAt:     &now,
		TgID:          req.TgId,
		TgUserID:      req.TgUserId,
		Profession:    req.Profession,
		EmailVerified: req.EmailVerified,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		LastLogin:     &now,
		Deleted:       &del,
	}

	if err := dbConn.Transaction(func(tx *gorm.DB) error {
		return tx.
			Omit(clause.Associations).
			Create(&user).Error
	}); err != nil {
		log.Printf("DB transaction error (create user): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании пользователя",
		})
	}

	return c.JSON(http.StatusCreated, response.UserUniversalResponse{
		ID:      newUUID.String(),
		Message: "Пользователь успешно создан",
	})
}*/

/*
func CreateUserWithIdFunc(req request.UserCreateRequest, c echo.Context, id uuid.UUID) error {
	now := time.Now()
	
	// Создаем пользователя без ассоциаций
	user := models.User{
		ID:            id,
		Email:         utils.GetString(req.Email),
		IsActive:      utils.GetBool(req.IsActive),
		CreatedAt:     now,
		UpdatedAt:     now,
		EmailVerified: utils.GetBool(req.EmailVerified),
		FirstName:     utils.GetString(req.FirstName),
		LastName:      utils.GetString(req.LastName),
		LastLogin:     now,
		Deleted:       false,
		TgID:          "",
		TgUserID:      0,
		Profession:    "",
		// UserRoles остается пустым слайсом по умолчанию
	}

	log.Printf("CreateUserWithIdFunc: start create user id=%s email=%v", user.ID, user.Email)

	if err := dbConn.Transaction(func(tx *gorm.DB) error {
		log.Printf("CreateUserWithIdFunc: db transaction started for id=%s", user.ID)
		
		// Явно указываем поля для создания, исключая ассоциации
		res := tx.Omit("UserRoles").Select(
			"ID", "Email", "IsActive", "CreatedAt", "UpdatedAt",
			"TgID", "TgUserID", "Profession", "EmailVerified",
			"FirstName", "LastName", "LastLogin", "Deleted",
		).Create(&user)

		if res.Error != nil {
			log.Printf("CreateUserWithIdFunc: db create error for id=%s: %v", user.ID, res.Error)
			return res.Error
		}

		log.Printf("CreateUserWithIdFunc: user created id=%s rows=%d", user.ID, res.RowsAffected)
		return nil
	}); err != nil {
		log.Printf("DB transaction error (create user with id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании пользователя",
		})
	}

	log.Printf("CreateUserWithIdFunc: finished create user id=%s", user.ID)
	return c.JSON(http.StatusCreated, map[string]string{
		"message": "Пользователь успешно создан",
		"id":      user.ID.String(),
	})
}*/

func CreateUserWithIdFunc(req request.UserCreateRequest, c echo.Context, id uuid.UUID) error {
	now := time.Now()
	
	email := utils.GetString(req.Email)
	isActive := utils.GetBool(req.IsActive)
	emailVerified := utils.GetBool(req.EmailVerified)
	firstName := utils.GetString(req.FirstName)
	lastName := utils.GetString(req.LastName)
	
	// Используем Raw SQL для полного контроля
	query := `INSERT INTO users (id, email, is_active, created_at, updated_at, email_verified, first_name, last_name, last_login, deleted, tg_id, tg_user_id, profession) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	
	result := dbConn.Exec(query, 
		id, 
		email, 
		isActive, 
		now, 
		now, 
		emailVerified, 
		firstName, 
		lastName, 
		now, 
		false, 
		"", 
		int64(0), 
		"")
	
	if result.Error != nil {
		log.Printf("CreateUserWithIdFunc: db create error for id=%s: %v", id, result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при создании пользователя",
		})
	}
	
	affectedRows := result.RowsAffected
	log.Printf("CreateUserWithIdFunc: user created id=%s rows=%d", id, affectedRows)
	
	return nil
}

// updateUser godoc
// @Summary Обновление пользователя
// @Description Обновляет данные пользователя по его ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param user body request.UpdateUserRequest true "Данные для обновления пользователя"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.UserUniversalResponse "Пользователь успешно обновлен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении пользователя"
// @Router /user/{id} [post]
func updateUser(c echo.Context) error {
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

	if len(updateData) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}

	updateData["updated_at"] = time.Now()

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(models.User{}).
			Where("id = ? AND deleted = FALSE", userId).
			Updates(updateData)
		if res.Error != nil {
			log.Print("DB error (update user)")
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update user): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении пользователя"})
	}

	updateResponse := response.UserUniversalResponse{
		ID:      userId.String(),
		Message: "Пользователь с ID " + id + " обновлен",
	}
	return c.JSON(http.StatusOK, updateResponse)
}

/*
// deleteUser godoc
// @Summary Удаление пользователя
// @Description Логическое удаление пользователя по ID, включая связанные данные (поле deleted = true)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.UserUniversalResponse "Пользователь успешно удален"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении пользователя"
// @Router /user/{id} [delete]
func deleteUser(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	return DeleteUserFunc(c, id)
}

func DeleteUserFunc(c echo.Context, id string) error {
	updateData := map[string]interface{}{
		"deleted": true,
	}

	err := dbConn.Transaction(func(tx *gorm.DB) error {
		// Удаляем пользователя
		result := tx.Session(&gorm.Session{}).Model(&models.User{}).Where("id = ?", id).Updates(updateData)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Ничего не удалено")
		}

		// Находим задачи
		var tasks []models.Task
		if err := tx.Session(&gorm.Session{}).Model(models.Task{}).Where("created_by = ? OR assigned_to = ?", id, id).Find(&tasks).Error; err != nil {
			return err
		}

		for _, task := range tasks {
			if err := tx.Session(&gorm.Session{}).Model(&models.HelpRequest{}).Where("task_id = ?", task.ID).Updates(updateData).Error; err != nil {
				return err
			}
		}

		if err := tx.Session(&gorm.Session{}).Model(&models.Task{}).Where("created_by = ? OR assigned_to = ?", id, id).Updates(updateData).Error; err != nil {
			return err
		}

		var dailyReports []models.DailyReport
		if err := tx.Session(&gorm.Session{}).Model(models.DailyReport{}).Where("user_id = ? and deleted = ?", id, false).Find(&dailyReports).Error; err != nil {
			return err
		}

		for _, dailyReport := range dailyReports {
			if err := tx.Session(&gorm.Session{}).Model(&models.HelpRequest{}).Where("report_id = ?", dailyReport.ID).Updates(updateData).Error; err != nil {
				return err
			}
		}

		if err := tx.Session(&gorm.Session{}).Model(&models.DailyReport{}).Where("user_id = ?", id).Updates(updateData).Error; err != nil {
			return err
		}

		if err := tx.Session(&gorm.Session{}).Model(&models.UserRole{}).Where("assigned_by = ? OR user_id = ?", id, id).Updates(updateData).Error; err != nil {
			return err
		}

		if err := tx.Session(&gorm.Session{}).Model(&models.TeamMember{}).Where("user_id = ?", id).Updates(updateData).Error; err != nil {
			return err
		}

		if err := tx.Session(&gorm.Session{}).Model(&models.Attendance{}).Where("user_id = ?", id).Updates(updateData).Error; err != nil {
			return err
		}

		var problems []models.Problem
		if err := tx.Session(&gorm.Session{}).Model(models.Problem{}).Where("deleted = ? and creator_id = ?", false, id).Find(&problems).Error; err != nil {
			return err
		}

		for _, problem := range problems {
			if err := tx.Session(&gorm.Session{}).Model(&models.ForumMessage{}).Where("problem_id = ?", problem.ID).Updates(updateData).Error; err != nil {
				return err
			}

			if err := tx.Session(&gorm.Session{}).Model(&models.ReportProblem{}).Where("problem_id = ?", problem.ID).Updates(updateData).Error; err != nil {
				return err
			}

			if err := tx.Session(&gorm.Session{}).Model(&models.Problem{}).Where("id = ?", problem.ID).Updates(updateData).Error; err != nil {
				return err
			}
		}

		if err := tx.Session(&gorm.Session{}).Model(&models.Problem{}).Where("creator_id = ?", id).Updates(updateData).Error; err != nil {
			return err
		}

		if err := tx.Session(&gorm.Session{}).Model(&models.ForumMessage{}).Where("creator_id = ?", id).Updates(updateData).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return c.JSON(he.Code, map[string]string{"error": he.Message.(string)})
		}
		log.Printf("DB transaction error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении пользователя",
		})
	}

	deleteResponse := response.UserUniversalResponse{
		ID:      id,
		Message: "Пользователь с ID " + id + " удален",
	}
	return c.JSON(http.StatusOK, deleteResponse)
}*/

// deleteUser godoc
// @Summary Удаление пользователя(НЕ ИСПОЛЬЗОВАТЬ, ДОБАВЛЕНО ВРЕМЕННО ДЛЯ ТЕСТА ОШИБКИ 502)
// @Description Логическое удаление пользователя по ID, включая связанные данные (поле deleted = true)
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.UserUniversalResponse "Пользователь успешно удален"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении пользователя"
// @Router /user/full-delete/{id} [delete]
func deleteUser(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	return DeleteUserFunc(c, id)
}

func DeleteUserFunc(c echo.Context, id string) error {
	err := dbConn.Transaction(func(tx *gorm.DB) error {
		// Удаляем пользователя
		result := tx.Session(&gorm.Session{}).Model(&models.User{}).Where("id = ?", id).Delete(&models.User{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Ничего не удалено")
		}

		return nil
	})

	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return c.JSON(he.Code, map[string]string{"error": he.Message.(string)})
		}
		log.Printf("DB transaction error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении пользователя",
		})
	}

	deleteResponse := response.UserUniversalResponse{
		ID:      id,
		Message: "Пользователь с ID " + id + " удален",
	}
	return c.JSON(http.StatusOK, deleteResponse)
}

// banUser godoc
// @Summary Ban пользователя
// @Description Логическое удаление пользователя по ID
// @Tags Users
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.UserUniversalResponse "Пользователь успешно удален"
// @Failure 404 {object} map[string]string "Пользователь не найден"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении пользователя"
// @Router /user/{id} [delete]
func banUser(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")

	updateData := map[string]interface{}{
		"deleted":    true,
		"updated_at": time.Now(),
	}

	err := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).Model(&models.User{}).
			Where("id = ? AND deleted = FALSE", id).
			Updates(updateData)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Ничего не удалено")
		}
		return nil
	})

	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return c.JSON(he.Code, map[string]string{"error": he.Message.(string)})
		}
		log.Printf("DB transaction error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении пользователя",
		})
	}

	deleteResponse := response.UserUniversalResponse{
		ID:      id,
		Message: "Пользователь с ID " + id + " забанен",
	}
	return c.JSON(http.StatusOK, deleteResponse)
}

// restoreUser godoc
// @Summary Восстановление пользователя
// @Description Восстанавливает пользователя по email (логическое удаление снимается)
// @Tags Users
// @Accept json
// @Produce json
// @Param request body request.RestoreUserRequest true "Email пользователя для восстановления"
// @Security BearerAuth
// @Success 200 {object} response.UserUniversalResponse "Пользователь успешно восстановлен"
// @Failure 400 {object} map[string]string "Неверный запрос или пользователь не найден"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при восстановлении пользователя"
// @Router /user/restore [post]
func restoreUser(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	var req request.RestoreUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	updateData := map[string]interface{}{
		"deleted":    false,
		"updated_at": time.Now(),
	}

	var id string
	err := dbConn.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Session(&gorm.Session{}).Model(&models.User{}).
			Where("email = ? AND deleted = TRUE", req.Email).
			First(&user).Error; err != nil {
			return err
		}

		res := tx.Session(&gorm.Session{}).Model(&models.User{}).
			Where("email = ? AND deleted = TRUE", req.Email).
			Updates(updateData)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusBadRequest, "Ничего не восстановлено")
		}
		id = user.ID.String()
		return nil
	})

	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return c.JSON(he.Code, map[string]string{"error": he.Message.(string)})
		}
		log.Printf("DB transaction error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении пользователя",
		})
	}

	restoreResponse := response.UserUniversalResponse{
		ID:      id,
		Message: "Пользователь с email " + *req.Email + " восстановлен",
	}
	return c.JSON(http.StatusOK, restoreResponse)
}

// addUserRole godoc
// @Summary Добавление роли пользователю
// @Description Добавляет роль указанному пользователю
// @Tags Users
// @Accept json
// @Produce json
// @Param addRole body request.AddRoleUserRequest true "Данные для добавления роли пользователю"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.AddRoleUserResponse "Роль успешно добавлена"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 500 {object} map[string]string "Ошибка сервера при добавлении роли"
// @Router /user/role [post]
func addUserRole(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	var req request.AddRoleUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	var addResponse response.AddRoleUserResponse
	err := dbConn.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Session(&gorm.Session{}).Model(models.User{}).
			Select("id, profession").
			Where("id = ? AND deleted = FALSE", req.UserId).
			First(&user).Error; err != nil {
			return err
		}

		var role models.Role
		if err := tx.Session(&gorm.Session{}).Model(models.Role{}).
			Select("id").
			Where("id = ? AND deleted = FALSE", req.RoleId).
			First(&role).Error; err != nil {
			return err
		}

		var assignerId uuid.UUID
		assignerId, err := uuid.Parse(req.AssignerId)
		if err != nil {
			return err
		}

		now := time.Now()
		del := false

		userRole := models.UserRole{
			UserID:     user.ID,
			RoleID:     role.ID,
			AssignedAt: &now,
			Deleted:    &del,
			AssignedBy: &assignerId,
		}

		if err := tx.Session(&gorm.Session{}).Omit(clause.Associations).Create(&userRole).Error; err != nil {
			return err
		}

		addResponse = response.AddRoleUserResponse{
			UserId:  user.ID.String(),
			RoleId:  role.ID.String(),
			Message: "Роль успешно добавлена",
		}
		return nil
	})

	if err != nil {
		log.Printf("DB transaction error (add user role): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Не удалось добавить роль пользователям",
		})
	}

	return c.JSON(http.StatusOK, addResponse)
}

// removeUserRole godoc
// @Summary Удаление роли у пользователя
// @Description Удаляет роль у указанного пользователя
// @Tags Users
// @Accept json
// @Produce json
// @Param removeRole body request.RemoveRoleUserRequest true "Данные для удаления роли у пользователя"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.RemoveRoleUserResponse "Роль успешно удалена"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 404 {object} map[string]string "Ничего не удалено"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении роли"
// @Router /user/role [delete]
func removeUserRole(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	var req request.RemoveRoleUserRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	var deleteResponse response.RemoveRoleUserResponse
	err := dbConn.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Session(&gorm.Session{}).Model(models.User{}).
			Select("id, profession").
			Where("id = ? AND deleted = FALSE", req.UserId).
			First(&user).Error; err != nil {
			return err
		}

		var role models.Role
		if err := tx.Session(&gorm.Session{}).Model(models.Role{}).
			Select("id").
			Where("id = ? AND deleted = FALSE", req.RoleId).
			First(&role).Error; err != nil {
			return err
		}

		updateData := map[string]interface{}{
			"deleted":    true,
			"updated_at": time.Now(),
		}

		result := tx.Session(&gorm.Session{}).Model(models.UserRole{}).
			Where("user_id = ? AND role_id = ? AND deleted = FALSE", user.ID, role.ID).
			Updates(updateData)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "Ничего не удалено")
		}

		deleteResponse = response.RemoveRoleUserResponse{
			RoleId:  role.ID.String(),
			UserId:  user.ID.String(),
			Message: "Роль успешно удалена",
		}
		return nil
	})

	if err != nil {
		if he, ok := err.(*echo.HTTPError); ok {
			return c.JSON(he.Code, map[string]string{"message": he.Message.(string)})
		}
		log.Printf("DB transaction error (remove user role): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при удалении роли у пользователя",
		})
	}

	return c.JSON(http.StatusOK, deleteResponse)
}
