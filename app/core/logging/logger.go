package logging

import (
	"io"
	"log"
	"log/slog"
	"time"
)

const FormatJSON = "json"

const FormatText = "text"

func New(w io.Writer, format string, level slog.Level) *slog.Logger {
	options := &slog.HandlerOptions{Level: level}
	if format == FormatText {
		return slog.New(contextHandler{slog.NewTextHandler(w, options)})
	}

	return slog.New(contextHandler{slog.NewJSONHandler(w, options)})
}

func ErrorLog(logger *slog.Logger) *log.Logger {
	return slog.NewLogLogger(logger.Handler(), slog.LevelError)
}

func Err(err error) slog.Attr {
	return slog.Any(KeyError, err)
}

func DurationMs(d time.Duration) slog.Attr {
	return slog.Float64(KeyDurationMs, float64(d.Microseconds())/1000)
}
