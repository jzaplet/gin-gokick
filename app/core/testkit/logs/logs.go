package logs

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"testing"

	"gokick/app/core/logging"
)

func Discard() *slog.Logger {
	return logging.New(io.Discard, logging.FormatJSON, slog.LevelInfo)
}

func Capture() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer

	return logging.New(&buf, logging.FormatJSON, slog.LevelDebug), &buf
}

func Entries(tb testing.TB, buf *bytes.Buffer) []map[string]any {
	tb.Helper()
	var entries []map[string]any

	for line := range strings.Lines(buf.String()) {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			tb.Fatal(err)
		}

		entries = append(entries, entry)
	}

	return entries
}
