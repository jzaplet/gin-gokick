package telemetry

import (
	"context"
	"regexp"

	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var sentryTraceHeader = regexp.MustCompile(`^([0-9a-f]{32})-([0-9a-f]{16})(?:-([01]))?$`)

type sentryTrace struct{}

func (sentryTrace) Extract(ctx context.Context, carrier propagation.TextMapCarrier) context.Context {
	match := sentryTraceHeader.FindStringSubmatch(carrier.Get("sentry-trace"))
	if match == nil {
		return ctx
	}

	traceID, _ := trace.TraceIDFromHex(match[1])
	spanID, _ := trace.SpanIDFromHex(match[2])

	var flags trace.TraceFlags
	if match[3] == "1" {
		flags = trace.FlagsSampled
	}

	parent := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID: traceID,

		SpanID: spanID,

		TraceFlags: flags,

		Remote: true,
	})
	if parent.IsValid() == false {
		return ctx
	}

	return trace.ContextWithRemoteSpanContext(ctx, parent)
}

func (sentryTrace) Inject(context.Context, propagation.TextMapCarrier) {}

func (sentryTrace) Fields() []string {
	return nil
}
