package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterProjectRoutes(e *echo.Echo) {
	projectGroup := e.Group("/project")
	{
		projectGroup.GET("", GetAllProjects)
		projectGroup.GET("/id/:id", GetProjectByID)
		projectGroup.POST("", CreateProject)
		projectGroup.PUT("/:id", UpdateProject)
		projectGroup.DELETE("/:id", DeleteProject)
	}
}

func GetAllProjects(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{"data": "Все проекты"})
}

func GetProjectByID(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]interface{}{"id": id})
}

func CreateProject(c echo.Context) error {
	// TODO: Реализовать создание нового проекта
	return c.JSON(http.StatusCreated, map[string]interface{}{"message": "Проект создан"})
}

func UpdateProject(c echo.Context) error {
	// TODO: Реализовать обновление проекта
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "Проект с ID " + id + " обновлен"})
}

func DeleteProject(c echo.Context) error {
	// TODO: Реализовать удаление проекта
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "Проект с ID " + id + " удален"})
}
