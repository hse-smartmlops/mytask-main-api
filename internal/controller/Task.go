package controller

import (
	"emplacc-api/internal/dto/request"
	"emplacc-api/internal/dto/response"
	"emplacc-api/internal/grpc/client"
	"emplacc-api/internal/service"
	utils "emplacc-api/internal/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type TaskController struct {
	taskService service.TaskService
	userService service.UserService
	projectService service.ProjectService
	llmClient      *client.LLMClient
}

func NewTaskController(taskService service.TaskService, userService service.UserService, projectService service.ProjectService, llmClient *client.LLMClient) *TaskController {
	return &TaskController{
		taskService: taskService,
		userService: userService,
		projectService: projectService,
		llmClient:      llmClient,
	}
}

func RegisterTaskRoutes(e *echo.Echo, taskService service.TaskService, userService service.UserService, projectService service.ProjectService, llmClient *client.LLMClient,) {
	controller := NewTaskController(taskService, userService, projectService, llmClient)
	taskGroup := e.Group("/task")
	{
		taskGroup.GET("/all/:page/:pagesize", controller.GetAllTasks)
		taskGroup.GET("/:id", controller.GetTaskByID)
		taskGroup.POST("", controller.CreateTask)
		taskGroup.PATCH("/:id", controller.UpdateTask)
		taskGroup.DELETE("/:id", controller.DeleteTask)
		taskGroup.GET("/user/:id/:page/:pagesize", controller.GetTasksByUserId)
		taskGroup.POST("/move", controller.TaskMoveFunc)
		taskGroup.GET("/user/:user_id/project/:project_id/:page/:pagesize", controller.GetUserProjectTasks)
		taskGroup.POST("/:id/improve-report", controller.ImproveTaskReport)
		taskGroup.GET("/user/:id/:page/:pagesize/active", controller.GetActiveTasksByUserId)
		taskGroup.GET("/board-project/:id", controller.GetTaskBoardAndProject)
	}
}

// GetAllTasks godoc
// @Summary Получение списка всех задач
// @Description Получает список всех задач с учетом пагинации, исключая удаленные
// @Tags Tasks
// @Accept json
// @Produce json
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.TaskListResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Ошибка в запросе"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/all/{page}/{pagesize} [get]
func (tc *TaskController) GetAllTasks(c echo.Context) error {
	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	tasks, totalCount, err := tc.taskService.GetAllTasks(page, pageSize)
	if err != nil {
		log.Printf("service error (get all tasks): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при подсчёте задач"})
	}

	taskList := response.TaskListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		Tasks:      make([]response.TaskShort, 0, len(tasks)),
	}

	for _, task := range tasks {
		taskList.Tasks = append(taskList.Tasks, response.TaskShort{
			ID:        task.ID.String(),
			StatusID:  task.StatusID.String(), // ← uuid.UUID → string
			Name:      utils.GetString(task.Name),
			Priority:  utils.GetInt16(task.Priority),
			StartDate: utils.GetTime(task.StartDate),   // ← важно: обработка nil
			Deadline:  utils.GetTime(task.Deadline),    // ← важно: обработка nil
			CreatedAt: utils.GetTime(task.CreatedAt),   // ← важно
			UpdatedAt: utils.GetTime(task.UpdatedAt),   // ← важно
		})
	}

	return c.JSON(http.StatusOK, taskList)
}

// GetTaskByID godoc
// @Summary Получение задачи по ID
// @Description Получает данные задачи по её уникальному идентификатору, включая статус и пользователей
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Security BearerAuth
// @Success 200 {object} response.GetTaskByIDResponse "Задача успешно получена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор задачи"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задачи"
// @Router /task/{id} [get]
func (tc *TaskController) GetTaskByID(c echo.Context) error {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор задачи",
		})
	}

	task, err := tc.taskService.GetTaskByID(taskID)
	if err != nil {
		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}
		log.Printf("service error (find task by id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задачи"})
	}

	// Создатель
	creatorInfo := response.UserShort{}
	if task.CreatedByUser != nil {
		creatorInfo = response.UserShort{
			ID:        task.CreatedByUser.ID.String(),
			FirstName: task.CreatedByUser.FirstName,
			LastName:  task.CreatedByUser.LastName,
		}
	}

	// Исполнитель
	var assignerInfo response.UserShort
	if task.AssignedToUser != nil {
		assignerInfo = response.UserShort{
			ID:        task.AssignedToUser.ID.String(),
			FirstName: task.AssignedToUser.FirstName,
			LastName:  task.AssignedToUser.LastName,
		}
	}

	taskResponse := response.GetTaskByIDResponse{
		ID:            task.ID.String(),
		StatusID:      task.StatusID.String(), // ← только ID статуса
		Name:          utils.GetString(task.Name),
		Description:   utils.GetString(task.Description),
		Priority:      utils.GetInt16(task.Priority),
		CreatedBy:     creatorInfo,
		AssignedTo:    assignerInfo, // ← указатель, чтобыomitempty работал
		Deadline:      utils.GetTime(task.Deadline),
		TimeSpent:     utils.GetString(task.TimeSpent),
		StartDate:     utils.GetTime(task.StartDate),
		GitlabIssueID: utils.GetInt(task.GitlabIssueID),
		Category:      utils.GetInt8(task.Category),
		UpdatedAt:     utils.GetTime(task.UpdatedAt),
		CreatedAt:     utils.GetTime(task.CreatedAt),
	}

	return c.JSON(http.StatusOK, taskResponse)
}

