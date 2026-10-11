package server

import (
	"reflect"
	"testing"
)

func emptyEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{"PORT", "GIN_MODE", "TRUSTED_PROXIES"} {
		t.Setenv(key, "")
	}
}

func TestDefaults(t *testing.T) {
	emptyEnv(t)

	cfg, err := Read()
	if err != nil {
		t.Fatal(err)
	}

	if want := (Config{
		Port: 8020,

		GinMode: "release",
	}); reflect.DeepEqual(cfg, want) == false || cfg.Addr() != ":8020" {
		t.Errorf("got %+v on %q, want %+v", cfg, cfg.Addr(), want)
	}
}

func TestReadsTheEnvironment(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("GIN_MODE", "debug")
	t.Setenv("TRUSTED_PROXIES", " 127.0.0.1 , 10.0.0.0/8,,::1 ")

	cfg, err := Read()
	if err != nil {
		t.Fatal(err)
	}

	want := Config{
		Port: 9000,

		GinMode: "debug",

		TrustedProxies: []string{"127.0.0.1", "10.0.0.0/8", "::1"},
	}
	if reflect.DeepEqual(cfg, want) == false || cfg.Addr() != ":9000" {
		t.Errorf("got %+v on %q, want %+v", cfg, cfg.Addr(), want)
	}
}

func TestRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name, key, value string
	}{
		{"port not a number", "PORT", "http"},
		{"port zero", "PORT", "0"},
		{"port too high", "PORT", "65536"},
		{"port with a sign", "PORT", "+8020"},
		{"unknown gin mode", "GIN_MODE", "production"},
		{"host name as proxy", "TRUSTED_PROXIES", "127.0.0.1,localhost"},
		{"prefix out of range", "TRUSTED_PROXIES", "10.0.0.0/33"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emptyEnv(t)
			t.Setenv(tt.key, tt.value)

			if _, err := Read(); err == nil {
				t.Errorf("%s=%q passed", tt.key, tt.value)
			}
		})
	}
}
