package app

import "testing"

func TestReadsTheEnvironment(t *testing.T) {
	t.Setenv("APP_URL", "http://gokick.local:8020/")
	t.Setenv("APP_ENVIRONMENT", "staging")
	t.Setenv("APP_RELEASE", "v1.2.0")

	cfg, err := Read()
	if want := (Config{
		URL: "http://gokick.local:8020",

		Environment: "staging",

		Release: "v1.2.0",
	}); err != nil || cfg != want {
		t.Errorf("got %+v, %v, want %+v", cfg, err, want)
	}
}

func TestDefaultsToProductionAndTheBuild(t *testing.T) {
	t.Setenv("APP_URL", "https://gokick.dev")
	t.Setenv("APP_ENVIRONMENT", "")
	t.Setenv("APP_RELEASE", "")

	cfg, err := Read()
	if err != nil || cfg.Environment != "production" || cfg.Release != buildRevision() || cfg.Release == "" {
		t.Errorf("got %+v, %v", cfg, err)
	}
}

func TestRejectsAnInvalidURL(t *testing.T) {
	for _, value := range []string{"", "gokick.dev", "ftp://gokick.dev", "https://gokick.dev/cs", "https://gokick.dev/?a=1", "https://jan@gokick.dev"} {
		t.Setenv("APP_URL", value)

		if _, err := Read(); err == nil {
			t.Errorf("APP_URL=%q passed", value)
		}
	}
}
