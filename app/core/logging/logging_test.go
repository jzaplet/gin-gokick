package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"maps"
	"testing"
)

func TestAddsTraceAndUserFromContext(t *testing.T) {
	for _, tc := range []struct {
		ctx context.Context

		want map[string]any
	}{
		{
			WithUserID(WithTraceID(t.Context(), "771a43a4192642f0b136d5159a501700"), "42"),
			map[string]any{KeyTraceID: "771a43a4192642f0b136d5159a501700", KeyUserID: "42"},
		},
		{t.Context(), map[string]any{}},
	} {
		var buf bytes.Buffer

		New(&buf, FormatJSON, slog.LevelInfo).InfoContext(tc.ctx, "signed in")

		entry := decode(t, &buf)
		got := map[string]any{}

		for _, key := range []string{KeyTraceID, KeyUserID} {
			if value, ok := entry[key]; ok {
				got[key] = value
			}
		}

		if maps.Equal(got, tc.want) == false || entry["msg"] != "signed in" {
			t.Errorf("got %v, want %v", entry, tc.want)
		}
	}
}

func TestKeepsTheTraceAfterWith(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, FormatJSON, slog.LevelInfo).With(slog.String(KeyDetail, "worker"))

	logger.InfoContext(WithTraceID(t.Context(), "abc"), "tick")

	entry := decode(t, &buf)
	if entry[KeyTraceID] != "abc" || entry[KeyDetail] != "worker" {
		t.Errorf("got %v", entry)
	}
}

func TestErrorLogWritesErrorRecords(t *testing.T) {
	var buf bytes.Buffer

	ErrorLog(New(&buf, FormatJSON, slog.LevelInfo)).Print("http: TLS handshake error")

	entry := decode(t, &buf)
	if entry["level"] != "ERROR" || entry["msg"] != "http: TLS handshake error" {
		t.Errorf("got %v", entry)
	}
}

func decode(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}

	return entry
}
