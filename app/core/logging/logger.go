package logging

import (
	"io"
	"log"
	"log/slog"
	"time"
)

const FormatJSON = "json"

const FormatText = "text"

func New(w io.Writer, format string, level slog.Level, exports ...slog.Handler) *slog.Logger {
	options := &slog.HandlerOptions{Level: level}

	var console slog.Handler = slog.NewJSONHandler(w, options)
	if format == FormatText {
		console = slog.NewTextHandler(w, options)
	}

	handlers := []slog.Handler{contextHandler{console}}
	for _, export := range exports {
		handlers = append(handlers, exportHandler{Handler: export, level: level})
	}

	return slog.New(slog.NewMultiHandler(handlers...))
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
