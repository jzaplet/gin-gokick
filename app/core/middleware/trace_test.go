package middleware

import (
	"net/http"
	"regexp"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
	"gokick/app/core/testkit"
)

func TestTraceAdoptsTheSentryTraceID(t *testing.T) {
	req := testkit.Request(t, http.MethodGet, "/trace")
	req.Header.Set("sentry-trace", "771a43a4192642f0b136d5159a501700-b7ad6b7169203331-1")

	if got := traceOf(t, req); got != "771a43a4192642f0b136d5159a501700" {
		t.Errorf("trace id %q", got)
	}
}

func TestTraceGeneratesAnIDForAMissingOrInvalidHeader(t *testing.T) {
	generated := regexp.MustCompile(`^[0-9a-f]{32}$`)
	seen := map[string]bool{}

	for _, header := range []string{"", "771A43A4192642F0B136D5159A501700-B7AD6B7169203331", "771a43a4192642f0b136d5159a501700\ninjected"} {
		req := testkit.Request(t, http.MethodGet, "/trace")
		req.Header.Set("sentry-trace", header)

		got := traceOf(t, req)
		if generated.MatchString(got) == false || seen[got] || got == "771a43a4192642f0b136d5159a501700" {
			t.Errorf("header %q gave trace id %q", header, got)
		}

		seen[got] = true
	}
}

func traceOf(t *testing.T, req *http.Request) string {
	t.Helper()
	router := newTestRouter(t, testkit.DiscardLogs())
	router.GET("/trace", func(c *gin.Context) {
		if logging.TraceID(c) != logging.TraceID(c.Request.Context()) {
			t.Error("gin context and request context disagree on the trace id")
		}

		c.String(http.StatusOK, logging.TraceID(c))
	})

	res := testkit.Serve(router, req)
	if res.Header().Get("X-Trace-Id") != res.Body.String() {
		t.Errorf("X-Trace-Id %q differs from %q", res.Header().Get("X-Trace-Id"), res.Body.String())
	}

	return res.Body.String()
}