/*
// GetTasksByProjectID godoc
// @Summary Получение задач по ID проекта
// @Description Получает список задач, связанных с указанным проектом через доски и статусы
// @Tags Tasks
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.TaskListResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Некорректный идентификатор проекта"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/project/{projectId}/{page}/{pagesize} [get]
func (tc *TaskController) GetTasksByProjectID(c echo.Context) error {
	projectID, err := uuid.Parse(c.Param("projectId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор проекта"})
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	tasks, totalCount, err := tc.taskService.GetTasksByProjectID(projectID, page, pageSize)
	if err != nil {
		if err.Error() == "project not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Проект не найден"})
		}
		log.Printf("service error (get tasks by project id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении статусов проекта"})
	}

	taskList := response.TaskListResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: totalCount,
		Tasks:      make([]response.TaskShort, 0, len(tasks)),
	}

	for _, task := range tasks {
		taskList.Tasks = append(taskList.Tasks, response.TaskShort{
			ID:        task.ID.String(),
			StatusID:  task.StatusID.String(),
			Name:      utils.GetString(task.Name),
			Priority:  utils.GetInt16(task.Priority),
			StartDate: utils.GetTime(task.StartDate),
			Deadline:  utils.GetTime(task.Deadline),
			CreatedAt: utils.GetTime(task.CreatedAt),
			UpdatedAt: utils.GetTime(task.UpdatedAt),
		})
	}

	return c.JSON(http.StatusOK, taskList)
}*/

// CreateTask godoc
// @Summary Создание новой задачи
// @Description Создает новую задачу с указанными параметрами
// @Tags Tasks
// @Accept json
// @Produce json
// @Param task body request.TaskCreateRequest true "Данные для создания задачи"
// @Security BearerAuth
// @Success 201 {object} response.TaskUniversaResponse "Задача успешно создана"
// @Failure 400 {object} map[string]string "Ошибка в запросе или некорректные идентификаторы"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Исполнитель, поручитель или статус не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при создании задачи"
// @Router /task [post]
func (tc *TaskController) CreateTask(c echo.Context) error {
	var req request.TaskCreateRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Не удалось получить данные из запроса",
		})
	}

	taskID, err := tc.taskService.CreateTask(req)
	if err != nil {
		if err.Error() == "invalid assigned_to" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор исполнителя"})
		}
		if err.Error() == "assignee not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Исполнитель не найден"})
		}
		if err.Error() == "invalid creator_id" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор поручителя"})
		}
		if err.Error() == "creator not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Поручитель не найден"})
		}
		if err.Error() == "invalid status_id" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор статуса"})
		}
		if err.Error() == "status not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Статус не найден или привязан к удалённой доске"})
		}
		log.Printf("service error (create task): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при создании задачи"})
	}

	return c.JSON(http.StatusCreated, response.TaskUniversaResponse{
		ID:      taskID.String(),
		Message: "Задача создана",
	})
}

