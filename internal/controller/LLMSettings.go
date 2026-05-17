package controller

import (
	models "emplacc-api/internal/domain"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type LLMSettingsController struct {
	db *gorm.DB
}

func NewLLMSettingsController(db *gorm.DB) *LLMSettingsController {
	return &LLMSettingsController{db: db}
}

func RegisterLLMSettingsRoutes(e *echo.Echo, db *gorm.DB, adminMw echo.MiddlewareFunc) {
	c := NewLLMSettingsController(db)
	g := e.Group("/admin/llm-settings")
	g.GET("", c.Get, adminMw)
	g.PATCH("", c.Update, adminMw)
}

type LLMSettingsResponse struct {
	WebUIURL     string `json:"webui_url"`
	WebUIModel   string `json:"webui_model"`
	SystemPrompt string `json:"system_prompt"`
	HasToken     bool   `json:"has_token"` // токен не возвращаем, только факт наличия
	UpdatedAt    string `json:"updated_at,omitempty"`
	UpdatedBy    string `json:"updated_by,omitempty"`
}

type LLMSettingsUpdateRequest struct {
	WebUIURL     *string `json:"webui_url"`
	WebUIToken   *string `json:"webui_token"`
	WebUIModel   *string `json:"webui_model"`
	SystemPrompt *string `json:"system_prompt"`
}

// Get godoc
// @Summary Получить настройки LLM
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Router /admin/llm-settings [get]
func (c *LLMSettingsController) Get(ctx echo.Context) error {
	var s models.LLMSettings
	if err := c.db.First(&s).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Возвращаем дефолтные значения из env
			return ctx.JSON(http.StatusOK, LLMSettingsResponse{
				WebUIURL:     "",
				WebUIModel:   "",
				SystemPrompt: "",
				HasToken:     false,
			})
		}
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to load settings"})
	}

	return ctx.JSON(http.StatusOK, LLMSettingsResponse{
		WebUIURL:     s.WebUIURL,
		WebUIModel:   s.WebUIModel,
		SystemPrompt: s.SystemPrompt,
		HasToken:     s.WebUIToken != "",
		UpdatedAt:    s.UpdatedAt.Format(time.RFC3339),
		UpdatedBy:    s.UpdatedBy,
	})
}

// Update godoc
// @Summary Обновить настройки LLM
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Router /admin/llm-settings [patch]
func (c *LLMSettingsController) Update(ctx echo.Context) error {
	var req LLMSettingsUpdateRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Получаем кто делает запрос
	updatedBy := ""
	if token, _ := ctx.Get("auth_token").(string); token != "" {
		if sub, err := extractSubFromJWT(token); err == nil {
			updatedBy = sub.String()
		}
	}

	var s models.LLMSettings
	c.db.First(&s) // игнорируем ошибку — создадим новую запись если нет

	if req.WebUIURL != nil {
		s.WebUIURL = *req.WebUIURL
	}
	if req.WebUIToken != nil && *req.WebUIToken != "" {
		s.WebUIToken = *req.WebUIToken
	}
	if req.WebUIModel != nil {
		s.WebUIModel = *req.WebUIModel
	}
	if req.SystemPrompt != nil {
		s.SystemPrompt = *req.SystemPrompt
	}
	s.UpdatedAt = time.Now()
	s.UpdatedBy = updatedBy

	if s.ID == 0 {
		if err := c.db.Create(&s).Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create settings"})
		}
	} else {
		if err := c.db.Save(&s).Error; err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save settings"})
		}
	}

	return ctx.JSON(http.StatusOK, map[string]string{"message": "settings updated"})
}
