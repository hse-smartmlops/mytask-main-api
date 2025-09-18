package controller

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// KeycloakAuthMiddleware validates Authorization header using Keycloak introspection
// and attempts token exchange if necessary. On success it stores the active token
// in the context under key "auth_token".
func KeycloakAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Allow public routes and preflight: don't enforce auth for them
        p := c.Path()
        if c.Request().Method == http.MethodOptions || strings.HasPrefix(p, "/swagger") || strings.HasPrefix(p, "/auth") || strings.HasPrefix(p, "/event") {
            return next(c)
        }
        authHeader := c.Request().Header.Get("Authorization")

        if authHeader == ""{
            log.Print("No authorization header")
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "missing authorization header",
            })
        }

        token := strings.TrimPrefix(authHeader, "Bearer ")
        if token == authHeader{
            return c.JSON(http.StatusBadRequest, map[string]string{
                "error": "invalid authorization header format",
            })
        }

        ctx := context.Background()

        // Try backend client introspection
        retrResult, err := keycloakClient.RetrospectToken(ctx, token, clientID, clientSecret, realm)
        if err != nil || retrResult == nil || !*retrResult.Active {
            // Attempt token exchange for Flutter token
            exchanged, exErr := ExchangeToken(ctx, token)
            if exErr != nil {
                log.Printf("Token exchange failed: %v, retrocpect error: %v", exErr, err)
                return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token invalid"})
            }
            token = exchanged.AccessToken

            // Validate again with backend client
            retrResult, err = keycloakClient.RetrospectToken(ctx, token, clientID, clientSecret, realm)
            if err != nil || retrResult == nil || !*retrResult.Active {
                log.Printf("Token inactive after exchange: %v", err)
                return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token inactive"})
            }
        }

        // Save token for handlers
        c.Set("auth_token", token)

        return next(c)
    }
}

// authorize can be used by handlers that were not yet refactored away from inline checks.
// It prefers an already-stored token (set by middleware) but will validate header if missing.
func authorize(c echo.Context) error {
    if t := c.Get("auth_token"); t != nil {
        if _, ok := t.(string); ok {
            return nil
        }
    }

    // Fallback: validate header (same logic as middleware)
    authHeader := c.Request().Header.Get("Authorization")

    if authHeader == ""{
        log.Print("No authorization header")
        return c.JSON(http.StatusUnauthorized, map[string]string{
            "error": "missing authorization header",
        })
    }

    token := strings.TrimPrefix(authHeader, "Bearer ")
    if token == authHeader{
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "invalid authorization header format",
        })
    }

    ctx := context.Background()

    // Try backend client introspection
    retrResult, err := keycloakClient.RetrospectToken(ctx, token, clientID, clientSecret, realm)
    if err != nil || retrResult == nil || !*retrResult.Active {
        // Attempt token exchange for Flutter token
        exchanged, exErr := ExchangeToken(ctx, token)
        if exErr != nil {
            log.Printf("Token exchange failed: %v, retrocpect error: %v", exErr, err)
            return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token invalid"})
        }
        token = exchanged.AccessToken

        // Validate again with backend client
        retrResult, err = keycloakClient.RetrospectToken(ctx, token, clientID, clientSecret, realm)
        if err != nil || retrResult == nil || !*retrResult.Active {
            log.Printf("Token inactive after exchange: %v", err)
            return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token inactive"})
        }
    }

    c.Set("auth_token", token)
    return nil
}
