package controller

import (
	"github.com/gin-gonic/gin"
)

func RegisterTaskRouters(router *gin.Engine) {
	taskGroup := router.Group("/task") 
	{
		taskGroup.GET("/", GetAllTasks)
		taskGroup.GET("/id/:id", GetTaskByID)
		taskGroup.GET("/board/:boardID", GetTasksByBoardID)
		taskGroup.GET("/project/:projectID", GetTasksByProjectID)
		taskGroup.GET("/filter", GetTasksByFilter)
	}
}

func GetAllTasks(c *gin.Context) {
    c.JSON(200, gin.H{"data": "Все задачи"})
}

 func GetTaskByID(c *gin.Context) {
	id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
}

 func GetTasksByBoardID(c *gin.Context) {
	boardID := c.Param("boardID")
    c.JSON(200, gin.H{"boardID": boardID})
}

 func GetTasksByProjectID(c *gin.Context) {
	projectID:= c.Param("projectID")
    c.JSON(200, gin.H{"projectID": projectID})
}

 func GetTasksByFilter(c *gin.Context) {
	filter := c.Param("filter")
    c.JSON(200, gin.H{"filter": filter})
	
}




