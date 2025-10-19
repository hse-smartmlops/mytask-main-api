package middleware

import (
	"net/http"
	"strings"

	"emplacc-api/api/v1/dto"
	"emplacc-api/internal/app"

	"github.com/labstack/echo/v4"
)

func KeycloakAuth(validator app.TokenValidator) echo.MiddlewareFunc {
	if validator == nil {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, dto.NewError("missing_authorization_header", "authorization header is required"))
			}

			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				return c.JSON(http.StatusBadRequest, dto.NewError("invalid_authorization_header", "authorization header must use Bearer scheme"))
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
			if token == "" {
				return c.JSON(http.StatusBadRequest, dto.NewError("invalid_authorization_header", "token is empty"))
			}

			if err := validator.ValidateToken(c.Request().Context(), token); err != nil {
				return c.JSON(http.StatusUnauthorized, dto.NewError("token_invalid", "token validation failed"))
			}

			return next(c)
		}
	}
}
