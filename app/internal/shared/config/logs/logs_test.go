package logs

import (
	"log/slog"
	"testing"
)

func TestDefaults(t *testing.T) {
	t.Setenv("LOG_FORMAT", "")
	t.Setenv("LOG_LEVEL", "")

	cfg, err := Read()
	if err != nil || cfg != (Config{Format: "json", Level: slog.LevelInfo}) {
		t.Errorf("got %+v, %v", cfg, err)
	}
}

func TestReadsTheEnvironment(t *testing.T) {
	for value, level := range map[string]slog.Level{
		"debug": slog.LevelDebug,

		"info": slog.LevelInfo,

		"warn": slog.LevelWarn,

		"error": slog.LevelError,
	} {
		t.Setenv("LOG_FORMAT", "text")
		t.Setenv("LOG_LEVEL", value)

		cfg, err := Read()
		if err != nil || cfg != (Config{Format: "text", Level: level}) {
			t.Errorf("LOG_LEVEL=%s: got %+v, %v", value, cfg, err)
		}
	}
}

func TestRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name, key, value string
	}{
		{"unknown log format", "LOG_FORMAT", "logfmt"},
		{"unknown log level", "LOG_LEVEL", "trace"},
		{"log level in capitals", "LOG_LEVEL", "INFO"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("LOG_FORMAT", "")
			t.Setenv("LOG_LEVEL", "")
			t.Setenv(tt.key, tt.value)

			if _, err := Read(); err == nil {
				t.Errorf("%s=%q passed", tt.key, tt.value)
			}
		})
	}
}
