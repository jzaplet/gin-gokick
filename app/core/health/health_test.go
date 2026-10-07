package health

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/testkit"
)

func TestFailedCheckHidesItsError(t *testing.T) {
	refused := func(context.Context) error { return errors.New("dial tcp 10.0.1.5:5432: connection refused") }
	database := Check{Name: "database", Probe: refused}
	code, body, reported := serve(t, time.Second, database, Check{Name: "cache", Probe: pass})

	if code != http.StatusServiceUnavailable || body != `{"status":"unhealthy","checks":{"cache":"healthy","database":"unhealthy"}}` {
		t.Errorf("%d %s", code, body)
	}

	if strings.Contains(reported, "database: dial tcp 10.0.1.5:5432") == false {
		t.Errorf("reported %q", reported)
	}
}

func TestSlowCheckFailsAtTheTimeout(t *testing.T) {
	hanging := func(ctx context.Context) error {
		<-ctx.Done()

		return ctx.Err()
	}
	started := time.Now()
	code, _, reported := serve(t, 50*time.Millisecond, Check{Name: "database", Probe: hanging})

	if code != http.StatusServiceUnavailable || strings.Contains(reported, context.DeadlineExceeded.Error()) == false || time.Since(started) > time.Second {
		t.Errorf("%d %q after %s", code, reported, time.Since(started))
	}
}

func pass(context.Context) error {
	return nil
}

func serve(t *testing.T, timeout time.Duration, checks ...Check) (code int, body, reported string) {
	t.Helper()

	engine := gin.New()
	engine.GET("/readyz", func(c *gin.Context) {
		c.Next()
		reported = c.Errors.String()
	}, Handler(timeout, checks...))
	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/readyz"))

	return res.Code, res.Body.String(), reported
}
