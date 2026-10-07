package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const unmatchedRoute = "unmatched"

type Metrics struct {
	registry *prometheus.Registry

	requests *prometheus.CounterVec

	duration *prometheus.HistogramVec

	inFlight prometheus.Gauge
}

func RequireBasicAuth(user, password string) gin.HandlerFunc {
	if user == "" || password == "" {
		return func(c *gin.Context) { c.AbortWithStatus(http.StatusNotFound) }
	}

	return gin.BasicAuth(gin.Accounts{user: password})
}

func New() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),

		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",

			Help: "HTTP requests by method, route and status.",
		}, []string{"method", "route", "status"}),

		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "http_request_duration_seconds",

			Help: "HTTP request duration by method and route.",

			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),

		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",

			Help: "HTTP requests being served.",
		}),
	}
	m.registry.MustRegister(m.requests, m.duration, m.inFlight, collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	return m
}

func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		m.inFlight.Inc()
		defer m.inFlight.Dec()

		c.Next()

		route := c.FullPath()
		if route == "" {
			route = unmatchedRoute
		}

		m.requests.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
		m.duration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
	}
}

func (m *Metrics) Handler() gin.HandlerFunc {
	return gin.WrapH(promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}))
}