// UpdateTask godoc
// @Summary Обновление задачи
// @Description Обновляет данные задачи по её ID
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Param task body request.TaskUpdateRequest true "Данные для обновления задачи"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TaskUniversaResponse "Задача успешно обновлена"
// @Failure 400 {object} map[string]string "Некорректный идентификатор или ошибка в запросе"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при обновлении задачи"
// @Router /task/{id} [patch]
func (tc *TaskController) UpdateTask(c echo.Context) error {
	id := c.Param("id")
	taskID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор задачи"})
	}

	var req request.TaskUpdateRequest
	if err = c.Bind(&req); err != nil {
		log.Printf("Bind error (update task): %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверные данные запроса"})
	}

	err = tc.taskService.UpdateTask(taskID, req)
	if err != nil {
		if err.Error() == "no fields to update" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Нет данных для обновления"})
		}
		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}
		if err.Error() == "invalid assigned_to" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор исполнителя"})
		}
		log.Printf("service error (update task): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении задачи"})
	}

	return c.JSON(http.StatusOK, response.TaskUniversaResponse{
		ID:      taskID.String(),
		Message: "Задача изменена",
	})
}

// DeleteTask godoc
// @Summary Удаление задачи
// @Description Логическое удаление задачи по ID, включая связанные отчеты (поле deleted = true)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Security BearerAuth
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Success 200 {object} response.TaskUniversaResponse "Задача успешно удалена"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Ошибка сервера при удалении задачи"
// @Router /task/{id} [delete]
func (tc *TaskController) DeleteTask(c echo.Context) error {
	id := c.Param("id")
	taskID, err := uuid.Parse(id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор задачи"})
	}

	err = tc.taskService.DeleteTask(taskID)
	if err != nil {
		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Ничего не удалено"})
		}
		log.Printf("service error (delete task): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при удалении задачи"})
	}

	return c.JSON(http.StatusOK, response.TaskUniversaResponse{
		ID:      taskID.String(),
		Message: "Задача удалена",
	})
}

// GetTasksByUserId godoc
// @Summary Получение задач по ID пользователя
// @Description Получение списка задач, назначенных на конкретного пользователя, с пагинацией (deleted = false)
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.UserTasksResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров или некорректный ID пользователя"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/user/{id}/{page}/{pagesize} [get]
func (tc *TaskController) GetTasksByUserId(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор пользователя"})
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	tasks, totalCount, err := tc.taskService.GetTasksByUserId(userID, page, pageSize)
	if err != nil {
		log.Printf("service error (get tasks by user id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задач"})
	}

	// === МАППИНГ В TaskFull ===
	tasksResp := make([]response.TaskFull, len(tasks))
	for i, task := range tasks {
		tasksResp[i] = response.TaskFull{
			ID:            task.ID.String(),
			Name:          utils.GetString(task.Name),
			Description:   utils.GetString(task.Description),
			Priority:      utils.GetInt16(task.Priority),
			Deadline:      utils.GetTime(task.Deadline),
			StartDate:     utils.GetTime(task.StartDate),
			TimeSpent:     utils.GetString(task.TimeSpent),
			GitlabIssueID: utils.GetInt(task.GitlabIssueID),
			Category:      utils.GetInt8(task.Category),
			Deleted:       utils.GetBool(task.Deleted),
			CreatedAt:     utils.GetTime(task.CreatedAt),
			UpdatedAt:     utils.GetTime(task.UpdatedAt),
			Status: response.StatusFull{
				ID:     task.Status.ID.String(),
				Name:   utils.GetString(task.Status.Name),
				Key:    utils.GetString(task.Status.Key),
				Color:  utils.GetString(task.Status.Color),
				IsOpen: utils.GetBool(task.Status.IsOpen),
				Board: response.BoardRef{
					ID:        task.Status.Board.ID.String(),
					Name:      utils.GetString(task.Status.Board.Name),
					ProjectID: task.Status.Board.ProjectID.String(),
				},
			},
			CreatedByUser: response.UserFull{
				ID:        task.CreatedByUser.ID.String(),
				FirstName: task.CreatedByUser.FirstName,
				LastName:  task.CreatedByUser.LastName,
				Email:     task.CreatedByUser.Email,
			},
			AssignedToUser: response.UserFull{
				ID:        task.AssignedToUser.ID.String(),
				FirstName: task.AssignedToUser.FirstName,
				LastName:  task.AssignedToUser.LastName,
				Email:     task.AssignedToUser.Email,
			},
		}
	}

	resp := response.UserTasksResponse{
		Tasks:      tasksResp,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	return c.JSON(http.StatusOK, resp)
}

