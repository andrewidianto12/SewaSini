package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"sewasini/util"
)

const ContextUserIDKey = "user_id"

func BearerAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authorizationHeader := c.Request().Header.Get("Authorization")
			if authorizationHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "missing authorization header"})
			}

			parts := strings.SplitN(authorizationHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid authorization header"})
			}

			userID, err := util.ParseToken(strings.TrimSpace(parts[1]))
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "invalid or expired token"})
			}

			c.Set(ContextUserIDKey, userID)
			return next(c)
		}
	}
}
