package middleware

import (
	"net/http"
	"strings"
	"testing"

	"gokick/app/core/testkit"

	"github.com/gin-gonic/gin"
)

func TestRecoveryAnswers500AndLogsThePanic(t *testing.T) {
	logger, buf := testkit.CaptureLogs()
	router := newTestRouter(t, logger)
	router.GET("/panic", func(*gin.Context) { panic("boom") })

	if res := testkit.Serve(router, testkit.Request(t, http.MethodGet, "/panic")); res.Code != http.StatusInternalServerError {
		t.Errorf("status %d", res.Code)
	}

	entries := testkit.LogEntries(t, buf)
	if len(entries) != 2 {
		t.Fatalf("got %d log lines", len(entries))
	}

	panicked, access := entries[0], entries[1]
	if panicked["msg"] != "panic recovered" || panicked["level"] != "ERROR" || panicked["panic"] != "boom" || panicked["trace_id"] == nil {
		t.Errorf("panic line %v", panicked)
	}

	if stack, _ := panicked["stack"].(string); strings.Contains(stack, "recovery_test.go") == false {
		t.Errorf("the stack does not reach the panic: %q", stack)
	}

	if access["status"] != 500.0 || access["level"] != "ERROR" || access["trace_id"] != panicked["trace_id"] {
		t.Errorf("access line %v", access)
	}
}