// GetActiveTasksByUserId godoc
// @Summary Получение активных задач по ID пользователя (исключая статус Done)
// @Description Получение списка задач, назначенных на пользователя, где статус != 'done' и deleted = false
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID пользователя"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.UserTasksResponse "Список активных задач успешно получен"
// @Failure 400 {object} map[string]string "Ошибка при парсинге параметров или некорректный ID пользователя"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/user/{id}/{page}/{pagesize}/active [get]
func (tc *TaskController) GetActiveTasksByUserId(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный идентификатор пользователя"})
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	tasks, totalCount, err := tc.taskService.GetActiveTasksByUserId(userID, page, pageSize)
	if err != nil {
		log.Printf("service error (get active tasks by user id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении активных задач"})
	}

	tasksResp := make([]response.TaskFull, len(tasks))
	for i, task := range tasks {
		tasksResp[i] = response.TaskFull{
			ID:            task.ID.String(),
			Name:          utils.GetString(task.Name),
			Description:   utils.GetString(task.Description),
			Priority:      utils.GetInt16(task.Priority),
			Deadline:      utils.GetTime(task.Deadline),
			StartDate:     utils.GetTime(task.StartDate),
			TimeSpent:     utils.GetString(task.TimeSpent),
			GitlabIssueID: utils.GetInt(task.GitlabIssueID),
			Category:      utils.GetInt8(task.Category),
			Deleted:       utils.GetBool(task.Deleted),
			CreatedAt:     utils.GetTime(task.CreatedAt),
			UpdatedAt:     utils.GetTime(task.UpdatedAt),
			Status: response.StatusFull{
				ID:     task.Status.ID.String(),
				Name:   utils.GetString(task.Status.Name),
				Key:    utils.GetString(task.Status.Key),
				Color:  utils.GetString(task.Status.Color),
				IsOpen: utils.GetBool(task.Status.IsOpen),
				Board: response.BoardRef{
					ID:        task.Status.Board.ID.String(),
					Name:      utils.GetString(task.Status.Board.Name),
					ProjectID: task.Status.Board.ProjectID.String(),
				},
			},
			CreatedByUser: response.UserFull{
				ID:        task.CreatedByUser.ID.String(),
				FirstName: task.CreatedByUser.FirstName,
				LastName:  task.CreatedByUser.LastName,
				Email:     task.CreatedByUser.Email,
			},
			AssignedToUser: response.UserFull{
				ID:        task.AssignedToUser.ID.String(),
				FirstName: task.AssignedToUser.FirstName,
				LastName:  task.AssignedToUser.LastName,
				Email:     task.AssignedToUser.Email,
			},
		}
	}

	resp := response.UserTasksResponse{
		Tasks:      tasksResp,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	return c.JSON(http.StatusOK, resp)
}

// TaskMoveFunc godoc
// @Summary Переместить задачу в другой статус (столбец)
// @Description Перемещает задачу в указанный статус на той же доске
// @Tags Tasks
// @Accept json
// @Produce json
// @Param request body request.MoveTaskToAnotherStatus true "Данные для перемещения задачи"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Задача успешно перемещена"
// @Failure 400 {object} map[string]string "Некорректные данные запроса"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Задача или статус не найдены"
// @Failure 409 {object} map[string]string "Нельзя переместить задачу в статус с другой доски"
// @Failure 500 {object} map[string]string "Ошибка сервера при перемещении задачи"
// @Router /task/move [post]
func (tc *TaskController) TaskMoveFunc(c echo.Context) error {
	var req request.MoveTaskToAnotherStatus
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
	}

	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID задачи"})
	}

	toStatusID, err := uuid.Parse(req.ToStatusID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID статуса"})
	}

	statuses, err := tc.taskService.TaskMoveFunc(taskID, toStatusID)
	if err != nil {
		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}
		if err.Error() == "status not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Целевой статус не найден"})
		}
		if err.Error() == "different board" {
			return c.JSON(http.StatusConflict, map[string]string{"error": "Нельзя переместить задачу в статус с другой доски"})
		}
		log.Printf("service error (task move): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при обновлении статуса задачи"})
	}

	// Формируем ответ
	resp := response.StatusByBoardIdResponse{
		BoardId:  statuses[0].BoardID.String(),
		Statuses: make([]response.StatusResponse, 0, len(statuses)),
	}

	for _, status := range statuses {
		tasks := make([]response.TaskShort, 0, len(status.Tasks))
		for _, task := range status.Tasks {
			tasks = append(tasks, response.TaskShort{
				ID:          task.ID.String(),
				Name:        utils.GetString(task.Name),
				StatusID:    task.StatusID.String(),
				Priority:    utils.GetInt16(task.Priority),
				CreatedAt:   utils.GetTime(task.CreatedAt),
				UpdatedAt:   utils.GetTime(task.UpdatedAt),
				StartDate:   utils.GetTime(task.StartDate),
				Deadline:    utils.GetTime(task.Deadline),
			})
		}

		resp.Statuses = append(resp.Statuses, response.StatusResponse{
			ID:        status.ID.String(),
			Key:       utils.GetString(status.Key),
			Name:      utils.GetString(status.Name),
			Color:     utils.GetString(status.Color),
			Order:     utils.GetInt(status.SortOrder), // ← не забудь!
			IsDefault: utils.GetBool(status.IsDefault),
			IsActive:  utils.GetBool(status.IsActive),
			IsOpen:    utils.GetBool(status.IsOpen),
			CreatedAt: utils.GetTime(status.CreatedAt),
			UpdatedAt: utils.GetTime(status.UpdatedAt),
			Tasks:     tasks, // ← задачи внутри статуса
		})
	}

	return c.JSON(http.StatusOK, resp)
}

