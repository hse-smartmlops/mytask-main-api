package controller

import (
	"github.com/gin-gonic/gin"
)

func RegisterProjectRouters(router *gin.Engine) {
	projectGroup := router.Group("/project") 
	{
		projectGroup.GET("/", GetAllProjects)
		projectGroup.GET("/id/:id", GetProjectByID)
		projectGroup.POST("/", CreateProject)
		projectGroup.PUT("/:id", UpdateProject)
		projectGroup.DELETE("/:id", DeleteProject)
	}
}

func GetAllProjects(c *gin.Context) {
    c.JSON(200, gin.H{"data": "Все проекты"})
}

 func GetProjectByID(c *gin.Context) {
	id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
 }

func CreateProject(c *gin.Context) {
	// TODO: Реализовать создание проекта
	c.JSON(201, gin.H{"message": "Проект создан"})
}

func UpdateProject(c *gin.Context) {
	// TODO: Реализовать обновление проекта
	c.JSON(200, gin.H{"message": "Проект обновлен"})
}

func DeleteProject(c *gin.Context) {
	// TODO: Реализовать удаление проекта
	c.JSON(200, gin.H{"message": "Проект удален"})
}
