package telemetry_test

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/database"
	"gokick/app/core/httpserver"
	"gokick/app/core/telemetry"
	"gokick/app/core/testkit"
)

const parentTrace = "4bf92f3577b34da6a3ce929d0e0e4736"

var traceID = regexp.MustCompile(`^[0-9a-f]{32}$`)

func TestShutdownSendsSpansAndMetricsToTheirPaths(t *testing.T) {
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

	paths := map[string]bool{}
	for _, request := range collector.Requests() {
		paths[request.Path] = request.Header.Get("Authorization") == "Bearer token"
	}

	if len(paths) != 2 || paths["/otlp/v1/traces"] == false || paths["/otlp/v1/metrics"] == false {
		t.Fatalf("paths and their authorization %v", paths)
	}

	spans := collector.Spans()
	if len(spans) != 1 || spans[0].Name != "GET /notes/:id" || spans[0].TraceID != res.Header().Get("X-Trace-Id") {
		t.Fatalf("spans %+v, X-Trace-Id %s", spans, res.Header().Get("X-Trace-Id"))
	}

	points := collector.Latest("http.server.request.duration")
	if len(points) == 0 {
		t.Fatal("no request metrics")
	}

	for _, resource := range []map[string]string{spans[0].Resource, points[0].Resource} {
		if resource["service.name"] != "app" || resource["service.namespace"] != "shop" || resource["service.version"] != "v1" || resource["deployment.environment.name"] != "staging" {
			t.Errorf("resource %v", resource)
		}
	}
}

func TestEveryStartIsAnInstanceOfItsOwn(t *testing.T) {
	collector := testkit.Collector(t)

	for _, attributes := range []map[string]string{nil, nil, {"service.instance.id": "app-1"}} {
		tel := telemetry.New(telemetry.Config{
			Endpoint: collector.URL,

			ServiceName: "app",

			Attributes: attributes,

			SampleRatio: 1,
		}, "test", "v1", testkit.DiscardLogs())

		testkit.Serve(newRouter(t, tel), testkit.Request(t, http.MethodGet, "/notes/1"))
		tel.Shutdown(5 * time.Second)
	}

	var instances []string
	for _, span := range collector.Spans() {
		instances = append(instances, span.Resource["service.instance.id"])
	}

	if len(instances) != 3 || instances[0] == "" || instances[0] == instances[1] || instances[2] != "app-1" {
		t.Errorf("instances %v", instances)
	}
}

func TestMetricsCountRequestsByTheirRoutePattern(t *testing.T) {
	collector := testkit.Collector(t)
	tel := start(t, collector.URL, testkit.DiscardLogs())
	router := newRouter(t, tel)

	for _, path := range []string{"/notes/7", "/notes/8", "/readyz", "/missing"} {
		req := testkit.Request(t, http.MethodGet, path)
		req.Host = "forged.example"
		req.Header.Set("User-Agent", "forged-agent")
		testkit.Serve(router, req)
	}

	tel.Shutdown(5 * time.Second)

	counts := map[string]uint64{}

	for _, point := range collector.Latest("http.server.request.duration") {
		for key, value := range point.Attributes {
			if strings.Contains(value, "forged") || strings.Contains(value, "/notes/7") || strings.Contains(value, "/missing") {
				t.Errorf("request metric carries %s=%s", key, value)
			}
		}

		counts[point.Attributes["http.request.method"]+" "+point.Attributes["http.route"]+" "+point.Attributes["http.response.status_code"]] += point.Count
	}

	if counts["GET /notes/:id 204"] != 2 || counts["GET /readyz 200"] != 1 || counts["GET  404"] != 1 {
		t.Errorf("requests by route %v", counts)
	}

	names := collector.Metrics()
	for _, dropped := range []string{"http.server.request.body.size", "http.server.response.body.size"} {
		if slices.Contains(names, dropped) {
			t.Errorf("%s was sent", dropped)
		}
	}

	for _, runtime := range []string{"go.goroutine.count", "go.memory.used", "go.schedule.duration", "go.gc.cycles", "go.gc.cpu_time", "go.gc.pause.cpu_time"} {
		if slices.Contains(names, runtime) == false {
			t.Errorf("%s is missing from %v", runtime, names)
		}
	}

	if active := collector.Latest("http.server.active_requests"); len(active) != 1 || active[0].Value != 0 {
		t.Errorf("active requests after the requests ended %+v", active)
	}
}

