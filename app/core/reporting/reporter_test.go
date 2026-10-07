package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
	"gokick/app/core/testkit"
)

const traceID = "771a43a4192642f0b136d5159a501700"

const userID = "0192f3a4-5b6c-7d8e-9f01-23456789abcd"

func TestReportsAPanicUnderTheRequestTrace(t *testing.T) {
	reporter, sent := newCapturingReporter(t)
	router := newTestRouter(reporter)
	router.GET("/dekujeme/:token", func(*gin.Context) { panic("boom") })

	req := testkit.Request(t, http.MethodGet, "/dekujeme/secret-token?utm_source=x")
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Authorization", "Bearer secret-bearer")
	req.Header.Set("Cookie", "session=secret-cookie")

	testkit.Serve(router, req)

	event := sent.only(t)
	if event.Message != "boom" || event.Transaction != "GET /dekujeme/:token" {
		t.Errorf("message %q, transaction %q", event.Message, event.Transaction)
	}

	if got := fmt.Sprint(event.Contexts["trace"]["trace_id"]); got != traceID {
		t.Errorf("trace id %s", got)
	}

	want := map[string]string{"User-Agent": "test-agent", "Authorization": "[Filtered]", "Cookie": "[Filtered]"}
	if event.Request.URL != "/dekujeme/:token" || event.Request.QueryString != "utm_source=x" || fmt.Sprint(event.Request.Headers) != fmt.Sprint(want) {
		t.Errorf("request %+v", event.Request)
	}

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(payload), "secret") {
		t.Errorf("a secret reached the event: %s", payload)
	}

	if strings.Contains(string(payload), "reporter_test.go") == false {
		t.Error("the event has no stack reaching the panic")
	}
}

func TestLinksTheEventToTheFrontendSpan(t *testing.T) {
	reporter, sent := newCapturingReporter(t)
	router := newTestRouter(reporter)
	router.GET("/panic", func(*gin.Context) { panic("boom") })

	req := testkit.Request(t, http.MethodGet, "/panic")
	req.Header.Set("sentry-trace", traceID+"-b7ad6b7169203331-1")
	req.Header.Set("baggage", "sentry-trace_id="+traceID+",sentry-public_key=public")

	testkit.Serve(router, req)

	if got := fmt.Sprint(sent.only(t).Contexts["trace"]["parent_span_id"]); got != "b7ad6b7169203331" {
		t.Errorf("parent span id %s", got)
	}
}

func TestReportsTheStackWhereTheHandlerReportedTheError(t *testing.T) {
	reporter, sent := newCapturingReporter(t)
	router := newTestRouter(reporter)
	router.GET("/fail", func(c *gin.Context) {
		Error(c, errors.New("database unreachable"))
		c.Status(http.StatusServiceUnavailable)
	})

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/fail"))

	exceptions := sent.only(t).Exception

	stack := exceptions[len(exceptions)-1].Stacktrace
	if stack == nil || len(stack.Frames) < 2 {
		t.Fatalf("exceptions %+v", exceptions)
	}

	newest, handler := stack.Frames[len(stack.Frames)-1], stack.Frames[len(stack.Frames)-2]
	if newest.Function != "Error" || newest.InApp {
		t.Errorf("newest frame %s in app %v", newest.Function, newest.InApp)
	}

	if strings.HasPrefix(handler.Function, "TestReportsTheStackWhereTheHandlerReportedTheError") == false {
		t.Errorf("the frame under reporting.Error is %s, not the handler", handler.Function)
	}
}

func TestWarningReportsWithoutFailingTheRequest(t *testing.T) {
	reporter, sent := newCapturingReporter(t)
	router := newTestRouter(reporter)
	router.GET("/degraded", func(c *gin.Context) {
		Warning(c, errors.New("cache unreachable"))
		c.Status(http.StatusNoContent)
	})

	res := testkit.Serve(router, testkit.Request(t, http.MethodGet, "/degraded"))

	event := sent.only(t)

	stack := event.Exception[len(event.Exception)-1].Stacktrace
	if res.Code != http.StatusNoContent || event.Level != sentry.LevelWarning || stack == nil || len(stack.Frames) < 2 {
		t.Fatalf("got %d, event %+v", res.Code, event)
	}

	newest, handler := stack.Frames[len(stack.Frames)-1], stack.Frames[len(stack.Frames)-2]
	if newest.Function != "Warning" || strings.HasPrefix(handler.Function, "TestWarningReportsWithoutFailingTheRequest") == false {
		t.Errorf("newest frame %s, the frame under it %s", newest.Function, handler.Function)
	}
}

func TestReportsHandlerErrorsOfServerErrors(t *testing.T) {
	reporter, sent := newCapturingReporter(t)
	router := newTestRouter(reporter)
	router.GET("/fail", func(c *gin.Context) {
		_ = c.Error(errors.New("database unreachable"))
		c.Status(http.StatusServiceUnavailable)
	})

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/fail"))

	exceptions := sent.only(t).Exception
	if len(exceptions) == 0 || exceptions[len(exceptions)-1].Value != "database unreachable" {
		t.Fatalf("exceptions %+v", exceptions)
	}

	for _, exception := range exceptions {
		if exception.Stacktrace != nil {
			t.Error("the error carries the middleware's stack, so every handler error would share one culprit")
		}
	}
}

