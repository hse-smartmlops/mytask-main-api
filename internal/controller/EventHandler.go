package controller

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
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

    // Log the request URI
    log.Print("Request URI: ", c.Request().RequestURI)

    // Print the whole payload
    log.Printf("Payload: %+v", payload)

    // Extract the "type" field safely
    if eventType, ok := payload["type"].(string); ok {
        log.Printf("Event type: %s", eventType)
    } else {
        log.Printf("Event type not found or not a string")
    }

    return c.JSON(http.StatusOK, map[string]string{"status": "received"})
}