package inapp

import (
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestMarksOnlyOurFramesOutsideReporting(t *testing.T) {
	frames := []sentry.Frame{
		{Module: "gokick/app/internal/user/handler", Function: "Me"},
		{Module: "gokick", Function: "main"},
		{Module: "gokick/app/core/reporting", Function: "Error"},
		{Module: "gokick/app/core/reporting/request", Function: "Describe"},
		{Module: "gokick", AbsPath: reportingDir + "/reporter.go"},
		{Module: "github.com/gin-gonic/gin", Function: "(*Context).Next"},
		{Module: "gokickfake/app", Function: "Run"},
	}
	want := []bool{true, true, false, false, false, false, false}
	event := &sentry.Event{
		Threads: []sentry.Thread{{Stacktrace: &sentry.Stacktrace{Frames: frames}}, {}},

		Exception: []sentry.Exception{{Stacktrace: &sentry.Stacktrace{Frames: append([]sentry.Frame(nil), frames...)}}},
	}

	Mark(event)

	for _, stacktrace := range []*sentry.Stacktrace{event.Threads[0].Stacktrace, event.Exception[0].Stacktrace} {
		for i, frame := range stacktrace.Frames {
			if frame.InApp != want[i] {
				t.Errorf("frame %s %s in app: %v", frame.Module, frame.AbsPath, frame.InApp)
			}
		}
	}
}
