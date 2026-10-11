package telemetry_test

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/httpserver"
	"gokick/app/core/telemetry"
	"gokick/app/core/testkit"
)

const parentTrace = "4bf92f3577b34da6a3ce929d0e0e4736"

var traceID = regexp.MustCompile(`^[0-9a-f]{32}$`)

func TestShutdownSendsTheSpansToTheTracesPath(t *testing.T) {
	collector := testkit.Collector(t)
	cfg := telemetry.Config{
		Endpoint: collector.URL + "/otlp/",

		Headers: map[string]string{"Authorization": "Bearer token"},

		ServiceName: "app",

		Attributes: map[string]string{"service.namespace": "shop"},

		SampleRatio: 1,
	}

	tel := telemetry.New(cfg, "staging", "v1", testkit.DiscardLogs())

	res := testkit.Serve(newRouter(t, tel), testkit.Request(t, http.MethodGet, "/notes/7"))
	tel.Shutdown(5 * time.Second)

	requests := collector.Requests()
	if len(requests) != 1 || requests[0].Path != "/otlp/v1/traces" || requests[0].Header.Get("Authorization") != "Bearer token" {
		t.Fatalf("requests %+v", requests)
	}

	spans := collector.Spans()
	if len(spans) != 1 || spans[0].Name != "GET /notes/:id" || spans[0].TraceID != res.Header().Get("X-Trace-Id") {
		t.Fatalf("spans %+v, X-Trace-Id %s", spans, res.Header().Get("X-Trace-Id"))
	}

	resource := spans[0].Resource
	if resource["service.name"] != "app" || resource["service.namespace"] != "shop" || resource["service.version"] != "v1" || resource["deployment.environment.name"] != "staging" {
		t.Errorf("resource %v", resource)
	}
}

func TestSpansKeepTheQueryButNotTheIPOrTheRealPath(t *testing.T) {
	collector := testkit.Collector(t)
	tel := start(t, collector.URL, testkit.DiscardLogs())
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	t.Cleanup(api.Close)

	var dialer net.Dialer

	toAPI := &http.Transport{DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, api.Listener.Addr().String())
	}}
	client := &http.Client{Transport: tel.Transport(toAPI)}

	router := newRouter(t, tel)
	router.GET("/dekujeme/:token", call(client, api.URL+"/v1/messages?model=small", "http://api.telegram.org/bot42:secret-bot-token/sendMessage?chat_id=7"))

	req := testkit.Request(t, http.MethodGet, "/dekujeme/secret-token?utm_source=mail")
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "203.0.113.7")
	req.RemoteAddr = "198.51.100.9:4242"
	testkit.Serve(router, req)
	tel.Shutdown(5 * time.Second)

	spans := collector.Spans()
	if len(spans) != 3 {
		t.Fatalf("spans %+v", spans)
	}

	var server map[string]string

	addresses := map[string]bool{}

	for _, span := range spans {
		for key, value := range span.Attributes {
			if strings.Contains(value, "secret") || strings.Contains(value, "203.0.113.7") || strings.Contains(value, "198.51.100.9") {
				t.Errorf("span %q carries %s=%s", span.Name, key, value)
			}
		}

		if address, found := span.Attributes["url.full"]; found {
			addresses[address] = true
		}

		if _, found := span.Attributes["http.route"]; found {
			server = span.Attributes
		}
	}

	for _, key := range []string{"client.address", "network.peer.address", "network.peer.port", "url.path"} {
		if _, found := server[key]; found {
			t.Errorf("the server span carries %s", key)
		}
	}

	if server["url.query"] != "utm_source=mail" || server["user_agent.original"] != "test-agent" || server["http.route"] != "/dekujeme/:token" {
		t.Errorf("server span %v", server)
	}

	if addresses[api.URL+"/v1/messages?model=small"] == false || addresses["http://api.telegram.org"] == false {
		t.Errorf("outgoing addresses %v", addresses)
	}
}

