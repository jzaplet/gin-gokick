package business

import (
	"context"
	"log/slog"

	"gokick/app/core/logging"
	"gokick/app/core/telemetry"
)

type Event string

type Recorder struct {
	logger *slog.Logger

	events telemetry.Counter
}

func New(logger *slog.Logger, tel *telemetry.Telemetry) *Recorder {
	return &Recorder{logger: logger, events: tel.Counter("business.events", "Business events by name.")}
}

func (r *Recorder) Record(ctx context.Context, event Event, attrs ...slog.Attr) {
	named := slog.String(logging.KeyEvent, string(event))

	r.events.Increment(ctx, named)
	telemetry.AddEvent(ctx, string(event))
	r.logger.LogAttrs(ctx, slog.LevelInfo, "business event", append([]slog.Attr{named}, attrs...)...)
}
