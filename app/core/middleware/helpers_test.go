package middleware

import (
	"log/slog"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/httpserver"
	"gokick/app/core/reporting"
)

func newTestRouter(t *testing.T, logger *slog.Logger) *gin.Engine {
	t.Helper()

	return newReportingRouter(t, logger, &reporting.Reporter{})
}

func newReportingRouter(t *testing.T, logger *slog.Logger, reporter *reporting.Reporter) *gin.Engine {
	t.Helper()

	engine, err := httpserver.NewEngine(gin.TestMode, nil, logger)
	if err != nil {
		t.Fatal(err)
	}

	engine.Use(Trace(), reporter.Middleware(), AccessLog(logger), Recovery(logger, reporter))

	return engine
}