func TestOutgoingRequestsJoinTheTrace(t *testing.T) {
	tel := start(t, "", testkit.DiscardLogs())
	received := make(chan http.Header, 1)
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received <- r.Header.Clone()

		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(api.Close)

	router := newRouter(t, tel)
	router.GET("/call", call(&http.Client{Transport: tel.Transport(http.DefaultTransport)}, api.URL))

	req := testkit.Request(t, http.MethodGet, "/call")
	req.Header.Set("sentry-trace", parentTrace+"-00f067aa0ba902b7-1")
	testkit.Serve(router, req)

	header := <-received
	if strings.Contains(header.Get("traceparent"), "-"+parentTrace+"-") == false || header.Get("sentry-trace") != "" {
		t.Errorf("traceparent %q, sentry-trace %q", header.Get("traceparent"), header.Get("sentry-trace"))
	}
}

func TestHealthChecksAndUnmatchedPathsAreNotSent(t *testing.T) {
	collector := testkit.Collector(t)
	tel := start(t, collector.URL, testkit.DiscardLogs())
	router := newRouter(t, tel)

	ids := map[string]string{}

	for _, path := range []string{"/readyz", "/missing", "/notes/1"} {
		req := testkit.Request(t, http.MethodGet, path)
		req.Header.Set("traceparent", "00-"+parentTrace+"-00f067aa0ba902b7-00")
		ids[path] = testkit.Serve(router, req).Header().Get("X-Trace-Id")
	}

	tel.Shutdown(5 * time.Second)

	for path, id := range ids {
		if id != parentTrace {
			t.Errorf("%s: X-Trace-Id %q", path, id)
		}
	}

	if spans := collector.Spans(); len(spans) != 1 || spans[0].Name != "GET /notes/:id" || spans[0].TraceID != parentTrace {
		t.Errorf("spans %+v", spans)
	}
}

func TestWithoutAnEndpointNothingIsSent(t *testing.T) {
	collector := testkit.Collector(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", collector.URL)
	tel := start(t, "", testkit.DiscardLogs())

	res := testkit.Serve(newRouter(t, tel), testkit.Request(t, http.MethodGet, "/notes/1"))
	tel.Shutdown(5 * time.Second)

	if traceID.MatchString(res.Header().Get("X-Trace-Id")) == false || len(collector.Requests()) != 0 {
		t.Errorf("X-Trace-Id %q, requests %+v", res.Header().Get("X-Trace-Id"), collector.Requests())
	}
}

func TestExportFailuresReachTheLog(t *testing.T) {
	var config net.ListenConfig

	listener, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	closed := "http://" + listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	logger, buf := testkit.CaptureLogs()
	tel := start(t, closed, logger)

	testkit.Serve(newRouter(t, tel), testkit.Request(t, http.MethodGet, "/notes/1"))
	tel.Shutdown(500 * time.Millisecond)

	entries := testkit.LogEntries(t, buf)
	if len(entries) == 0 || entries[0]["level"] != "WARN" || entries[0]["error"] == nil {
		t.Errorf("log %v", entries)
	}
}

func start(t *testing.T, endpoint string, logger *slog.Logger) *telemetry.Telemetry {
	t.Helper()

	return telemetry.New(telemetry.Config{
		Endpoint: endpoint,

		ServiceName: "app",

		SampleRatio: 1,
	}, "test", "v1", logger)
}

func newRouter(t *testing.T, tel *telemetry.Telemetry) *gin.Engine {
	t.Helper()

	engine, err := httpserver.NewEngine(gin.TestMode, nil, testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	engine.Use(tel.Middleware("/readyz")...)
	engine.GET("/readyz", func(c *gin.Context) { c.Status(http.StatusOK) })
	engine.GET("/notes/:id", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	return engine
}

func call(client *http.Client, addresses ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, address := range addresses {
			req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, address, http.NoBody)
			if err != nil {
				c.Status(http.StatusInternalServerError)

				return
			}

			res, err := client.Do(req)
			if err != nil {
				c.Status(http.StatusBadGateway)

				return
			}

			_ = res.Body.Close()
		}

		c.Status(http.StatusNoContent)
	}
}
