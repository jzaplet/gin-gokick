package glitchtip

import "testing"

func TestReadsTheEnvironment(t *testing.T) {
	t.Setenv("GLITCHTIP_DSN", "https://public@glitchtip.example.com/2")
	t.Setenv("GLITCHTIP_FRONTEND_DSN", "https://public@glitchtip.example.com/3")

	want := Config{DSN: "https://public@glitchtip.example.com/2", FrontendDSN: "https://public@glitchtip.example.com/3"}
	if cfg := Read(); cfg != want {
		t.Errorf("got %+v, want %+v", cfg, want)
	}
}
