package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"gokick/app/core/logging"
)

type Config struct {
	Endpoint string

	Headers map[string]string

	ServiceName string

	Attributes map[string]string

	SampleRatio float64
}

type Telemetry struct {
	provider *sdktrace.TracerProvider

	silent *sdktrace.TracerProvider

	meters *sdkmetric.MeterProvider

	propagator propagation.TextMapPropagator

	logger *slog.Logger
}

func New(cfg Config, environment, release string, logger *slog.Logger) *Telemetry {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.LogAttrs(context.Background(), slog.LevelWarn, "opentelemetry failed", logging.Err(err))
	}))

	described := describe(cfg, environment, release)

	return &Telemetry{
		provider: sdktrace.NewTracerProvider(append(tracing(cfg, logger), sdktrace.WithResource(described))...),

		silent: sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample())),

		meters: metering(cfg, described, logger),

		propagator: propagation.NewCompositeTextMapPropagator(sentryTrace{}, propagation.TraceContext{}),

		logger: logger,
	}
}

func (t *Telemetry) QueryTracer() pgx.QueryTracer {
	return otelpgx.NewTracer(
		otelpgx.WithTracerProvider(t.provider),
		otelpgx.WithMeterProvider(t.meters),
		otelpgx.WithSpanNameCtxFunc(queryName),
		otelpgx.WithMetricOperationName(queryName),
	)
}

func (t *Telemetry) ObservePool(pool *pgxpool.Pool) {
	if err := otelpgx.RecordStats(pool, otelpgx.WithStatsMeterProvider(t.meters)); err != nil {
		t.warn(err)
	}
}

func (t *Telemetry) Transport(base http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(base, otelhttp.WithTracerProvider(t.provider), otelhttp.WithPropagators(t.propagator), otelhttp.WithMeterProvider(t.meters))
}

func (t *Telemetry) Shutdown(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var (
		wg sync.WaitGroup

		traces, metrics error
	)

	wg.Go(func() { traces = t.provider.Shutdown(ctx) })
	wg.Go(func() { metrics = t.meters.Shutdown(ctx) })
	wg.Wait()

	if err := errors.Join(traces, metrics); err != nil {
		t.logger.LogAttrs(ctx, slog.LevelWarn, "opentelemetry failed", logging.Err(err))
	}
}

func describe(cfg Config, environment, release string) *resource.Resource {
	attributes := make([]attribute.KeyValue, 0, len(cfg.Attributes)+4)
	for key, value := range cfg.Attributes {
		attributes = append(attributes, attribute.String(key, value))
	}

	if _, set := cfg.Attributes["service.instance.id"]; set == false {
		attributes = append(attributes, attribute.String("service.instance.id", uuid.NewString()))
	}

	attributes = append(attributes,
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", release),
		attribute.String("deployment.environment.name", environment),
	)

	return resource.NewSchemaless(attributes...)
}

func endpoint(base, path string) string {
	return strings.TrimSuffix(base, "/") + path
}
