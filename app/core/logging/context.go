package logging

import (
	"context"
	"log/slog"
)

type traceIDKey struct{}

type userIDKey struct{}

func WithTraceID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceIDKey{}, id)
}

func TraceID(ctx context.Context) string {
	id, _ := ctx.Value(traceIDKey{}).(string)

	return id
}

func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}

func UserID(ctx context.Context) string {
	id, _ := ctx.Value(userIDKey{}).(string)

	return id
}

type contextHandler struct {
	slog.Handler
}

func (h contextHandler) Handle(ctx context.Context, record slog.Record) error {
	if id := TraceID(ctx); id != "" {
		record.AddAttrs(slog.String(KeyTraceID, id))
	}

	if id := UserID(ctx); id != "" {
		record.AddAttrs(slog.String(KeyUserID, id))
	}

	return h.Handler.Handle(ctx, record)
}

func (h contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return contextHandler{h.Handler.WithAttrs(attrs)}
}

func (h contextHandler) WithGroup(name string) slog.Handler {
	return contextHandler{h.Handler.WithGroup(name)}
}

type exportHandler struct {
	slog.Handler

	level slog.Level
}

func (h exportHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level && h.Handler.Enabled(ctx, level)
}

func (h exportHandler) Handle(ctx context.Context, record slog.Record) error {
	exported := slog.NewRecord(record.Time, record.Level, record.Message, record.PC)
	record.Attrs(func(attr slog.Attr) bool {
		exported.AddAttrs(plain(attr))

		return true
	})

	if id := UserID(ctx); id != "" {
		exported.AddAttrs(slog.String(KeyUserID, id))
	}

	return h.Handler.Handle(ctx, exported)
}

func (h exportHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	converted := make([]slog.Attr, 0, len(attrs))
	for _, attr := range attrs {
		converted = append(converted, plain(attr))
	}

	return exportHandler{Handler: h.Handler.WithAttrs(converted), level: h.level}
}

func (h exportHandler) WithGroup(name string) slog.Handler {
	return exportHandler{Handler: h.Handler.WithGroup(name), level: h.level}
}

func plain(attr slog.Attr) slog.Attr {
	attr.Value = attr.Value.Resolve()
	if err, ok := attr.Value.Any().(error); ok {
		attr.Value = slog.StringValue(err.Error())
	}

	return attr
}
