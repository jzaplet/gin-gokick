package middleware

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/testkit"
)

func TestRequestTimeoutCancelsTheContext(t *testing.T) {
	router := newTestRouter(t, testkit.DiscardLogs())
	router.Use(RequestTimeout(50 * time.Millisecond))
	var cancelled error

	router.GET("/slow", func(c *gin.Context) {
		<-c.Request.Context().Done()
		cancelled = c.Request.Context().Err()
		c.Status(http.StatusGatewayTimeout)
	})

	started := time.Now()

	res := testkit.Serve(router, testkit.Request(t, http.MethodGet, "/slow"))
	if errors.Is(cancelled, context.DeadlineExceeded) == false || res.Code != http.StatusGatewayTimeout || time.Since(started) > time.Second {
		t.Errorf("%v, status %d after %s", cancelled, res.Code, time.Since(started))
	}
}