// GetUserProjectTasks godoc
// @Summary Получение задач пользователя в рамках проекта
// @Description Получает список задач, назначенных на указанного пользователя и принадлежащих заданному проекту, с пагинацией
// @Tags Tasks
// @Accept json
// @Produce json
// @Param user_id path string true "ID пользователя"
// @Param project_id path string true "ID проекта"
// @Param page path int true "Номер страницы"
// @Param pagesize path int true "Размер страницы"
// @Security BearerAuth
// @Success 200 {object} response.UserProjectTasksResponse "Список задач успешно получен"
// @Failure 400 {object} map[string]string "Некорректный ID пользователя или проекта"
// @Failure 401 {object} map[string]string "Нет или неверный токен"
// @Failure 404 {object} map[string]string "Пользователь или проект не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера при получении задач"
// @Router /task/user/{user_id}/project/{project_id}/{page}/{pagesize} [get]
func (tc *TaskController) GetUserProjectTasks(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID пользователя"})
	}

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректный ID проекта"})
	}

	page, err := strconv.Atoi(c.Param("page"))
	if err != nil || page <= 0 {
		page = 1
	}
	pageSize, err := strconv.Atoi(c.Param("pagesize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// Получаем задачи (уже отфильтрованы на уровне SQL)
	tasks, totalCount, err := tc.taskService.GetTasksByUserIDAndProjectID(userID, projectID, page, pageSize)
	if err != nil {
		log.Printf("service error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задач"})
	}

	// Получаем данные пользователя и проекта для верхнего уровня
	user, err := tc.userService.GetUserById(userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Пользователь не найден"})
	}
	project, err := tc.projectService.GetProjectByID(projectID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Проект не найден"})
	}

	userResp := response.UserFull{
		ID:        user.ID.String(),
		FirstName: utils.GetString(&user.FirstName),
		LastName:  utils.GetString(&user.LastName),
		Email:     user.Email,
	}

	projectResp := response.ProjectShort{
		ID:          project.ID.String(),
		Name:        utils.GetString(project.Name),
		Description: utils.GetString(project.Description),
		GitlabURL:   utils.GetString(project.GitlabURL),
		CreatedBy:   utils.GetUUIDString(project.CreatedBy),
	}

	tasksResp := make([]response.TaskFull, len(tasks))
	for i, task := range tasks {
		tasksResp[i] = response.TaskFull{
			ID:            task.ID.String(),
			Name:          utils.GetString(task.Name),
			Description:   utils.GetString(task.Description),
			Priority:      utils.GetInt16(task.Priority),
			Deadline:      utils.GetTime(task.Deadline),
			StartDate:     utils.GetTime(task.StartDate),
			TimeSpent:     utils.GetString(task.TimeSpent),
			GitlabIssueID: utils.GetInt(task.GitlabIssueID),
			Category:      utils.GetInt8(task.Category),
			Deleted:       utils.GetBool(task.Deleted),
			CreatedAt:     utils.GetTime(task.CreatedAt),
			UpdatedAt:     utils.GetTime(task.UpdatedAt),
			Status: response.StatusFull{
				ID:     task.Status.ID.String(),
				Name:   utils.GetString(task.Status.Name),
				Key:    utils.GetString(task.Status.Key),
				Color:  utils.GetString(task.Status.Color),
				IsOpen: utils.GetBool(task.Status.IsOpen),
				Board: response.BoardRef{
					ID:        task.Status.Board.ID.String(),
					Name:      utils.GetString(task.Status.Board.Name),
					ProjectID: task.Status.Board.ProjectID.String(),
				},
			},
			CreatedByUser: response.UserFull{
				ID:        task.CreatedByUser.ID.String(),
				FirstName: task.CreatedByUser.FirstName,
				LastName:  task.CreatedByUser.LastName,
				Email:     task.CreatedByUser.Email,
			},
			AssignedToUser: response.UserFull{
				ID:        task.AssignedToUser.ID.String(),
				FirstName: task.AssignedToUser.FirstName,
				LastName:  task.AssignedToUser.LastName,
				Email:     task.AssignedToUser.Email,
			},
		}
	}

	resp := response.UserProjectTasksResponse{
		User:       userResp,
		Project:    projectResp,
		Tasks:      tasksResp,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}

	return c.JSON(http.StatusOK, resp)
}