func TestReportsEveryHandlerErrorOfTheRequest(t *testing.T) {
	reporter, sent := newCapturingReporter(t)
	router := newTestRouter(reporter)
	router.GET("/fail", func(c *gin.Context) {
		_ = c.Error(errors.New("database unreachable"))
		_ = c.Error(errors.New("template missing"))
		c.Status(http.StatusInternalServerError)
	})

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/fail"))

	payload, err := json.Marshal(sent.only(t).Exception)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(string(payload), "database unreachable") == false || strings.Contains(string(payload), "template missing") == false {
		t.Errorf("exceptions %s", payload)
	}
}

func TestReportsTheSignedInUserByIDOnly(t *testing.T) {
	for name, handler := range map[string]gin.HandlerFunc{
		"a panic": func(*gin.Context) { panic("boom") },

		"a handler error": func(c *gin.Context) {
			Error(c, errors.New("database unreachable"))
			c.Status(http.StatusInternalServerError)
		},

		"a warning": func(c *gin.Context) {
			Warning(c, errors.New("cache unreachable"))
			c.Status(http.StatusNoContent)
		},
	} {
		t.Run(name, func(t *testing.T) {
			reporter, sent := newCapturingReporter(t)
			router := newTestRouter(reporter)
			router.GET("/me", signIn, handler)

			testkit.Serve(router, testkit.Request(t, http.MethodGet, "/me"))

			if user := reportedUser(t, sent.only(t)); user != `{"id":"`+userID+`"}` {
				t.Errorf("user %s", user)
			}
		})
	}
}

func TestLeavesTheUserOutOfTheNextRequest(t *testing.T) {
	reporter, sent := newCapturingReporter(t)
	router := newTestRouter(reporter)
	fail := func(c *gin.Context) {
		Error(c, errors.New("database unreachable"))
		c.Status(http.StatusInternalServerError)
	}
	router.GET("/me", signIn, fail)
	router.GET("/public", fail)

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/me"))
	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/public"))

	events := sent.all()
	if len(events) != 2 {
		t.Fatalf("sent %d events", len(events))
	}

	if events[1].User.IsEmpty() == false {
		t.Errorf("the second user %+v", events[1].User)
	}
}

func TestNamesAnUnmatchedRoute(t *testing.T) {
	reporter, sent := newCapturingReporter(t)
	router := newTestRouter(reporter)
	router.NoRoute(func(*gin.Context) { panic("boom") })

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/missing"))

	if event := sent.only(t); event.Request.URL != "unmatched" || event.Transaction != "GET unmatched" {
		t.Errorf("url %q, transaction %q", event.Request.URL, event.Transaction)
	}
}

func TestReleaseFallsBackToTheBuild(t *testing.T) {
	reporter, err := New("https://public@example.com/1", "test", "")
	if err != nil {
		t.Fatal(err)
	}

	if got := reporter.hub.Client().Options().Release; got != buildRevision() || got == "" {
		t.Errorf("release %q", got)
	}
}

func TestLeavesClientErrorsOut(t *testing.T) {
	reporter, sent := newCapturingReporter(t)
	router := newTestRouter(reporter)
	router.GET("/bad", func(c *gin.Context) {
		_ = c.Error(errors.New("invalid e-mail"))
		c.Status(http.StatusUnprocessableEntity)
	})

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/bad"))

	if events := sent.all(); len(events) != 0 {
		t.Errorf("sent %d events", len(events))
	}
}

type transport struct {
	mu sync.Mutex

	events []*sentry.Event
}

func (t *transport) Configure(sentry.ClientOptions) {}

func (t *transport) SendEvent(event *sentry.Event) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.events = append(t.events, event)
}

func (t *transport) Flush(time.Duration) bool { return true }

func (t *transport) FlushWithContext(context.Context) bool { return true }

func (t *transport) Close() {}

func (t *transport) all() []*sentry.Event {
	t.mu.Lock()
	defer t.mu.Unlock()

	return append([]*sentry.Event(nil), t.events...)
}

func (t *transport) only(tb testing.TB) *sentry.Event {
	tb.Helper()

	events := t.all()
	if len(events) != 1 {
		tb.Fatalf("sent %d events", len(events))
	}

	return events[0]
}

func signIn(c *gin.Context) {
	c.Request = c.Request.WithContext(logging.WithUserID(c.Request.Context(), userID))
}

func reportedUser(t *testing.T, event *sentry.Event) string {
	t.Helper()

	user, err := json.Marshal(event.User)
	if err != nil {
		t.Fatal(err)
	}

	return string(user)
}

func newCapturingReporter(t *testing.T) (*Reporter, *transport) {
	t.Helper()

	sent := &transport{}

	reporter, err := newReporter(&sentry.ClientOptions{Dsn: "https://public@example.com/1", Transport: sent})
	if err != nil {
		t.Fatal(err)
	}

	return reporter, sent
}

func newTestRouter(reporter *Reporter) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.ContextWithFallback = true
	withTrace := func(c *gin.Context) {
		c.Request = c.Request.WithContext(logging.WithTraceID(c.Request.Context(), traceID))
	}
	recovery := gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, recovered any) {
		reporter.Recover(c, recovered)
		c.AbortWithStatus(http.StatusInternalServerError)
	})
	router.Use(withTrace, reporter.Middleware(), recovery)

	return router
}
