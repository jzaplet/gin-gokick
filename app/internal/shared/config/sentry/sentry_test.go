package sentry

import "testing"

func TestReadsTheEnvironment(t *testing.T) {
	t.Setenv("SENTRY_DSN", "https://public@o1.ingest.sentry.io/2")
	t.Setenv("SENTRY_FRONTEND_DSN", "https://public@o1.ingest.sentry.io/3")
	t.Setenv("SENTRY_ENVIRONMENT", "staging")
	t.Setenv("SENTRY_RELEASE", "v1.2.0")

	want := Config{
		DSN: "https://public@o1.ingest.sentry.io/2",

		FrontendDSN: "https://public@o1.ingest.sentry.io/3",

		Environment: "staging",

		Release: "v1.2.0",
	}
	if cfg := Read(); cfg != want {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}
