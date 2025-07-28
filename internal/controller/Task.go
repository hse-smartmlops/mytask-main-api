package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterTaskRoutes(e *echo.Echo) {
	taskGroup := e.Group("/task")
	{
		taskGroup.GET("", GetAllTasks)
		taskGroup.GET("/id/:id", GetTaskByID)
		taskGroup.GET("/board/:boardID", GetTasksByBoardID)
		taskGroup.GET("/project/:projectID", GetTasksByProjectID)
		taskGroup.GET("/filter", GetTasksByFilter)
	}
}

func GetAllTasks(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"data": "Все задачи"})
}

func GetTaskByID(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{"id": id})
}

func GetTasksByBoardID(c echo.Context) error {
	boardID := c.Param("boardID")
	return c.JSON(http.StatusOK, map[string]string{"boardID": boardID})
}

func GetTasksByProjectID(c echo.Context) error {
	projectID := c.Param("projectID")
	return c.JSON(http.StatusOK, map[string]string{"projectID": projectID})
}

func GetTasksByFilter(c echo.Context) error {
	// В Echo параметры query string получают через QueryParam()
	filter := c.QueryParam("filter")
	return c.JSON(http.StatusOK, map[string]string{"filter": filter})
}