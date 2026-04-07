package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"sewasini/database"
	"sewasini/util"
)

const ContextUserIDKey = "user_id"
const ContextUserRoleKey = "user_role"

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

func AdminOnly() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID, ok := c.Get(ContextUserIDKey).(string)
			if !ok || strings.TrimSpace(userID) == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
			}

			var role string
			err := database.DB.QueryRowContext(
				c.Request().Context(),
				`SELECT role FROM users WHERE id::text = $1`,
				userID,
			).Scan(&role)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return c.JSON(http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
				}
				return c.JSON(http.StatusInternalServerError, map[string]string{"message": "internal server error"})
			}

			if !strings.EqualFold(strings.TrimSpace(role), "admin") {
				return c.JSON(http.StatusForbidden, map[string]string{"message": "forbidden: admin only"})
			}

			c.Set(ContextUserRoleKey, role)
			return next(c)
		}
	}
}
