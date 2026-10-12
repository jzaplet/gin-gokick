package telemetry

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"

	"gokick/app/core/logging"
)

const metricsPath = "/v1/metrics"

const metricsInterval = 30 * time.Second

var requestAttributes = []attribute.Key{"http.request.method", "http.route", "http.response.status_code", "error.type"}

func metering(cfg Config, described *resource.Resource, logger *slog.Logger) *sdkmetric.MeterProvider {
	options := []sdkmetric.Option{
		sdkmetric.WithResource(described),
		sdkmetric.WithView(
			sdkmetric.NewView(sdkmetric.Instrument{Name: "http.server.request.duration"}, sdkmetric.Stream{AttributeFilter: attribute.NewAllowKeysFilter(requestAttributes...)}),
			sdkmetric.NewView(sdkmetric.Instrument{Name: "http.server.request.body.size"}, sdkmetric.Stream{Aggregation: sdkmetric.AggregationDrop{}}),
			sdkmetric.NewView(sdkmetric.Instrument{Name: "http.server.response.body.size"}, sdkmetric.Stream{Aggregation: sdkmetric.AggregationDrop{}}),
		),
	}
	if cfg.Endpoint == "" {
		return sdkmetric.NewMeterProvider(options...)
	}

	exporter, err := otlpmetrichttp.New(context.Background(), otlpmetrichttp.WithEndpointURL(endpoint(cfg.Endpoint, metricsPath)), otlpmetrichttp.WithHeaders(cfg.Headers), otlpmetrichttp.WithTemporalitySelector(sdkmetric.DefaultTemporalitySelector))
	if err != nil {
		logger.LogAttrs(context.Background(), slog.LevelWarn, "opentelemetry failed", logging.Err(err))

		return sdkmetric.NewMeterProvider(options...)
	}

	provider := sdkmetric.NewMeterProvider(append(options, sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter, sdkmetric.WithInterval(metricsInterval), sdkmetric.WithProducer(runtime.NewProducer()))))...)
	if err := runtime.Start(runtime.WithMeterProvider(provider)); err != nil {
		logger.LogAttrs(context.Background(), slog.LevelWarn, "opentelemetry failed", logging.Err(err))
	}

	return provider
}
