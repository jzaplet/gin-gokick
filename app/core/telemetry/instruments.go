package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"

	"gokick/app/core/logging"
)

const scope = "gokick/app/core/telemetry"

const observeTimeout = 2 * time.Second

type Counter struct {
	counter metric.Int64Counter
}

func (t *Telemetry) Counter(name, description string) Counter {
	counter, err := t.meters.Meter(scope).Int64Counter(name, metric.WithDescription(description))
	if err != nil {
		t.warn(err)

		return Counter{counter: noop.Int64Counter{}}
	}

	return Counter{counter: counter}
}

func (c Counter) Increment(ctx context.Context, attrs ...slog.Attr) {
	c.counter.Add(ctx, 1, metric.WithAttributes(attributes(attrs)...))
}

func (t *Telemetry) Gauge(name, description string, observe func(context.Context) (int64, error)) {
	callback := func(ctx context.Context, observer metric.Int64Observer) error {
		ctx, cancel := context.WithTimeout(ctx, observeTimeout)
		defer cancel()

		value, err := observe(ctx)
		if err != nil {
			t.warn(fmt.Errorf("gauge %s: %w", name, err))

			return nil
		}

		observer.Observe(value)

		return nil
	}

	if _, err := t.meters.Meter(scope).Int64ObservableGauge(name, metric.WithDescription(description), metric.WithInt64Callback(callback)); err != nil {
		t.warn(err)
	}
}

func AddEvent(ctx context.Context, name string) {
	trace.SpanFromContext(ctx).AddEvent(name)
}

func (t *Telemetry) warn(err error) {
	t.logger.LogAttrs(context.Background(), slog.LevelWarn, "opentelemetry failed", logging.Err(err))
}

func attributes(attrs []slog.Attr) []attribute.KeyValue {
	converted := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		converted = append(converted, attribute.String(attr.Key, attr.Value.String()))
	}

	return converted
}
