package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
)

var sentryTrace = regexp.MustCompile(`^([0-9a-f]{32})-[0-9a-f]{16}(-[01])?$`)

func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := traceID(c.GetHeader("sentry-trace"))
		c.Request = c.Request.WithContext(logging.WithTraceID(c.Request.Context(), id))
		c.Header("X-Trace-Id", id)
		c.Next()
	}
}

func traceID(sentryTraceHeader string) string {
	if match := sentryTrace.FindStringSubmatch(sentryTraceHeader); match != nil {
		return match[1]
	}

	id := make([]byte, 16)
	_, _ = rand.Read(id)

	return hex.EncodeToString(id)
}
