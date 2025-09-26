// forum_message_test.go
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

func TestForumMessage_FullCRUD(t *testing.T) {
	testDB := setupTestDB(t)

	// Подменяем глобальную БД и авторизацию
	controller.DBConn = testDB
	defer func() { controller.DBConn = origDBConn }()

	controller.Authorize = func(c echo.Context) error { return nil }
	defer func() { controller.Authorize = origAuthorize }()

	e := echo.New()

	// Создаём тестовые сущности
	userID := createTestUser(t, testDB, "user@example.com")
	problemID := createTestProblem(t, testDB, "Test Problem", userID)

	var messageID uuid.UUID

	// === 1. CreateForumMessage ===
	t.Run("createForumMessage", func(t *testing.T) {
		description := []string{"This is a test message"}

		reqBody := request.CreateForumMessageRequest{
			ProblemID:   problemID.String(),
			Description: &description,
			CreatorID:   userID.String(),
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/forum-messages", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := controller.CreateForumMessage(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Contains(t, resp, "id")
		assert.NotEmpty(t, resp["id"])
		assert.Equal(t, "Сообщение форума создано", resp["message"])

		// Сохраняем ID сообщения
		messageID = uuid.MustParse(resp["id"].(string))
	})

	// === 2. GetForumMessageById ===
	t.Run("getForumMessageById", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/forum-messages/%s", messageID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(messageID.String())

		err := controller.GetForumMessageById(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, messageID.String(), resp["id"])
		assert.Equal(t, problemID.String(), resp["problem_id"])
		assert.Equal(t, userID.String(), resp["creator_id"])
	})

	// === 3. GetForumMessagesByProblemId ===
	t.Run("getForumMessagesByProblemId", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/forum-messages/problem/%s/1/10", problemID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id", "page", "pagesize")
		c.SetParamValues(problemID.String(), "1", "10")

		err := controller.GetForumMessagesByProblemId(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, problemID.String(), resp["problem_id"])
		assert.Equal(t, float64(1), resp["page"])
		assert.Equal(t, float64(10), resp["page_size"])

		messages, ok := resp["messages"].([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(messages), 0)
	})

	// === 4. UpdateForumMessage ===
	t.Run("updateForumMessage", func(t *testing.T) {
		newDescription := []string{"Updated message"}

		reqBody := request.UpdateForumMessageRequest{
			Description: &newDescription,
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/forum-messages/%s", messageID), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(messageID.String())

		err := controller.UpdateForumMessage(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, "Сообщение форума обновлено", resp["message"])
	})

	// === 5. GetAllForumMessages ===
	t.Run("getAllForumMessages", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/forum-messages/all/1/10", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("page", "pagesize")
		c.SetParamValues("1", "10")

		err := controller.GetAllForumMessages(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, float64(1), resp["page"])
		assert.Equal(t, float64(10), resp["page_size"])

		messages, ok := resp["messages"].([]interface{})
		assert.True(t, ok)
		assert.Greater(t, len(messages), 0)
	})

	// === 6. DeleteForumMessage ===
	t.Run("deleteForumMessage", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/forum-messages/%s", messageID), nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("id")
		c.SetParamValues(messageID.String())

		err := controller.DeleteForumMessage(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)

		var resp map[string]interface{}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
		assert.Equal(t, messageID.String(), resp["id"])
		assert.Equal(t, "Сообщение форума удалено", resp["message"])

		// Проверим, что сообщение действительно удалено
		var deleted models.ForumMessage
		require.NoError(t, testDB.Unscoped().First(&deleted, "id = ?", messageID).Error)
		assert.True(t, *deleted.Deleted)
	})
}