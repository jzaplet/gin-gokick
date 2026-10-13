package telemetry

import (
	"context"
	"runtime/metrics"

	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func observeGC(provider *sdkmetric.MeterProvider) error {
	meter := provider.Meter(scope)

	cycles, err := meter.Float64ObservableCounter("go.gc.cycles", metric.WithDescription("Completed GC cycles."), metric.WithUnit("{cycle}"))
	if err != nil {
		return err
	}

	busy, err := meter.Float64ObservableCounter("go.gc.cpu_time", metric.WithDescription("CPU time of the GC."), metric.WithUnit("s"))
	if err != nil {
		return err
	}

	paused, err := meter.Float64ObservableCounter("go.gc.pause.cpu_time", metric.WithDescription("Pauses of the GC, GOMAXPROCS times their duration."), metric.WithUnit("s"))
	if err != nil {
		return err
	}

	_, err = meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		samples := []metrics.Sample{{Name: "/gc/cycles/total:gc-cycles"}, {Name: "/cpu/classes/gc/total:cpu-seconds"}, {Name: "/cpu/classes/gc/pause:cpu-seconds"}}
		metrics.Read(samples)

		observer.ObserveFloat64(cycles, float64(samples[0].Value.Uint64()))
		observer.ObserveFloat64(busy, samples[1].Value.Float64())
		observer.ObserveFloat64(paused, samples[2].Value.Float64())

		return nil
	}, cycles, busy, paused)

	return err
}
