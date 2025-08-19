package main

import (
	"emplacc-api/internal/controller"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	controller.RegisterTeamRoutes(e)
	controller.RegisterProjectRoutes(e)
	controller.RegisterTaskRoutes(e)
	controller.RegisterBoardRoutes(e)
	controller.RegisterUserRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
}
