package telemetry

import (
	"slices"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"

	"gokick/app/core/logging"
)

func (t *Telemetry) Middleware(untraced ...string) gin.HandlersChain {
	traced := t.server(t.provider)
	silent := t.server(t.silent)

	active, err := t.meters.Meter(scope).Int64UpDownCounter("http.server.active_requests", metric.WithDescription("Number of active HTTP server requests."), metric.WithUnit("{request}"))
	if err != nil {
		t.warn(err)

		active = noop.Int64UpDownCounter{}
	}

	return gin.HandlersChain{
		func(c *gin.Context) {
			if route := c.FullPath(); route == "" || slices.Contains(untraced, route) {
				silent(c)

				return
			}

			traced(c)
		},
		func(c *gin.Context) {
			span := trace.SpanFromContext(c.Request.Context())
			if query := c.Request.URL.RawQuery; query != "" {
				span.SetAttributes(attribute.String("url.query", query))
			}

			id := span.SpanContext().TraceID().String()
			c.Request = c.Request.WithContext(logging.WithTraceID(c.Request.Context(), id))
			c.Header("X-Trace-Id", id)

			active.Add(c.Request.Context(), 1)
			defer active.Add(c.Request.Context(), -1)

			c.Next()
		},
	}
}

func (t *Telemetry) server(provider trace.TracerProvider) gin.HandlerFunc {
	return otelgin.Middleware("", otelgin.WithTracerProvider(provider), otelgin.WithPropagators(t.propagator), otelgin.WithMeterProvider(t.meters))
}
