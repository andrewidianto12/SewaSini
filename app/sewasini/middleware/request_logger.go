package middleware

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const maxLoggedResponseLen = 800

type responseBodyWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
}

func (w *responseBodyWriter) Write(b []byte) (int, error) {
	_, _ = w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func RequestHitLogger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			startedAt := time.Now()

			respBody := new(bytes.Buffer)
			originalWriter := c.Response().Writer
			c.Response().Writer = &responseBodyWriter{ResponseWriter: originalWriter, body: respBody}

			err := next(c)
			c.Response().Writer = originalWriter

			route := c.Path()
			if route == "" {
				route = c.Request().URL.Path
			}

			userID, _ := c.Get(ContextUserIDKey).(string)
			if strings.TrimSpace(userID) == "" {
				userID = "-"
			}

			responseText := strings.TrimSpace(respBody.String())
			responseText = strings.ReplaceAll(responseText, "\n", "")
			if len(responseText) > maxLoggedResponseLen {
				responseText = responseText[:maxLoggedResponseLen] + "...(truncated)"
			}
			if responseText == "" {
				responseText = "-"
			}

			latencyMs := time.Since(startedAt).Milliseconds()
			log.Printf(
				"[API_HIT] method=%s route=%s uri=%s status=%d latency_ms=%d ip=%s user_id=%s response=%s",
				c.Request().Method,
				route,
				c.Request().RequestURI,
				c.Response().Status,
				latencyMs,
				c.RealIP(),
				userID,
				responseText,
			)

			return err
		}
	}
}
