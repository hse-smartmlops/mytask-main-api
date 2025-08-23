package controller

import (
	models "emplacc-api/internal/domain"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func RegisterWebhookRoutes(e *echo.Echo) {
    eventGroup := e.Group("/event")
    eventGroup.POST("*", KeycloakEventHandler)
}

func KeycloakEventHandler(c echo.Context) error {
    var payload map[string]interface{}
    if err := c.Bind(&payload); err != nil {
        log.Printf("Parse error: %v", err)
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }

    // Вытаскиваем основные поля
    eventIDStr, _ := payload["id"].(string)
    userIDStr, _ := payload["userId"].(string)
    clientID, _ := payload["clientId"].(string)
    realmIDStr, _ := payload["realmId"].(string)
    ipAddress, _ := payload["ipAddress"].(string)
    resourcePath, _ := payload["resourcePath"].(string)
    eventType, _ := payload["type"].(string)
    timestampFloat, _ := payload["time"].(float64)

    eventID, _ := uuid.Parse(eventIDStr)
    var userID *uuid.UUID
    if userIDStr != "" {
        u, _ := uuid.Parse(userIDStr)
        userID = &u
    }
    var realmID *uuid.UUID
    if realmIDStr != "" {
        r, _ := uuid.Parse(realmIDStr)
        realmID = &r
    }
    occurredAt := time.UnixMilli(int64(timestampFloat))

    // Создаем AuthEvent
    authEvent := models.AuthEvent{
        ID:           eventID,
        EventID:      eventID,
        EventType:    &eventType,
        UserID:       userID,
        ClientID:     &clientID,
        RealmID:      realmID,
        IPAddress:    &ipAddress,
        ResourcePath: &resourcePath,
        OccurredAt:   &occurredAt,
    }

    // Обрабатываем поля в зависимости от типа события
    switch eventType {
    case "LOGIN", "LOGIN_ERROR", "LOGOUT", "LOGOUT_ERROR", "REFRESH_TOKEN", "REFRESH_TOKEN_ERROR", "INTROSPECT_TOKEN", "INTROSPECT_TOKEN_ERROR", "USER_INFO_REQUEST", "USER_INFO_REQUEST_ERROR":
        // детали auth события
        if detailsRaw, ok := payload["details"].(map[string]interface{}); ok {
            detail := models.AuthEventDetail{
                EventID: authEvent.ID,
            }
            if v, ok := detailsRaw["auth_method"].(string); ok {
                detail.AuthMethod = &v
            }
            if v, ok := detailsRaw["client_auth_method"].(string); ok {
                detail.ClientAuthMethod = &v
            }
            if v, ok := detailsRaw["grant_type"].(string); ok {
                detail.GrantType = &v
            }
            if v, ok := detailsRaw["token_id"].(string); ok {
                tid, _ := uuid.Parse(v)
                detail.TokenID = &tid
            }
            if v, ok := detailsRaw["refresh_token_id"].(string); ok {
                rid, _ := uuid.Parse(v)
                detail.RefreshTokenID = &rid
            }
            if v, ok := detailsRaw["refresh_token_type"].(string); ok {
                detail.RefreshTokenType = &v
            }
            if v, ok := detailsRaw["updated_refresh_token_id"].(string); ok {
                urid, _ := uuid.Parse(v)
                detail.UpdatedRefreshTokenID = &urid
            }
            if v, ok := detailsRaw["scope"].(string); ok {
                detail.Scope = &v
            }
            if v, ok := detailsRaw["username"].(string); ok {
                detail.Username = &v
            }
            authEvent.Details = []models.AuthEventDetail{detail}
        }

    case "USER-CREATE", "USER-ACTION":
        // репрезентация пользователя
        if reprRaw, ok := payload["representation"].(map[string]interface{}); ok {
            userRepr := models.AuthEventUserRepresentation{
                EventID: authEvent.ID,
            }
            if v, ok := reprRaw["username"].(string); ok {
                userRepr.Username = &v
            }
            if v, ok := reprRaw["firstName"].(string); ok {
                userRepr.FirstName = &v
            }
            if v, ok := reprRaw["lastName"].(string); ok {
                userRepr.LastName = &v
            }
            if v, ok := reprRaw["email"].(string); ok {
                userRepr.Email = &v
            }
            if v, ok := reprRaw["enabled"].(bool); ok {
                userRepr.Enabled = &v
            }
            authEvent.Representation = []models.AuthEventUserRepresentation{userRepr}
        }
    }

    // Всегда сохраняем поле error, если есть
    if errMsg, ok := payload["error"].(string); ok {
        authEvent.Error = &errMsg
    }

    // Сохраняем в БД
    if err := dbConn.Session(&gorm.Session{}).Create(&authEvent).Error; err != nil {
        log.Printf("DB error: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "DB error"})
    }

    log.Printf("Event %s saved: type=%s user=%v", authEvent.ID, eventType, userID)
    return c.JSON(http.StatusOK, map[string]string{"status": "received"})
}
