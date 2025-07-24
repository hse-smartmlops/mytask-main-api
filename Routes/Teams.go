package Routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterTeamRoutes(router *gin.Engine) {
	teamGroup := router.Group("/teams")
	{
		teamGroup.GET("/", GetTeams)
		teamGroup.GET(":id", GetTeamByID)
		teamGroup.POST("/", CreateTeam)
		teamGroup.PUT(":id", UpdateTeam)
		teamGroup.DELETE(":id", DeleteTeam)
	}
}

func GetTeams(c *gin.Context) {
	// TODO: Реализовать получение списка команд
	c.JSON(200, gin.H{"message": "Список команд"})
}

func GetTeamByID(c *gin.Context) {
	// TODO: Реализовать получение команды по id
	c.JSON(200, gin.H{"message": "Команда по id"})
}

func CreateTeam(c *gin.Context) {
	// TODO: Реализовать создание команды
	c.JSON(201, gin.H{"message": "Команда создана"})
}

func UpdateTeam(c *gin.Context) {
	// TODO: Реализовать обновление команды
	c.JSON(200, gin.H{"message": "Команда обновлена"})
}

func DeleteTeam(c *gin.Context) {
	// TODO: Реализовать удаление команды
	c.JSON(200, gin.H{"message": "Команда удалена"})
}
