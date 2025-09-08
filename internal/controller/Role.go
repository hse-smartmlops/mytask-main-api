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
)

func RegisterRoleRoutes(e *echo.Echo) {
	projectGroup := e.Group("/role")
	projectGroup.Use(KeycloakAuthMiddleware)
	{
		projectGroup.GET("/all/:page/:pagesize", getAllRoles)
		projectGroup.GET("/:id", getRoleById)
		projectGroup.POST("", createRole)
		projectGroup.PATCH("/:id", updateRole)
		projectGroup.DELETE("/:id", deleteRole)
	}
}

// getAllRoles godoc
// @Summary Получение списка всех ролей
// @Description Получает список всех ролей с учетом пагинации, исключая удаленные
// @Tags Roles
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.GetAllRolesResponse "Список ролей успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении ролей"
// @Router /role/all/{page}/{pagesize} [get]
func getAllRoles(c echo.Context) error {
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
	result := dbConn.Session(&gorm.Session{}).Model(models.Role{}).Where("deleted = ?", false).Count(&totalCount)
	if result.Error != nil {
		log.Printf("DB error (count roles): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при подсчете ролей",
		})
	}

	var roles []models.Role
	if err := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).
		Limit(pageSize).
		Offset(offset).
		Find(&roles).Error; err != nil {
		log.Printf("DB error (find roles): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении ролей из базы данных",
		})
	}

	roleList := response.GetAllRolesResponse{
		TotalCount: totalCount,
		PageSize:   pageSize,
		Page:       page,
	}

	for _, role := range roles {
		roleList.Roles = append(roleList.Roles, response.GetRoleResponse{
			ID:          role.ID.String(),
			Name:        utils.GetString(role.Name),
			Description: utils.GetString(role.Description),
			UpdatedAt:   utils.GetTime(role.UpdatedAt),
			CreatedAt:   utils.GetTime(role.CreatedAt),
		})
	}
	return c.JSON(http.StatusOK, roleList)
}

// getRoleById godoc
// @Summary Получение роли по ID
// @Description Получает данные роли по её уникальному идентификатору
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "ID роли"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.GetRoleResponse "Роль успешно получена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор роли"
// @Failure 404 {object} map[string]string "Роль не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении роли"
// @Router /role/{id} [get]
func getRoleById(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	roleId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор роли",
		})
	}

	var role models.Role
	result := dbConn.Session(&gorm.Session{}).Where("deleted = ?", false).First(&role, "id = ?", roleId)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Роль не найден",
			})
		}
		log.Printf("DB error (find role by id): %v", result.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при получении роли из базы данных",
		})
	}

	roleResponse := response.GetRoleResponse{
		ID:          role.ID.String(),
		Name:        utils.GetString(role.Name),
		Description: utils.GetString(role.Description),
		UpdatedAt:   utils.GetTime(role.UpdatedAt),
		CreatedAt:   utils.GetTime(role.CreatedAt),
	}
	return c.JSON(http.StatusOK, roleResponse)
}

// createRole godoc
// @Summary Создание новой роли
// @Description Создает новую роль с указанными параметрами
// @Tags Roles
// @Accept json
// @Produce json
// @Param role body request.RoleCreateRequest true "Данные для создания роли"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 201 {object} response.RoleUniversalResponse "Роль успешно создана"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании роли"
// @Router /role [post]
func createRole(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	var req request.RoleCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	del := false

	now := time.Now()

	role := models.Role{
		ID:          newUUID,
		Name:        req.Name,
		Description: req.Description,
		Deleted: &del,
		CreatedAt: &now,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Create(&role); res.Error != nil {
			log.Printf("DB error (create role): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (create role): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании роли"})
	}

	createResponse := response.RoleUniversalResponse{
		ID:      role.ID.String(),
		Message: "Роль успешно создана",
	}

	return c.JSON(http.StatusCreated, createResponse)
}

// updateRole godoc
// @Summary Обновление роли
// @Description Обновляет данные роли по её ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "ID роли"
// @Param role body request.RoleUpdateRequest true "Данные для обновления роли"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.RoleUniversalResponse "Роль успешно обновлена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении роли"
// @Router /role/{id} [patch]
func updateRole(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	roleId, err := uuid.Parse(id)
	if err != nil {
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор роли",
		})
	}

	var req request.RoleUpdateRequest
	if err = c.Bind(&req); err != nil {
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
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не указаны поля для обновления",
		})
	}
	updateData["updated_at"] = time.Now()

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		if res := tx.Model(models.Role{}).Where("id = ?", roleId).Updates(updateData); res.Error != nil {
			log.Printf("DB error (update role): %v", res.Error)
			return res.Error
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update role): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении роль"})
	}

	updateResponse := response.RoleUniversalResponse{
		ID:      roleId.String(),
		Message: "Роль с ID " + id + " обновлен",
	}

	return c.JSON(http.StatusOK, updateResponse)
}

// deleteRole godoc
// @Summary Удаление роли
// @Description Логическое удаление роли по ID, включая связанные данные (поле deleted = true)
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "ID роли"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.RoleUniversalResponse "Роль успешно удалена"
// @Failure 404 {object} map[string]string "Роль не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении роли"
// @Router /role/{id} [delete]
func deleteRole(c echo.Context) error {
	if err := authorize(c); err != nil {
		return err
	}
	id := c.Param("id")
	roleId, err := uuid.Parse(id)
	if err != nil{
		log.Printf("UUID parse error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор роли",
		})
	}
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(models.Role{}).Where("id = ? and deleted = ?", roleId, false).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete role): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}

		if res = tx.Model(&models.UserRole{}).Where("role_id = ?", roleId).Updates(updateData); res.Error != nil {
			log.Printf("DB error (delete user roles): %v", res.Error)
			return res.Error
		}

		return nil
	}); txErr != nil {
		if he, ok := txErr.(*echo.HTTPError); ok {
			return c.JSON(he.Code, he.Message)
		}
		log.Printf("DB transaction error (delete role): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении роли"})
	}

	delResponse := response.RoleUniversalResponse{ID: id, Message: "Роль с ID " + id + " удалена"}

	return c.JSON(http.StatusOK, delResponse)
}
