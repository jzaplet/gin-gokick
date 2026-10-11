package sentry

import "testing"

func TestReadsTheEnvironment(t *testing.T) {
	t.Setenv("SENTRY_DSN", "https://public@o1.ingest.sentry.io/2")
	t.Setenv("SENTRY_FRONTEND_DSN", "https://public@o1.ingest.sentry.io/3")

	want := Config{DSN: "https://public@o1.ingest.sentry.io/2", FrontendDSN: "https://public@o1.ingest.sentry.io/3"}
	if cfg := Read(); cfg != want {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}
