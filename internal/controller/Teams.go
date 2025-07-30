package controller

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func RegisterTeamRoutes(e *echo.Echo) {
	teamGroup := e.Group("/team")
	teamGroup.GET("/all", GetTeams)
	teamGroup.GET(":id", GetTeamByID)
	teamGroup.POST("/", CreateTeam)
	teamGroup.PATCH(":id", UpdateTeam)
	teamGroup.DELETE(":id", DeleteTeam) // Только для admin, добавить проверку роли
}

func GetTeams(c echo.Context) error {
	// TODO: Реализовать получение списка команд
	return c.JSON(http.StatusOK, map[string]string{"message": "Список команд"})
}

func GetTeamByID(c echo.Context) error {
	// TODO: Реализовать получение команды по id
	return c.JSON(http.StatusOK, map[string]string{"message": "Команда по id"})
}

func CreateTeam(c echo.Context) error {
	// TODO: Реализовать создание команды
	return c.JSON(http.StatusCreated, map[string]string{"message": "Команда создана"})
}

func UpdateTeam(c echo.Context) error {
	// TODO: Реализовать обновление команды
	return c.JSON(http.StatusOK, map[string]string{"message": "Команда обновлена"})
}

func DeleteTeam(c echo.Context) error {
	// TODO: Реализовать удаление команды (только для admin)
	return c.JSON(http.StatusOK, map[string]string{"message": "Команда удалена"})
}
