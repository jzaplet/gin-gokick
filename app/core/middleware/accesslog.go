package middleware

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
	"gokick/app/core/reporting"
)

func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		warnings := reporting.Warnings(c)
		level := slog.LevelInfo

		switch {
		case status >= http.StatusInternalServerError:
			level = slog.LevelError
		case len(warnings) > 0:
			level = slog.LevelWarn
		}

		attrs := []slog.Attr{
			slog.String(logging.KeyMethod, c.Request.Method),
			slog.String(logging.KeyPath, loggedPath(c)),
			slog.Int(logging.KeyStatus, status),
			slog.String(logging.KeyIP, c.ClientIP()),
			slog.Int(logging.KeyBytes, max(c.Writer.Size(), 0)),
			logging.DurationMs(time.Since(start)),
		}
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String(logging.KeyError, strings.Join(c.Errors.Errors(), "; ")))
		}

		if len(warnings) > 0 {
			attrs = append(attrs, slog.String(logging.KeyWarning, joined(warnings)))
		}

		logger.LogAttrs(c, level, "http request", attrs...)
	}
}

func loggedPath(c *gin.Context) string {
	if route := c.FullPath(); route != "" {
		return route
	}

	return c.Request.URL.Path
}

func joined(errs []error) string {
	texts := make([]string, 0, len(errs))
	for _, err := range errs {
		texts = append(texts, err.Error())
	}

	return strings.Join(texts, "; ")
}
