// status_test.go
package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"emplacc-api/internal/controller"
	models "emplacc-api/internal/domain"
	"emplacc-api/internal/dto/request"
)

func TestStatus_FullCRUD(t *testing.T) {
	testDB := setupTestDB(t)

	// Подменяем глобальную БД и авторизацию
	controller.DBConn = testDB
	defer func() { controller.DBConn = origDBConn }()

	controller.Authorize = func(c echo.Context) error { return nil }
	defer func() { controller.Authorize = origAuthorize }()

	e := echo.New()

	// Создаём тестовые сущности
	userID := createTestUser(t, testDB, "user@example.com")
	projectID := createTestProject(t, testDB, "Test Project")
	boardID := createTestBoard(t, testDB, projectID, "Test Board")
	taskID := createTestTask(t, testDB, projectID, userID, "Test Task")

	var statusID uuid.UUID

	// === 1. CreateStatus ===
	t.Run("createStatus", func(t *testing.T) {
		reqBody := request.CreateStatusRequest{
			Name:      "Test Status",
			Color:     "#FF0000",
			IsDefault: false,
			IsActive:  true,
			IsOpen:    true,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/status", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.CreateStatus(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Contains(t, resp, "id")
		assert.NotEmpty(t, resp["id"])
		assert.Equal(t, "Статус успешно создан", resp["message"])

		// Сохраняем ID статуса
		statusID = uuid.MustParse(resp["id"].(string))
	})

	// === 2. GetStatusByID ===
	t.Run("getStatusById", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/status/%s", statusID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(statusID.String())

		err := controller.GetStatusByID(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, statusID.String(), resp["id"])
		assert.Equal(t, "Test Status", resp["name"])
		assert.Equal(t, "#FF0000", resp["color"])
	})

	// === 3. UpdateStatus ===
	t.Run("updateStatus", func(t *testing.T) {
		newName := "Updated Status Name"
		newColor := "#00FF00"

		reqBody := request.UpdateStatusRequest{
			Name:  &newName,
			Color: &newColor,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/status/%s", statusID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(statusID.String())

		err := controller.UpdateStatus(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "Статус успешно обновлён", resp["message"])
	})

	// === 4. GetAllStatuses ===
	t.Run("getAllStatuses", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/status/all/1/10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("page", "pagesize")
		c.SetParamValues("1", "10")

		err := controller.GetAllStatuses(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, float64(1), resp["page"])
		assert.Equal(t, float64(10), resp["page_size"])

		statuses, ok := resp["statuses"].([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(statuses), 0)
	})

	// === 5. AddStatusToTask ===
	t.Run("addStatusToTask", func(t *testing.T) {
		reqBody := request.AddStatusToTaskRequest{
			TaskId:   taskID.String(),
			StatusId: statusID.String(),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/status/add-to-task", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.AddStatusToTask(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, taskID.String(), resp["task_id"])
		assert.Equal(t, statusID.String(), resp["status_id"])
		assert.Equal(t, "Статус успешно добавлен к задаче", resp["message"])
	})

	// === 6. GetStatusesByTaskId ===
	t.Run("getStatusesByTaskId", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/status/task/%s", taskID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("task_id")
		c.SetParamValues(taskID.String())

		err := controller.GetStatusesByTaskId(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, taskID.String(), resp["task_id"])

		statuses, ok := resp["statuses"].([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(statuses), 0)
	})

	// === 7. AddStatusToBoard ===
	t.Run("addStatusToBoard", func(t *testing.T) {
		reqBody := request.AddStatusToBoardRequest{
			BoardId:  boardID.String(),
			StatusId: statusID.String(),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/status/add-to-board", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.AddStatusToBoard(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, boardID.String(), resp["board_id"])
		assert.Equal(t, statusID.String(), resp["status_id"])
		assert.Equal(t, "Статус успешно добавлен к доске", resp["message"])
	})

	// === 8. GetStatusesByBoardId ===
	t.Run("getStatusesByBoardId", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/status/project/%s", boardID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("board_id")
		c.SetParamValues(boardID.String())

		err := controller.GetStatusesByBoardId(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, boardID.String(), resp["project_id"]) // JSON-тег: project_id

		statuses, ok := resp["statuses"].([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(statuses), 0)
	})

	// === 9. DeleteStatusFromTask ===
	t.Run("deleteStatusFromTask", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/status/delete-from-task/%s/%s", taskID, statusID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("task_id", "status_id")
		c.SetParamValues(taskID.String(), statusID.String())

		err := controller.DeleteStatusFromTask(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, statusID.String(), resp["id"])
		assert.Equal(t, "Статус задачи успешно удалён", resp["message"])

		// Проверим, что связь действительно удалена
		var deleted models.StatusTask
		require.NoError(t, testDB.Unscoped().Where("status_id = ? AND task_id = ?", statusID, taskID).First(&deleted).Error)
		assert.True(t, *deleted.Deleted)
	})

	// === 10. DeleteStatusFromBoard ===
	t.Run("deleteStatusFromBoard", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/status/delete-from-board/%s/%s", boardID, statusID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("board_id", "status_id")
		c.SetParamValues(boardID.String(), statusID.String())

		err := controller.DeleteStatusFromBoard(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, statusID.String(), resp["id"])
		assert.Equal(t, "Статус доски успешно удалён", resp["message"])

		// Проверим, что связь действительно удалена
		var deleted models.StatusBoard
		require.NoError(t, testDB.Unscoped().Where("status_id = ? AND board_id = ?", statusID, boardID).First(&deleted).Error)
		assert.True(t, *deleted.Deleted)
	})

	// === 11. DeleteStatus ===
	t.Run("deleteStatus", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/status/%s", statusID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(statusID.String())

		err := controller.DeleteStatus(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, statusID.String(), resp["id"])
		assert.Equal(t, "Статус успешно удалён", resp["message"])

		// Проверим, что статус действительно удалён
		var deleted models.Status
		require.NoError(t, testDB.Unscoped().First(&deleted, "id = ?", statusID).Error)
		assert.True(t, *deleted.Deleted)
	})
}