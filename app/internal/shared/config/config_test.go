package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"

	corelocale "gokick/app/core/locale"
)

func unsetExample(t *testing.T) map[string]string {
	t.Helper()

	example, err := godotenv.Read("../../../../.env.example")
	if err != nil {
		t.Fatal(err)
	}

	for key := range example {
		t.Setenv(key, "")

		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}

	return example
}

func requiredEnvOnly(t *testing.T) {
	t.Helper()
	unsetExample(t)
	t.Setenv("DB_HOST", "db")
	t.Setenv("DB_NAME", "app")
	t.Setenv("DB_USERNAME", "app")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DEFAULT_LOCALE", "pt_BR")
	t.Setenv("APP_URL", "https://gokick.dev")
}

func TestStopsAtAnInvalidSection(t *testing.T) {
	for key, value := range map[string]string{
		"PORT": "http",

		"LOG_LEVEL": "trace",

		"DB_HOST": "",

		"APP_URL": "",

		"SMTP_ENABLED": "yes",

		"METRICS_USER": "prometheus",

		"UMAMI_WEBSITE_ID": "site",

		"DEFAULT_LOCALE": "",
	} {
		t.Run(key, func(t *testing.T) {
			requiredEnvOnly(t)
			t.Setenv(key, value)

			if cfg, err := fromEnv(); err == nil || cfg != nil {
				t.Errorf("%s=%q passed", key, value)
			}
		})
	}
}

func TestLoadReadsDotenvButTheEnvironmentWins(t *testing.T) {
	requiredEnvOnly(t)
	t.Setenv("GIN_MODE", "test")
	writeDotenv(t, "PORT=9100\nGIN_MODE=debug\n")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Server.Port != 9100 || cfg.Server.GinMode != "test" {
		t.Errorf("got %+v", cfg.Server)
	}
}

func TestLoadWithoutDotenv(t *testing.T) {
	requiredEnvOnly(t)
	t.Chdir(t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Server.Port != 8020 {
		t.Errorf("Port = %d", cfg.Server.Port)
	}
}

func TestEnvExampleIsValid(t *testing.T) {
	example := unsetExample(t)
	for key, value := range example {
		t.Setenv(key, value)
	}

	cfg, err := fromEnv()
	if err != nil {
		t.Fatalf(".env.example: %v", err)
	}

	if cfg.Server.GinMode != "debug" || cfg.Log.Format != "text" || cfg.Database.Host != "db."+example["APP_DOMAIN"] || cfg.Database.Params.Get("pool_max_conns") == "" || cfg.Server.URL != "https://"+example["APP_DOMAIN"] || cfg.SMTP.Enabled == false || cfg.SMTP.Addr() != "mail."+example["APP_DOMAIN"]+":1025" || cfg.SMTP.Username != "" || cfg.Sentry.Environment != example["SENTRY_ENVIRONMENT"] {
		t.Errorf("the example was not read: %+v", cfg)
	}

	if _, err := corelocale.New(cfg.Locale.Default, cfg.Locale.Locales); err != nil {
		t.Errorf(".env.example: %v", err)
	}
}

func writeDotenv(t *testing.T, content string) {
	t.Helper()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Chdir(dir)
}
