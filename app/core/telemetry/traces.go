package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"gokick/app/core/logging"
)

const tracesPath = "/v1/traces"

func tracing(cfg Config, logger *slog.Logger) []sdktrace.TracerProviderOption {
	off := []sdktrace.TracerProviderOption{sdktrace.WithSampler(sdktrace.NeverSample())}
	if cfg.Endpoint == "" {
		return off
	}

	exporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpointURL(endpoint(cfg.Endpoint, tracesPath)), otlptracehttp.WithHeaders(cfg.Headers))
	if err != nil {
		logger.LogAttrs(context.Background(), slog.LevelWarn, "opentelemetry failed", logging.Err(err))

		return off
	}

	ratio := sdktrace.TraceIDRatioBased(cfg.SampleRatio)
	sampler := sdktrace.ParentBased(ratio, sdktrace.WithRemoteParentSampled(ratio), sdktrace.WithRemoteParentNotSampled(ratio))

	return []sdktrace.TracerProviderOption{sdktrace.WithSampler(sampler), sdktrace.WithBatcher(scrubbingExporter{exporter})}
}
