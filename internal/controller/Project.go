package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterProjectRoutes(e *echo.Echo) {
	projectGroup := e.Group("/projects")
	{
		projectGroup.GET("", GetAllProjects)
		projectGroup.GET("/id/:id", GetProjectByID)
		projectGroup.POST("", CreateProject)
		projectGroup.PUT("/:id", UpdateProject)
		projectGroup.DELETE("/:id", DeleteProject)
	}
}

// GetAllProjects godoc
// @Summary Получение списка всех проектов
// @Description Возвращает список всех проектов
// @Tags Projects
// @Produce json
// @Success 200 {object} map[string]string
// @Router /projects [get]
func GetAllProjects(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"data": "Все проекты"})
}

// GetProjectByID godoc
// @Summary      Получение проекта по ID
// @Description  Возвращает проект по его идентификатору
// @Tags         Projects
// @Param        id path string true "Project ID"
// @Produce      json
// @Success      200 {object} map[string]string
// @Failure      404 {object} map[string]string "Проект не найден"
// @Router       /projects/{id} [get]
func GetProjectByID(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]interface{}{"id": id})
}


// CreateProject godoc
// @Summary Создание проекта
// @Description Создает новый проект
// @Tags Projects
// @Accept json
// @Produce json
// @Param project body object true "Данные проекта"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /projects [post]
func CreateProject(c echo.Context) error {
	// TODO: Реализовать создание проекта
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Проект создан",
	})
}

// UpdateProject godoc
// @Summary Обновление проекта
// @Description Обновляет существующий проект
// @Tags Projects
// @Accept json
// @Produce json
// @Param id path string true "ID проекта"
// @Param updates body object true "Новые данные"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /projects/{id} [put]
func UpdateProject(c echo.Context) error {
	id := c.Param("id")
	// TODO: Реализовать обновление проекта
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Проект с ID " + id + " обновлен",
	})
}

// DeleteProject godoc
// @Summary Удаление проекта
// @Description Удаляет проект по ID
// @Tags Projects
// @Param id path string true "ID проекта"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string
// @Router /projects/{id} [delete]
func DeleteProject(c echo.Context) error {
	id := c.Param("id")
	// TODO: Реализовать удаление проекта
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Проект с ID " + id + " удален",
	})
}
