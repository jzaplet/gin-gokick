package telemetry

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"

	"gokick/app/core/logging"
)

const logsPath = "/v1/logs"

const logScope = "gokick/app/core/logging"

func recording(cfg Config, described *resource.Resource, logger *slog.Logger) *sdklog.LoggerProvider {
	if cfg.Endpoint == "" {
		return sdklog.NewLoggerProvider()
	}

	exporter, err := otlploghttp.New(context.Background(), otlploghttp.WithEndpointURL(endpoint(cfg.Endpoint, logsPath)), otlploghttp.WithHeaders(cfg.Headers))
	if err != nil {
		logger.LogAttrs(context.Background(), slog.LevelWarn, "opentelemetry failed", logging.Err(err))

		return sdklog.NewLoggerProvider()
	}

	return sdklog.NewLoggerProvider(sdklog.WithResource(described), sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)))
}

func (t *Telemetry) LogHandler() slog.Handler {
	return otelslog.NewHandler(logScope, otelslog.WithLoggerProvider(t.logs))
}
