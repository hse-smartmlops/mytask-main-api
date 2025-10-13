package controller

import (
	"emplacc-api/internal/service"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

/*
type authController struct {
	authService service.AuthService
}*/

// KeycloakAuthMiddleware validates Authorization header using Keycloak introspection
// and attempts token exchange if necessary. On success it stores the active token
// in the context under key "auth_token".
func KeycloakAuthMiddleware(authService service.AuthService) echo.MiddlewareFunc {
	controller := NewAuthController(authService)
    return func(next echo.HandlerFunc) echo.HandlerFunc {
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

            err := controller.authService.ValidateTokenForMiddleware(token)
            if err != nil {
                log.Printf("Token validation failed: %v", err)
                return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token invalid"})
            }

            // Save token for handlers
            c.Set("auth_token", token)

            return next(c)
        }
    }
}

/*func Authorize(authService service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if t := c.Get("auth_token"); t != nil {
				if _, ok := t.(string); ok {
					return next(c)
				}
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

			err := authService.ValidateTokenForMiddleware(token)
			if err != nil {
				log.Printf("Token validation failed: %v", err)
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Token invalid"})
			}

			c.Set("auth_token", token)
			return next(c)
		}
	}
}*/