package trace

import (
	"encoding/hex"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
)

func Propagation(c *gin.Context) sentry.PropagationContext {
	header := c.GetHeader("sentry-trace")

	pc, err := sentry.PropagationContextFromHeaders(header, c.GetHeader("baggage"))
	if err != nil {
		pc, _ = sentry.PropagationContextFromHeaders(header, "")
	}

	id := logging.TraceID(c)
	if pc.TraceID.String() == id {
		return pc
	}

	fresh := sentry.NewPropagationContext()
	if decoded, err := hex.DecodeString(id); err == nil && len(decoded) == len(fresh.TraceID) {
		fresh.TraceID = sentry.TraceID(decoded)
	}

	return fresh
}
