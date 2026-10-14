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

func (t *Telemetry) Gauge(name, description string, observe func(context.Context) (int64, error)) (stop func()) {
	meter := t.meters.Meter(scope)

	gauge, err := meter.Int64ObservableGauge(name, metric.WithDescription(description))
	if err != nil {
		t.warn(err)

		return func() {}
	}

	callback := func(ctx context.Context, observer metric.Observer) error {
		ctx, cancel := context.WithTimeout(ctx, observeTimeout)
		defer cancel()

		value, observeErr := observe(ctx)
		if observeErr != nil {
			t.warn(fmt.Errorf("gauge %s: %w", name, observeErr))

			return nil
		}

		observer.ObserveInt64(gauge, value)

		return nil
	}

	registration, err := meter.RegisterCallback(callback, gauge)
	if err != nil {
		t.warn(err)

		return func() {}
	}

	return func() {
		if err := registration.Unregister(); err != nil {
			t.warn(err)
		}
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
