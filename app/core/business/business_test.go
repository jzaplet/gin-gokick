package business_test

import (
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/business"
	"gokick/app/core/httpserver"
	"gokick/app/core/logging"
	"gokick/app/core/telemetry"
	"gokick/app/core/testkit"
)

func TestAnEventReachesTheLogTheMetricAndTheSpan(t *testing.T) {
	collector := testkit.Collector(t)
	tel := telemetry.New(telemetry.Config{
		Endpoint: collector.URL,

		ServiceName: "app",

		SampleRatio: 1,
	}, "test", "v1", testkit.DiscardLogs())
	logger, logs := testkit.CaptureLogs()
	events := business.New(logger, tel)

	engine, err := httpserver.NewEngine(gin.TestMode, nil, testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	engine.Use(tel.Middleware()...)
	engine.GET("/orders", func(c *gin.Context) {
		events.Record(c.Request.Context(), "shop.ordered", slog.Int(logging.KeyDetail, 3))
		events.Record(c.Request.Context(), "shop.ordered")
		c.Status(http.StatusNoContent)
	})

	traceID := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/orders")).Header().Get("X-Trace-Id")

	tel.Shutdown(5 * time.Second)

	entries := testkit.LogEntries(t, logs)
	if len(entries) != 2 || entries[0][logging.KeyEvent] != "shop.ordered" || entries[0][logging.KeyDetail] != float64(3) || entries[0][logging.KeyTraceID] != traceID {
		t.Errorf("log %v", entries)
	}

	points := collector.Latest("business.events")
	if len(points) != 1 || points[0].Value != 2 || len(points[0].Attributes) != 1 || points[0].Attributes[logging.KeyEvent] != "shop.ordered" {
		t.Errorf("metric %+v", points)
	}

	spans := collector.Spans()
	if len(spans) != 1 || len(spans[0].Events) != 2 || spans[0].Events[0].Name != "shop.ordered" || len(spans[0].Events[0].Attributes) != 0 {
		t.Errorf("spans %+v", spans)
	}
}
