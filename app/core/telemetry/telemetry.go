package telemetry

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"gokick/app/core/logging"
)

const tracesPath = "/v1/traces"

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

	propagator propagation.TextMapPropagator

	logger *slog.Logger
}

func New(cfg Config, environment, release string, logger *slog.Logger) *Telemetry {
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.LogAttrs(context.Background(), slog.LevelWarn, "opentelemetry failed", logging.Err(err))
	}))

	options := append(export(cfg, logger), sdktrace.WithResource(describe(cfg, environment, release)))

	return &Telemetry{
		provider: sdktrace.NewTracerProvider(options...),

		silent: sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample())),

		propagator: propagation.NewCompositeTextMapPropagator(sentryTrace{}, propagation.TraceContext{}),

		logger: logger,
	}
}

func (t *Telemetry) QueryTracer() pgx.QueryTracer {
	return otelpgx.NewTracer(otelpgx.WithTracerProvider(t.provider), otelpgx.WithMeterProvider(noop.NewMeterProvider()))
}

func (t *Telemetry) Transport(base http.RoundTripper) http.RoundTripper {
	return otelhttp.NewTransport(base, otelhttp.WithTracerProvider(t.provider), otelhttp.WithPropagators(t.propagator), otelhttp.WithMeterProvider(noop.NewMeterProvider()))
}

func (t *Telemetry) Shutdown(timeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := t.provider.Shutdown(ctx); err != nil {
		t.logger.LogAttrs(ctx, slog.LevelWarn, "opentelemetry failed", logging.Err(err))
	}
}

func describe(cfg Config, environment, release string) *resource.Resource {
	attributes := make([]attribute.KeyValue, 0, len(cfg.Attributes)+3)
	for key, value := range cfg.Attributes {
		attributes = append(attributes, attribute.String(key, value))
	}

	attributes = append(attributes,
		attribute.String("service.name", cfg.ServiceName),
		attribute.String("service.version", release),
		attribute.String("deployment.environment.name", environment),
	)

	return resource.NewSchemaless(attributes...)
}

func export(cfg Config, logger *slog.Logger) []sdktrace.TracerProviderOption {
	off := []sdktrace.TracerProviderOption{sdktrace.WithSampler(sdktrace.NeverSample())}
	if cfg.Endpoint == "" {
		return off
	}

	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL(strings.TrimSuffix(cfg.Endpoint, "/")+tracesPath), otlptracehttp.WithHeaders(cfg.Headers))
	if err != nil {
		logger.LogAttrs(context.Background(), slog.LevelWarn, "opentelemetry failed", logging.Err(err))

		return off
	}

	ratio := sdktrace.TraceIDRatioBased(cfg.SampleRatio)
	sampler := sdktrace.ParentBased(ratio, sdktrace.WithRemoteParentSampled(ratio), sdktrace.WithRemoteParentNotSampled(ratio))

	return []sdktrace.TracerProviderOption{sdktrace.WithSampler(sampler), sdktrace.WithBatcher(scrubbingExporter{exporter})}
}
