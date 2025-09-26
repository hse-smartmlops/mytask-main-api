// task_test.go
package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"emplacc-api/internal/controller"
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
)

func TestTask_FullCRUD(t *testing.T) {
	testDB := setupTestDB(t)

	// Подменяем глобальную БД и авторизацию
	controller.DBConn = testDB
	defer func() { controller.DBConn = origDBConn }()

	controller.Authorize = func(c echo.Context) error { return nil }
	defer func() { controller.Authorize = origAuthorize }()

	e := echo.New()

	// Создаём тестовые сущности
	userID := createTestUser(t, testDB, "user@example.com")
	creatorID := createTestUser(t, testDB, "creator@example.com")
	projectID := createTestProject(t, testDB, "Test Project")

	var taskID uuid.UUID

	// === 1. CreateTask ===
	t.Run("createTask", func(t *testing.T) {
		name := "Test Task"
		description := "This is a test task"
		priority := int16(5)
		category := int8(1)
		startDate := time.Now()
		deadline := time.Now().Add(24 * time.Hour)
		creatorId := creatorID.String()
		userId := userID.String()

		reqBody := request.TaskCreateRequest{
			ProjectID:     projectID.String(),
			Name:          &name,
			Description:   &description,
			Priority:      &priority,
			CreatorID:     &creatorId,
			AssignedTo:    &userId,
			StartDate:     &startDate,
			Deadline:      &deadline,
			Category:      &category,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/task", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.CreateTask(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Contains(t, resp, "id")
		assert.NotEmpty(t, resp["id"])
		assert.Equal(t, "Задача создана", resp["message"])

		// Сохраняем ID задачи
		taskID = uuid.MustParse(resp["id"].(string))
	})

	// === 2. GetTaskByID ===
	t.Run("getTaskById", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/task/%s", taskID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(taskID.String())

		err := controller.GetTaskByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, taskID.String(), resp["id"])
		assert.Equal(t, projectID.String(), resp["project_id"])
		assert.Equal(t, "Test Task", resp["name"])
		assert.Equal(t, "This is a test task", resp["description"])
		assert.Equal(t, float64(5), resp["priority"])
	})

	// === 3. GetTasksByProjectID ===
	t.Run("getTasksByProjectId", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/task/project/%s/1/10", projectID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("projectId", "page", "pagesize")
		c.SetParamValues(projectID.String(), "1", "10")

		err := controller.GetTasksByProjectID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		// Убираем проверку resp["project_id"] — его нет в TaskListResponse
		assert.Equal(t, float64(1), resp["page"])
		assert.Equal(t, float64(10), resp["page_size"])

		tasks, ok := resp["tasks"].([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(tasks), 0)

		// Проверим, что задача принадлежит проекту
		firstTask := tasks[0].(map[string]interface{})
		assert.Equal(t, projectID.String(), firstTask["project_id"])
	})

	// === 4. GetTasksByUserId ===
	t.Run("getTasksByUserId", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/task/user/%s/1/10", userID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id", "page", "pagesize")
		c.SetParamValues(userID.String(), "1", "10")

		err := controller.GetTasksByUserId(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		// Убираем проверку resp["user_id"] — его нет в TaskListResponse
		assert.Equal(t, float64(1), resp["page"])
		assert.Equal(t, float64(10), resp["page_size"])

		tasks, ok := resp["tasks"].([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(tasks), 0)

		// Проверим, что задача в списке (не проверяем assigned_to, потому что его нет в TaskShort)
		firstTask := tasks[0].(map[string]interface{})
		assert.Equal(t, taskID.String(), firstTask["id"])
	})
	
	// === 5. UpdateTask ===
	t.Run("updateTask", func(t *testing.T) {
		newName := "Updated Task Name"
		newDescription := "Updated task description"

		reqBody := request.TaskUpdateRequest{
			Name:        &newName,
			Description: &newDescription,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/task/%s", taskID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(taskID.String())

		err := controller.UpdateTask(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "Задача изменена", resp["message"])
	})

	// === 6. GetAllTasks ===
	t.Run("getAllTasks", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/task/all/1/10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("page", "pagesize")
		c.SetParamValues("1", "10")

		err := controller.GetAllTasks(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, float64(1), resp["page"])
		assert.Equal(t, float64(10), resp["page_size"])

		tasks, ok := resp["tasks"].([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(tasks), 0)
	})

	// === 7. DeleteTask ===
	t.Run("deleteTask", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/task/%s", taskID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(taskID.String())

		err := controller.DeleteTask(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, taskID.String(), resp["id"])
		assert.Equal(t, "Задача удалена", resp["message"])

		// Проверим, что задача действительно удалена
		var deleted models.Task
		require.NoError(t, testDB.Unscoped().First(&deleted, "id = ?", taskID).Error)
		assert.True(t, *deleted.Deleted)
	})
}