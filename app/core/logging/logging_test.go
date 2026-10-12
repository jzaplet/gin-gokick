package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"maps"
	"strings"
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

func TestTheExportStartsAtTheLevelOfTheLog(t *testing.T) {
	var exported bytes.Buffer

	logger := New(io.Discard, FormatJSON, slog.LevelInfo, export(&exported))
	logger.DebugContext(t.Context(), "cache missed")
	logger.InfoContext(t.Context(), "signed in")

	if strings.Count(exported.String(), "\n") != 1 || decode(t, &exported)["msg"] != "signed in" {
		t.Errorf("exported %q", exported.String())
	}
}

func TestTheExportLeavesTheTraceToItsSpanAndTheConsoleStaysAsItWas(t *testing.T) {
	ctx := WithUserID(WithTraceID(t.Context(), "771a43a4192642f0b136d5159a501700"), "42")

	var alone, beside, exported bytes.Buffer

	New(&alone, FormatJSON, slog.LevelInfo).InfoContext(ctx, "signed in", slog.String(KeyDetail, "web"))
	New(&beside, FormatJSON, slog.LevelInfo, export(&exported)).InfoContext(ctx, "signed in", slog.String(KeyDetail, "web"))

	console := decode(t, &beside)
	want := decode(t, &alone)

	delete(console, "time")
	delete(want, "time")

	if maps.Equal(console, want) == false {
		t.Errorf("console %v, want %v", console, want)
	}

	entry := decode(t, &exported)
	if _, found := entry[KeyTraceID]; found || entry[KeyUserID] != "42" || entry[KeyDetail] != "web" {
		t.Errorf("exported %v", entry)
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

func export(w io.Writer) slog.Handler {
	return slog.NewJSONHandler(w, &slog.HandlerOptions{Level: slog.LevelDebug})
}

func decode(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatal(err)
	}

	return entry
}
