package telemetry

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

func RecordError(ctx context.Context, err error) {
	trace.SpanFromContext(ctx).RecordError(err, trace.WithStackTrace(true))
}
