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

func RegisterRoleRoutes(e *echo.Echo) {
	projectGroup := e.Group("/role")
	{
		projectGroup.GET("/all/:page/:pagesize", GetAllRoles)
		projectGroup.GET("/:id", GetRoleById)
		projectGroup.POST("", CreateRole)
		projectGroup.PATCH("/:id", UpdateRole)
		projectGroup.DELETE("/:id", DeleteRole)
	}
}

// GetAllRoles godoc
// @Summary Получение списка всех ролей
// @Description Получает список всех ролей с учетом пагинации, исключая удаленные
// @Tags Roles
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Success 200 {object} response.GetAllRolesResponse "Список ролей успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении ролей"
// @Router /role/all/{page}/{pagesize} [get]
func GetAllRoles(c echo.Context) error {
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
			"error": "Ошибка при подсчете проектов",
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
		var name string
		if role.Name != nil {
			name = *role.Name
		}
		var Descripton string
		if role.Description != nil {
			Descripton = *role.Description
		}
		var updatedAt time.Time
		if role.UpdatedAt != nil {
			updatedAt = *role.UpdatedAt
		}
		roleList.Roles = append(roleList.Roles, response.GetRoleResponse{
			ID:          role.ID.String(),
			Name:        name,
			Description: Descripton,
			UpdatedAt:   updatedAt,
		})
	}
	return c.JSON(http.StatusOK, roleList)
}

// GetRoleById godoc
// @Summary Получение роли по ID
// @Description Получает данные роли по её уникальному идентификатору
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "ID роли"
// @Success 200 {object} response.GetRoleResponse "Роль успешно получена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор роли"
// @Failure 404 {object} map[string]string "Роль не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении роли"
// @Router /role/{id} [get]
func GetRoleById(c echo.Context) error {
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
	var name string
	if role.Name != nil {
		name = *role.Name
	}
	var Descripton string
	if role.Description != nil {
		Descripton = *role.Description
	}
	var updatedAt time.Time
	if role.UpdatedAt != nil {
		updatedAt = *role.UpdatedAt
	}
	roleResponse := response.GetRoleResponse{
		ID:          role.ID.String(),
		Name:        name,
		Description: Descripton,
		UpdatedAt:   updatedAt,
	}
	return c.JSON(http.StatusOK, roleResponse)
}

// CreateRole godoc
// @Summary Создание новой роли
// @Description Создает новую роль с указанными параметрами
// @Tags Roles
// @Accept json
// @Produce json
// @Param role body request.RoleCreateRequest true "Данные для создания роли"
// @Success 201 {object} response.RoleUniversalReport "Роль успешно создана"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании роли"
// @Router /role [post]
func CreateRole(c echo.Context) error {
	var req request.RoleCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	newUUID := uuid.New()

	del := false

	role := models.Role{
		ID:          newUUID,
		Name:        req.Name,
		Description: req.Description,
		Deleted: &del,
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

	createResponse := response.RoleUniversalReport{
		ID:      role.ID.String(),
		Message: "Роль успешно создана",
	}

	return c.JSON(http.StatusCreated, createResponse)
}

// UpdateRole godoc
// @Summary Обновление роли
// @Description Обновляет данные роли по её ID
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "ID роли"
// @Param role body request.RoleUpdateRequest true "Данные для обновления роли"
// @Success 200 {object} response.RoleUniversalReport "Роль успешно обновлена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении роли"
// @Router /role/{id} [patch]
func UpdateRole(c echo.Context) error {
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

	updateResponse := response.RoleUniversalReport{
		ID:      roleId.String(),
		Message: "Роль с ID " + id + " обновлен",
	}

	return c.JSON(http.StatusCreated, updateResponse)
}

// DeleteRole godoc
// @Summary Удаление роли
// @Description Логическое удаление роли по ID, включая связанные данные (поле deleted = true)
// @Tags Roles
// @Accept json
// @Produce json
// @Param id path string true "ID роли"
// @Success 200 {object} response.ProjectUniversalResponse "Роль успешно удалена"
// @Failure 404 {object} map[string]string "Роль не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении роли"
// @Router /role/{id} [delete]
func DeleteRole(c echo.Context) error {
	id := c.Param("id")
	updateData := make(map[string]interface{})
	updateData["deleted"] = true
	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(models.Role{}).Where("id = ?", id).Updates(updateData)
		if res.Error != nil {
			log.Printf("DB error (delete role): %v", res.Error)
			return res.Error
		}
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}

		if res = tx.Model(&models.UserRole{}).Where("role_id = ?", id).Updates(updateData); res.Error != nil {
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

	delResponse := response.ProjectUniversalResponse{ID: id, Message: "Роль с ID " + id + " удалена"}

	return c.JSON(http.StatusOK, delResponse)
}
