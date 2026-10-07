package prometheus

import (
	"errors"
	"testing"
)

func TestReadsTheAccount(t *testing.T) {
	t.Setenv("METRICS_USER", "prometheus")
	t.Setenv("METRICS_PASSWORD", "secret")

	if cfg, err := Read(); err != nil || cfg != (Config{User: "prometheus", Password: "secret"}) {
		t.Errorf("got %+v, %v", cfg, err)
	}
}

func TestNoAccountTurnsTheMetricsOff(t *testing.T) {
	t.Setenv("METRICS_USER", "")
	t.Setenv("METRICS_PASSWORD", "")

	if cfg, err := Read(); err != nil || cfg != (Config{}) {
		t.Errorf("got %+v, %v", cfg, err)
	}
}

func TestRefusesHalfAnAccount(t *testing.T) {
	for user, password := range map[string]string{"prometheus": "", "": "secret"} {
		t.Setenv("METRICS_USER", user)
		t.Setenv("METRICS_PASSWORD", password)

		if _, err := Read(); errors.Is(err, ErrHalfAccount) == false {
			t.Errorf("user %q, password %q: got %v", user, password, err)
		}
	}
}
