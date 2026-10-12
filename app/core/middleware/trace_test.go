package middleware

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/reporting"
	"gokick/app/core/telemetry"
	"gokick/app/core/testkit"
)

const incomingTrace = "771a43a4192642f0b136d5159a501700"

var sentryTraceID = regexp.MustCompile(`"trace_id":"([0-9a-f]{32})"`)

func TestOneTraceIDJoinsTheLogTheSpanTheErrorAndTheHeader(t *testing.T) {
	tests := []struct {
		name string

		headers map[string]string

		want string
	}{
		{"traceparent", map[string]string{"traceparent": "00-" + incomingTrace + "-b7ad6b7169203331-01"}, incomingTrace},
		{"sentry-trace", map[string]string{"sentry-trace": incomingTrace + "-b7ad6b7169203331-1"}, incomingTrace},
		{"traceparent before sentry-trace", map[string]string{
			"traceparent": "00-" + incomingTrace + "-b7ad6b7169203331-00",

			"sentry-trace": strings.Repeat("a", 32) + "-b7ad6b7169203331-1",
		}, incomingTrace},
		{"no header", nil, ""},
		{"sentry-trace in capitals", map[string]string{"sentry-trace": strings.ToUpper(incomingTrace) + "-B7AD6B7169203331"}, ""},
		{"sentry-trace with an injected line", map[string]string{"sentry-trace": incomingTrace + "-b7ad6b7169203331\ninjected"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := traceIDs(t, tt.headers)

			want := tt.want
			if want == "" {
				want = ids["X-Trace-Id"]
			}

			if len(ids) != 4 || generated.MatchString(want) == false || (tt.want == "" && want == incomingTrace) {
				t.Fatalf("trace ids %v", ids)
			}

			for name, id := range ids {
				if id != want {
					t.Errorf("%s has trace id %s, want %s", name, id, want)
				}
			}
		})
	}
}

var generated = regexp.MustCompile(`^[0-9a-f]{32}$`)

func traceIDs(t *testing.T, headers map[string]string) map[string]string {
	t.Helper()

	collector := testkit.Collector(t)

	tel := telemetry.New(telemetry.Config{
		Endpoint: collector.URL,

		ServiceName: "app",

		SampleRatio: 1,
	}, "test", "v1", testkit.DiscardLogs())

	reporter, ingest := newIngestReporter(t)
	logger, buf := testkit.CaptureLogs()
	router := newTracedRouter(t, logger, reporter, tel)
	router.GET("/fail", func(c *gin.Context) {
		reporting.Error(c, errors.New("database unreachable"))
		c.Status(http.StatusInternalServerError)
	})

	req := testkit.Request(t, http.MethodGet, "/fail")
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	res := testkit.Serve(router, req)

	reporter.Flush(2 * time.Second)
	tel.Shutdown(5 * time.Second)

	ids := map[string]string{"X-Trace-Id": res.Header().Get("X-Trace-Id")}

	if entries := testkit.LogEntries(t, buf); len(entries) == 1 {
		ids["log"], _ = entries[0]["trace_id"].(string)
	}

	if spans := collector.Spans(); len(spans) == 1 {
		ids["span"] = spans[0].TraceID
	}

	reported := map[string]bool{}
	for _, match := range sentryTraceID.FindAllStringSubmatch(ingest.received(), -1) {
		reported[match[1]] = true
	}

	for id := range reported {
		ids["error"] += id
	}

	return ids
}

func TestReportedProblemsReachTheSpanWithTheirStack(t *testing.T) {
	tests := []struct {
		name string

		handler gin.HandlerFunc

		message string

		culprit string

		status string
	}{
		{"error", reportError, "database unreachable", "middleware.reportError", "STATUS_CODE_ERROR"},
		{"warning", reportWarning, "mail not sent", "middleware.reportWarning", "STATUS_CODE_UNSET"},
		{"panic", panicking, "panic: boom", "middleware.panicking", "STATUS_CODE_ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, exception := reportedProblem(t, tt.handler)

			if status != tt.status || exception["exception.message"] != tt.message {
				t.Errorf("status %s, message %q", status, exception["exception.message"])
			}

			if strings.Contains(exception["exception.stacktrace"], tt.culprit) == false {
				t.Errorf("no %s in the stack %s", tt.culprit, exception["exception.stacktrace"])
			}
		})
	}
}

func reportError(c *gin.Context) {
	reporting.Error(c, errors.New("database unreachable"))
	c.Status(http.StatusInternalServerError)
}

func reportWarning(c *gin.Context) {
	reporting.Warning(c, errors.New("mail not sent"))
	c.Status(http.StatusNoContent)
}

func panicking(*gin.Context) {
	panic("boom")
}

func reportedProblem(t *testing.T, handler gin.HandlerFunc) (status string, exception map[string]string) {
	t.Helper()

	collector := testkit.Collector(t)
	tel := telemetry.New(telemetry.Config{
		Endpoint: collector.URL,

		ServiceName: "app",

		SampleRatio: 1,
	}, "test", "v1", testkit.DiscardLogs())
	router := newTracedRouter(t, testkit.DiscardLogs(), &reporting.Reporter{}, tel)
	router.GET("/problem", handler)

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/problem"))
	tel.Shutdown(5 * time.Second)

	spans := collector.Spans()
	if len(spans) != 1 || len(spans[0].Events) != 1 || spans[0].Events[0].Name != "exception" {
		t.Fatalf("spans %+v", spans)
	}

	return spans[0].Status, spans[0].Events[0].Attributes
}
