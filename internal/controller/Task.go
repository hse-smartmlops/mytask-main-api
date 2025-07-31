package controller

import (
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func RegisterTaskRoutes(e *echo.Echo) {
	taskGroup := e.Group("/tasks")
	{
		taskGroup.GET("", GetAllTasks)
		taskGroup.GET("/:id", GetTaskByID)
		taskGroup.GET("/board/:boardId", GetTasksByBoardID)
		taskGroup.GET("/project/:projectId", GetTasksByProjectID)
		taskGroup.GET("/filter", GetTasksByFilter)
		taskGroup.POST("", CreateTask)
		taskGroup.PUT("/:id", UpdateTask)
		taskGroup.DELETE("/:id", DeleteTask)
	}
}

// GetAllTasks godoc
// @Summary Получение списка задач
// @Description Возвращает список всех задач (заглушка)
// @Tags Tasks
// @Produce json
// @Success 200 {object} response.TaskListResponse
// @Router /tasks [get]
func GetAllTasks(c echo.Context) error {
	// Заглушка с примером данных
	return c.JSON(http.StatusOK, response.TaskListResponse{
		Tasks: []response.TaskResponse{
			createSampleTask("1", "Пример задачи", "todo"),
		},
		TotalCount: 1,
		Page:       1,
		PageSize:   10,
	})
}

// GetTaskByID godoc
// @Summary Получение задачи по ID
// @Description Возвращает задачу по ID (заглушка)
// @Tags Tasks
// @Param id path string true "ID задачи"
// @Produce json
// @Success 200 {object} response.TaskResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /tasks/{id} [get]
func GetTaskByID(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, createSampleTask(id, "Пример задачи", "todo"))
}

// GetTasksByBoardID godoc
// @Summary Получение задач по доске
// @Description Возвращает задачи для указанной доски (заглушка)
// @Tags Tasks
// @Param boardId path string true "ID доски"
// @Produce json
// @Success 200 {object} response.TaskListResponse
// @Router /tasks/board/{boardId} [get]
func GetTasksByBoardID(c echo.Context) error {
	boardID := c.Param("boardId")
	return c.JSON(http.StatusOK, response.TaskListResponse{
		Tasks: []response.TaskResponse{
			createSampleTask("1", "Задача для доски "+boardID, "todo"),
		},
	})
}

// GetTasksByProjectID godoc
// @Summary Получение задач по проекту
// @Description Возвращает задачи для указанного проекта (заглушка)
// @Tags Tasks
// @Param projectId path string true "ID проекта"
// @Produce json
// @Success 200 {object} response.TaskListResponse
// @Router /tasks/project/{projectId} [get]
func GetTasksByProjectID(c echo.Context) error {
	projectID := c.Param("projectId")
	return c.JSON(http.StatusOK, response.TaskListResponse{
		Tasks: []response.TaskResponse{
			createSampleTask("1", "Задача для проекта "+projectID, "todo"),
		},
	})
}

// GetTasksByFilter godoc
// @Summary Фильтрация задач
// @Description Возвращает отфильтрованный список задач (заглушка)
// @Tags Tasks
// @Param project_id query string false "ID проекта"
// @Param status query string false "Статус задачи"
// @Param priority query int false "Приоритет задачи"
// @Produce json
// @Success 200 {object} response.TaskListResponse
// @Router /tasks/filter [get]
func GetTasksByFilter(c echo.Context) error {
	filter := new(request.TaskFilter)
	if err := c.Bind(filter); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid filter parameters"})
	}

	// Заглушка с применением фильтров
	task := createSampleTask("1", "Отфильтрованная задача", "todo")
	if filter.Status != nil {
		task.Status = *filter.Status
	}
	if filter.Priority != nil {
		task.Priority = *filter.Priority
	}

	return c.JSON(http.StatusOK, response.TaskListResponse{
		Tasks: []response.TaskResponse{task},
	})
}

// CreateTask godoc
// @Summary Создание задачи
// @Description Создает новую задачу (заглушка)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param input body request.CreateTask true "Данные задачи"
// @Success 201 {object} response.TaskCreateResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /tasks [post]
func CreateTask(c echo.Context) error {
	var req request.CreateTask
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	// Валидация должна быть здесь (пропущена для краткости)

	return c.JSON(http.StatusCreated, response.TaskCreateResponse{
		ID:      "generated-id",
		Message: "Task will be created after DB integration",
	})
}

// UpdateTask godoc
// @Summary Обновление задачи
// @Description Обновляет существующую задачу (заглушка)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Param input body request.UpdateTask true "Обновляемые данные"
// @Success 200 {object} response.TaskUpdateResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /tasks/{id} [put]
func UpdateTask(c echo.Context) error {
	id := c.Param("id")
	var req request.UpdateTask

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request data"})
	}

	return c.JSON(http.StatusOK, response.TaskUpdateResponse{
		ID:      id,
		Message: "Task will be updated after DB integration",
	})
}

// DeleteTask godoc
// @Summary Удаление задачи
// @Description Удаляет задачу по ID (заглушка)
// @Tags Tasks
// @Param id path string true "ID задачи"
// @Produce json
// @Success 200 {object} response.TaskDeleteResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /tasks/{id} [delete]
func DeleteTask(c echo.Context) error {
	id := c.Param("id")
	return c.JSON(http.StatusOK, response.TaskDeleteResponse{
		ID:      id,
		Message: "Task will be deleted after DB integration",
	})
}

// Вспомогательная функция для создания тестовых данных
func createSampleTask(id, name, status string) response.TaskResponse {
	now := time.Now()
	return response.TaskResponse{
		ID:          id,
		ProjectID:   "project-1",
		Name:        name,
		Description: "Пример описания задачи",
		Status:      status,
		Priority:    5,
		CreatedBy: response.UserShort{
			ID:        "user-1",
			FirstName: "Иван",
			LastName:  "Иванов",
			Email:     "ivan@example.com",
		},
		StartDate:  now,
		CreatedAt:  now,
		UpdatedAt:  now,
		TimeSpent:  0,
	}
}