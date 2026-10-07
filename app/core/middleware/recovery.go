package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
	"gokick/app/core/reporting"
)

func Recovery(logger *slog.Logger, reporter *reporting.Reporter) gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		logger.LogAttrs(c, slog.LevelError, "panic recovered",
			slog.Any(logging.KeyPanic, recovered),
			slog.String(logging.KeyStack, string(debug.Stack())),
		)
		reporter.Recover(c, recovered)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
}
