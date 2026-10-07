package logs

import (
	"cmp"
	"fmt"
	"log/slog"
	"os"

	"gokick/app/core/logging"
)

type Config struct {
	Format string

	Level slog.Level
}

func Read() (Config, error) {
	format := cmp.Or(os.Getenv("LOG_FORMAT"), logging.FormatJSON)
	if format != logging.FormatJSON && format != logging.FormatText {
		return Config{}, fmt.Errorf("LOG_FORMAT must be json or text, got %q", format)
	}

	level, err := parseLevel(cmp.Or(os.Getenv("LOG_LEVEL"), "info"))
	if err != nil {
		return Config{}, err
	}

	return Config{Format: format, Level: level}, nil
}

func parseLevel(value string) (slog.Level, error) {
	switch value {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("LOG_LEVEL must be debug, info, warn or error, got %q", value)
	}
}
