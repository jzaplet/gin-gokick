package metrics

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"gokick/app/core/httpserver"
	"gokick/app/core/middleware"
	"gokick/app/core/reporting"
	"gokick/app/core/testkit"
)

func TestMiddlewareCountsRequestsByRouteTemplate(t *testing.T) {
	m, router := newTestRouter(t)
	var inFlight float64

	router.GET("/users/:id", func(c *gin.Context) {
		inFlight = testutil.ToFloat64(m.inFlight)

		c.Status(http.StatusNoContent)
	})
	router.GET("/panic", func(*gin.Context) { panic("boom") })

	for _, path := range []string{"/users/42", "/users/43", "/wp-login.php", "/panic"} {
		testkit.Serve(router, testkit.Request(t, http.MethodGet, path))
	}

	body := scrape(t, m)
	for _, want := range []string{
		`http_requests_total{method="GET",route="/users/:id",status="204"} 2`,
		`http_requests_total{method="GET",route="unmatched",status="404"} 1`,
		`http_requests_total{method="GET",route="/panic",status="500"} 1`,
		`http_request_duration_seconds_count{method="GET",route="/users/:id"} 2`,
		"http_requests_in_flight 0",
		"go_goroutines",
	} {
		if strings.Contains(body, want) == false {
			t.Errorf("metrics lack %s", want)
		}
	}

	if strings.Contains(body, "/users/42") || strings.Contains(body, "wp-login") {
		t.Error("a raw path reached the metrics")
	}

	if inFlight != 1 {
		t.Errorf("in flight during the request: %v", inFlight)
	}
}

func TestRequireBasicAuthTurnsTheEndpointOffWithoutAFullAccount(t *testing.T) {
	for _, account := range [][2]string{{"", ""}, {"prometheus", ""}, {"", "secret"}} {
		if res := testkit.Serve(guarded(account[0], account[1]), withAccount(t, account[0], account[1])); res.Code != http.StatusNotFound {
			t.Errorf("%q: %d", account, res.Code)
		}
	}
}

func guarded(user, password string) *gin.Engine {
	engine := gin.New()
	engine.GET("/metrics", RequireBasicAuth(user, password), func(c *gin.Context) { c.Status(http.StatusOK) })

	return engine
}

func withAccount(t *testing.T, user, password string) *http.Request {
	t.Helper()
	req := testkit.Request(t, http.MethodGet, "/metrics")
	req.SetBasicAuth(user, password)

	return req
}

func newTestRouter(t *testing.T) (*Metrics, *gin.Engine) {
	t.Helper()

	engine, err := httpserver.NewEngine(gin.TestMode, nil, testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	m := New()
	engine.Use(m.Middleware(), middleware.Recovery(testkit.DiscardLogs(), &reporting.Reporter{}))

	return m, engine
}

func scrape(t *testing.T, m *Metrics) string {
	t.Helper()

	router := gin.New()
	router.GET("/metrics", m.Handler())

	res := testkit.Serve(router, testkit.Request(t, http.MethodGet, "/metrics"))
	if res.Code != http.StatusOK {
		t.Fatalf("status %d", res.Code)
	}

	return res.Body.String()
}
