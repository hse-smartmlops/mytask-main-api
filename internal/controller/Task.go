package controller

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func RegisterTaskRoutes(e *echo.Echo) {
	taskGroup := e.Group("/tasks")
	{
		taskGroup.GET("", GetAllTasks)
		taskGroup.GET("/:id", GetTaskByID)
		taskGroup.GET("/board/:boardID", GetTasksByBoardID)
		taskGroup.GET("/project/:projectID", GetTasksByProjectID)
		taskGroup.GET("/filter", GetTasksByFilter)
		taskGroup.POST("", CreateTask)
		taskGroup.PUT("/:id", UpdateTask)
		taskGroup.DELETE("/:id", DeleteTask)
	}
}
 // GetAllTasks godoc
// @Summary Получение списка всех задач
// @Description Возвращает список всех задач
// @Tags Tasks
// @Produce json
// @Success 200 {object} map[string]string
// @Router /tasks [get]
func GetAllTasks(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"data": "Все задачи"})
}

// GetTaskByID godoc
// @Summary Получение задачи по ID
// @Description Возвращает задачу по её идентификатору
// @Tags Tasks
// @Param id path string true "Task ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Router /tasks/{id} [get]
func GetTaskByID(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, map[string]string{"id": id})
}

// GetTasksByBoardID godoc
// @Summary Получение задач по ID доски
// @Description Возвращает список задач, принадлежащих указанной доске
// @Tags Tasks
// @Param boardID path string true "Board ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Router /tasks/board/{boardID} [get]
func GetTasksByBoardID(c echo.Context) error {
	boardID := c.Param("boardID")
	return c.JSON(http.StatusOK, map[string]string{"boardID": boardID})
}

// GetTasksByProjectID godoc
// @Summary Получение задач по ID проекта
// @Description Возвращает список задач, принадлежащих указанному проекту
// @Tags Tasks
// @Param projectID path string true "Project ID"
// @Produce json
// @Success 200 {object} map[string]string
// @Router /tasks/project/{projectID} [get]
func GetTasksByProjectID(c echo.Context) error {
	projectID := c.Param("projectID")
	return c.JSON(http.StatusOK, map[string]string{"projectID": projectID})
}

// GetTasksByFilter godoc
// @Summary Фильтрация задач
// @Description Возвращает список задач, отфильтрованных по параметрам
// @Tags Tasks
// @Param filter query string false "Filter criteria"
// @Produce json
// @Success 200 {object} map[string]string
// @Router /tasks/filter [get]
func GetTasksByFilter(c echo.Context) error {
	filter := c.QueryParam("filter")
	return c.JSON(http.StatusOK, map[string]string{"filter": filter})
}

// CreateTask godoc
// @Summary Создание новой задачи
// @Description Создает новую задачу с указанными параметрами
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task body controller.TaskRequest true "Данные задачи"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Router /tasks [post]
func CreateTask(c echo.Context) error {
	type TaskRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		ProjectID   uint   `json:"project_id"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
	}

	task := new(TaskRequest)
	if err := c.Bind(task); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
	}
	// Возврат созданной задачи 
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Задача успешно создана",
		"task":    task,
	})
}

// UpdateTask godoc
// @Summary Обновление задачи
// @Description Обновляет данные существующей задачи
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Param updates body controller.TaskRequest true "Обновляемые данные"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string "Неверный формат данных"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Router /tasks/{id} [put]
func UpdateTask(c echo.Context) error {
	id := c.Param("id")
	
	type TaskRequest struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Status      string `json:"status"`
		Priority    string `json:"priority"`
	}

	taskUpdate := new(TaskRequest)
	if err := c.Bind(taskUpdate); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат данных"})
	}
	// Возврат обновленных данных
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Задача успешно обновлена",
		"id":      id,
		"updates": taskUpdate,
	})
}

// DeleteTask godoc
// @Summary Удаление задачи
// @Description Удаляет задачу по её идентификатору
// @Tags Tasks
// @Param id path string true "Task ID"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Router /tasks/{id} [delete]
func DeleteTask(c echo.Context) error {
	id := c.Param("id")
	// Подтверждение удаления
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Задача успешно удалена",
		"id":      id,
	})
}