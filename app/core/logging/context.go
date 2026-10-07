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