func TestMetricsStayCumulativeWhenTheEnvironmentPrefersDeltas(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE", "delta")
	collector := testkit.Collector(t)
	tel := start(t, collector.URL, testkit.DiscardLogs())

	testkit.Serve(newRouter(t, tel), testkit.Request(t, http.MethodGet, "/notes/1"))
	tel.Shutdown(5 * time.Second)

	if points := collector.Latest("http.server.request.duration"); len(points) == 0 || points[0].Delta {
		t.Errorf("points %+v", points)
	}
}

func TestShutdownSendsTheMetricsWhileTheTracesHang(t *testing.T) {
	metrics := make(chan string, 1)
	release := make(chan struct{})
	hanging := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/v1/traces") {
			<-release

			return
		}

		metrics <- r.URL.Path

		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(hanging.Close)
	t.Cleanup(func() { close(release) })

	tel := start(t, hanging.URL, testkit.DiscardLogs())

	testkit.Serve(newRouter(t, tel), testkit.Request(t, http.MethodGet, "/notes/1"))
	tel.Shutdown(500 * time.Millisecond)

	select {
	case <-metrics:
	default:
		t.Error("no metrics were sent")
	}
}

func TestQueriesAreNamedBySqlcOrByTheirOperation(t *testing.T) {
	collector := testkit.Collector(t)
	tel := start(t, collector.URL, testkit.DiscardLogs())

	pool, err := database.Open(t.Context(), testkit.Database(t), 0, tel.QueryTracer())
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)
	tel.ObservePool(pool)

	router := newRouter(t, tel)
	router.GET("/queries", func(c *gin.Context) {
		for _, query := range []string{"-- name: CountNotes :one\nSELECT 1", "SELECT 1"} {
			if _, err := pool.Exec(c.Request.Context(), query); err != nil {
				c.Status(http.StatusInternalServerError)

				return
			}
		}

		c.Status(http.StatusNoContent)
	})

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/queries"))
	tel.Shutdown(5 * time.Second)

	var spans []string
	for _, span := range collector.Spans() {
		spans = append(spans, span.Name)
	}

	operations := map[string]bool{}
	for _, point := range collector.Latest("db.client.operation.duration") {
		operations[point.Attributes["db.operation.name"]] = true
	}

	if slices.Contains(spans, "CountNotes") == false || slices.Contains(spans, "SELECT") == false || operations["CountNotes"] == false || operations["SELECT"] == false {
		t.Errorf("spans %v, operations %v", spans, operations)
	}

	if slices.Contains(collector.Metrics(), "pgxpool.total_connections") == false {
		t.Errorf("pool metrics are missing from %v", collector.Metrics())
	}
}

func TestAFailingGaugeReachesTheLogAndStopsNoOtherMetric(t *testing.T) {
	collector := testkit.Collector(t)
	logger, logs := testkit.CaptureLogs()
	tel := start(t, collector.URL, logger)

	_ = tel.Gauge("notes.open", "Open notes.", func(context.Context) (int64, error) { return 7, nil })
	_ = tel.Gauge("notes.archived", "Archived notes.", func(context.Context) (int64, error) { return 0, errors.New("notes are gone") })
	tel.Shutdown(5 * time.Second)

	if open := collector.Latest("notes.open"); len(open) != 1 || open[0].Value != 7 {
		t.Errorf("open notes %+v", open)
	}

	if entries := testkit.LogEntries(t, logs); len(entries) != 1 || entries[0]["level"] != "WARN" || strings.Contains(fmt.Sprint(entries[0]["error"]), "notes.archived: notes are gone") == false {
		t.Errorf("log %v", entries)
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