// ImproveTaskReport godoc
// @Summary Улучшить отчет по задаче с помощью LLM
// @Description Принимает текст пользователя и улучшает его на основе описания задачи через LLM
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Param request body request.ImproveReportRequest true "Данные для улучшения отчета"
// @Security BearerAuth
// @Success 200 {object} response.ImprovedReportResponse "Улучшенный отчет"
// @Failure 400 {object} map[string]string "Некорректный ID задачи или данные запроса"
// @Failure 404 {object} map[string]string "Задача не найдена"
// @Failure 500 {object} map[string]string "Ошибка при обработке LLM"
// @Router /task/{id}/improve-report [post]
func (tc *TaskController) ImproveTaskReport(c echo.Context) error {
	taskId := c.Param("id")
	taskID, err := uuid.Parse(taskId)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Некорректный идентификатор задачи",
		})
	}

	// Получаем задачу из БД
	task, err := tc.taskService.GetTaskByID(taskID)
	if err != nil {
		if err.Error() == "task not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}
		log.Printf("service error (get task by id): %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении задачи"})
	}

	// Парсим запрос
	var req request.ImproveReportRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Неверный формат запроса",
		})
	}

	if req.UserText == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Текст пользователя не может быть пустым",
		})
	}

	// Получаем описание задачи (может быть nil)
	taskDescription := ""
	if task.Description != nil {
		taskDescription = *task.Description
	}

	// Если у задачи нет описания, используем название
	if taskDescription == "" && task.Name != nil {
		taskDescription = *task.Name
	}

	// Вызываем gRPC сервис
	improvedText, err := tc.llmClient.ProcessTaskWithLLM(c.Request().Context(), taskDescription, req.UserText, taskId)
	if err != nil {
		log.Printf("gRPC error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Ошибка при обработке текста LLM",
		})
	}

	return c.JSON(http.StatusOK, response.ImprovedReportResponse{
		TaskID:          taskID.String(),
		OriginalText:    req.UserText,
		ImprovedText:    improvedText,
		TaskTitle:       utils.GetString(task.Name),
		TaskDescription: taskDescription,
	})
}

// GetTaskBoardAndProject godoc
// @Summary Получение доски и проекта задачи
// @Description Возвращает ID доски и ID проекта, к которым принадлежит задача
// @Tags Tasks
// @Accept json
// @Produce json
// @Param id path string true "ID задачи"
// @Security BearerAuth
// @Success 200 {object} interface{} "Информация о доске и проекте"
// @Failure 400 {object} map[string]string "Некорректный ID задачи"
// @Failure 404 {object} map[string]string "Задача, статус или доска не найдены"
// @Failure 500 {object} map[string]string "Ошибка сервера"
// @Router /task/board-project/{id} [get]
func (tc *TaskController) GetTaskBoardAndProject(c echo.Context) error {
    taskID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Некорректный идентификатор задачи",
        })
    }
    
   	boardId, projectId , err := tc.taskService.GetTaskBoardAndProjectIDs(taskID)

    if err != nil {
        if err.Error() == "task not found" {
            return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
        }
        if err.Error() == "status not found" {
            return c.JSON(http.StatusNotFound, map[string]string{"error": "Статус задачи не найден"})
        }
        if err.Error() == "board not found" {
            return c.JSON(http.StatusNotFound, map[string]string{"error": "Доска не найдена"})
        }
        if err.Error() == "project not found" {
            return c.JSON(http.StatusNotFound, map[string]string{"error": "Проект не найден"})
        }
        log.Printf("service error (get task board and project): %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка при получении данных задачи"})
    }

    return c.JSON(http.StatusOK, response.TaskBoardProjectResponse{
		TaskID: taskID.String(),
		BoardID: boardId.String(),
		ProjectID: projectId.String(),
	})
}