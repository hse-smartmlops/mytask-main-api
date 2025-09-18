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

func RegisterRoleRoutes(e *echo.Echo) {
	group := e.Group("/role")
	group.Use(KeycloakAuthMiddleware)
	{
		group.GET("/all/:page/:pagesize", getAllRoles)
		group.GET("/:id", getRoleById)
		group.POST("", createRole)
		group.PATCH("/:id", updateRole)
		group.DELETE("/:id", deleteRole)
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
	if err := authorize(c); err != nil { return err }

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге страницы"})
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка при парсинге номера страницы"})
	}
	offset := (page - 1) * pageSize

	var totalCount int64
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.Role{}).
		Where("deleted = FALSE").
		Count(&totalCount).Error; err != nil {
		log.Printf("DB error (count roles): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчете ролей"})
	}

	var roles []models.Role
	if err := dbConn.Session(&gorm.Session{}).
		Model(&models.Role{}).
		Where("deleted = FALSE").
		Limit(pageSize).Offset(offset).
		Find(&roles).Error; err != nil {
		log.Printf("DB error (find roles): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении ролей из базы данных"})
	}

	out := response.GetAllRolesResponse{
		TotalCount: totalCount,
		PageSize:   pageSize,
		Page:       page,
	}
	for _, r := range roles {
		out.Roles = append(out.Roles, response.GetRoleResponse{
			ID:          r.ID.String(),
			Name:        utils.GetString(r.Name),
			Description: utils.GetString(r.Description),
			UpdatedAt:   utils.GetTime(r.UpdatedAt),
			CreatedAt:   utils.GetTime(r.CreatedAt),
		})
	}
	return c.JSON(http.StatusOK, out)
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
	if err := authorize(c); err != nil { return err }

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор роли"})
	}

	var role models.Role
	res := dbConn.Session(&gorm.Session{}).
		Model(&models.Role{}).
		Where("id = ? AND deleted = FALSE", roleID).
		First(&role)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Роль не найдена"})
		}
		log.Printf("DB error (find role by id): %v", res.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении роли из базы данных"})
	}

	return c.JSON(http.StatusOK, response.GetRoleResponse{
		ID:          role.ID.String(),
		Name:        utils.GetString(role.Name),
		Description: utils.GetString(role.Description),
		UpdatedAt:   utils.GetTime(role.UpdatedAt),
		CreatedAt:   utils.GetTime(role.CreatedAt),
	})
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
	if err := authorize(c); err != nil { return err }

	var req request.RoleCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	now := time.Now()
	del := false
	role := models.Role{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Deleted:     &del,
		CreatedAt:   &now,
	}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		return tx.Session(&gorm.Session{}).
			Model(&models.Role{}).
			Omit(clause.Associations).
			Create(&role).Error
	}); txErr != nil {
		log.Printf("DB transaction error (create role): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании роли"})
	}

	return c.JSON(http.StatusCreated, response.RoleUniversalResponse{
		ID:      role.ID.String(),
		Message: "Роль успешно создана",
	})
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
	if err := authorize(c); err != nil { return err }

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор роли"})
	}

	var req request.RoleUpdateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не удалось получить данные из запроса"})
	}

	update := map[string]interface{}{}
	if req.Name != nil        { update["name"] = *req.Name }
	if req.Description != nil { update["description"] = *req.Description }
	if len(update) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указаны поля для обновления"})
	}
	now := time.Now()
	update["updated_at"] = &now

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		res := tx.Session(&gorm.Session{}).
			Model(&models.Role{}).
			Where("id = ? AND deleted = FALSE", roleID).
			Updates(update)
		if res.Error != nil { return res.Error }
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не обновлено"})
		}
		return nil
	}); txErr != nil {
		log.Printf("DB transaction error (update role): %v", txErr)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении роли"})
	}

	return c.JSON(http.StatusOK, response.RoleUniversalResponse{
		ID:      roleID.String(),
		Message: "Роль успешно обновлена",
	})
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
	if err := authorize(c); err != nil { return err }

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор роли"})
	}

	delTrue := true
	now := time.Now()
	update := map[string]interface{}{"deleted": &delTrue, "updated_at": &now}

	if txErr := dbConn.Transaction(func(tx *gorm.DB) error {
		// сама роль
		res := tx.Session(&gorm.Session{}).
			Model(&models.Role{}).
			Where("id = ? AND deleted = FALSE", roleID).
			Updates(update)
		if res.Error != nil { return res.Error }
		if res.RowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		// помечаем связи user_roles как удалённые (если используешь soft delete там)
		if res = tx.Session(&gorm.Session{}).
			Model(&models.UserRole{}).
			Where("role_id = ?", roleID).
			Updates(update); res.Error != nil {
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

	return c.JSON(http.StatusOK, response.RoleUniversalResponse{
		ID:      roleID.String(),
		Message: "Роль удалена",
	})
}
