package middleware

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/reporting"
	"gokick/app/core/testkit"
)

func TestAccessLogShowsTheRouteNotTheTokenInThePath(t *testing.T) {
	logger, buf := testkit.CaptureLogs()
	router := newTestRouter(t, logger)
	router.GET("/dekujeme/:token", func(c *gin.Context) { c.String(http.StatusOK, "ok") })

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/dekujeme/secret-token?utm_source=x"))

	entries := testkit.LogEntries(t, buf)
	if len(entries) != 1 {
		t.Fatalf("got %d log lines", len(entries))
	}

	entry := entries[0]
	for key, want := range map[string]any{
		"msg": "http request",

		"level": "INFO",

		"method": "GET",

		"path": "/dekujeme/:token",

		"status": 200.0,

		"ip": "192.0.2.1",

		"bytes": 2.0,
	} {
		if entry[key] != want {
			t.Errorf("%s = %v, want %v", key, entry[key], want)
		}
	}

	if entry["trace_id"] == nil || entry["duration_ms"] == nil {
		t.Errorf("missing trace_id or duration_ms in %v", entry)
	}

	if strings.Contains(buf.String(), "secret-token") {
		t.Error("the token from the path reached the log")
	}
}

func TestAccessLogReportsServerErrorsAsErrors(t *testing.T) {
	logger, buf := testkit.CaptureLogs()
	router := newTestRouter(t, logger)
	router.GET("/fail", func(c *gin.Context) {
		reporting.Error(c, errors.New("database unreachable"))
		c.Status(http.StatusServiceUnavailable)
	})

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/fail"))

	if entry := testkit.LogEntries(t, buf)[0]; entry["level"] != "ERROR" || entry["error"] != "database unreachable" {
		t.Errorf("got %v", entry)
	}
}

func TestAccessLogWritesWarningsAtWarnLevel(t *testing.T) {
	logger, buf := testkit.CaptureLogs()
	router := newTestRouter(t, logger)
	router.GET("/degraded", func(c *gin.Context) {
		reporting.Warning(c, errors.New("cache unreachable"))
		reporting.Warning(c, errors.New("queue slow"))
		c.Status(http.StatusNoContent)
	})

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/degraded"))

	if entry := testkit.LogEntries(t, buf)[0]; entry["level"] != "WARN" || entry["warning"] != "cache unreachable; queue slow" || entry["error"] != nil {
		t.Errorf("got %v", entry)
	}
}

func TestAccessLogKeepsServerErrorsAtErrorLevelNextToWarnings(t *testing.T) {
	logger, buf := testkit.CaptureLogs()
	router := newTestRouter(t, logger)
	router.GET("/fail", func(c *gin.Context) {
		reporting.Warning(c, errors.New("cache unreachable"))
		reporting.Error(c, errors.New("database unreachable"))
		c.Status(http.StatusServiceUnavailable)
	})

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/fail"))

	if entry := testkit.LogEntries(t, buf)[0]; entry["level"] != "ERROR" || entry["warning"] != "cache unreachable" || entry["error"] != "database unreachable" {
		t.Errorf("got %v", entry)
	}
}
