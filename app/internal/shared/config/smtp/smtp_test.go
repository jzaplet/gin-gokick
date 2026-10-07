package smtp

import (
	"reflect"
	"testing"
)

func emptyEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{"SMTP_ENABLED", "SMTP_HOST", "SMTP_PORT", "SMTP_USERNAME", "SMTP_PASSWORD", "SMTP_SENDER_EMAIL", "SMTP_SENDER_NAME"} {
		t.Setenv(key, "")
	}
}

func TestMailIsOffUnlessEnabledAndReadsNothingElse(t *testing.T) {
	for _, enabled := range []string{"", "false"} {
		emptyEnv(t)
		t.Setenv("SMTP_ENABLED", enabled)
		t.Setenv("SMTP_HOST", "smtp://mail.gokick.local")
		t.Setenv("SMTP_SENDER_EMAIL", "not an address")

		cfg, err := Read()
		if err != nil || reflect.DeepEqual(cfg, Config{}) == false || cfg.Server().Addr != "" {
			t.Errorf("SMTP_ENABLED=%q: got %+v, %v", enabled, cfg, err)
		}
	}
}

func TestReadsMailpitAndSES(t *testing.T) {
	for name, tc := range map[string]struct {
		env map[string]string

		addr, from string

		want Config
	}{
		"mailpit": {
			env: map[string]string{
				"SMTP_ENABLED": "true",

				"SMTP_HOST": "mail.gokick.local",

				"SMTP_PORT": "1025",

				"SMTP_SENDER_EMAIL": "noreply@gokick.local",

				"SMTP_SENDER_NAME": "Gin GoKick",
			},

			addr: "mail.gokick.local:1025",

			from: `"Gin GoKick" <noreply@gokick.local>`,

			want: Config{
				Enabled: true,

				Host: "mail.gokick.local",

				Port: 1025,

				SenderEmail: "noreply@gokick.local",

				SenderName: "Gin GoKick",
			},
		},

		"ses": {
			env: map[string]string{
				"SMTP_ENABLED": "true",

				"SMTP_HOST": "email-smtp.eu-central-1.amazonaws.com",

				"SMTP_USERNAME": "AKIA",

				"SMTP_PASSWORD": "secret",

				"SMTP_SENDER_EMAIL": "noreply@gokick.dev",
			},

			addr: "email-smtp.eu-central-1.amazonaws.com:587",

			from: "<noreply@gokick.dev>",

			want: Config{
				Enabled: true,

				Host: "email-smtp.eu-central-1.amazonaws.com",

				Port: 587,

				Username: "AKIA",

				Password: "secret",

				SenderEmail: "noreply@gokick.dev",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			expectRead(t, tc.env, tc.addr, tc.from, &tc.want)
		})
	}
}

func expectRead(t *testing.T, env map[string]string, addr, from string, want *Config) {
	t.Helper()
	emptyEnv(t)

	for key, value := range env {
		t.Setenv(key, value)
	}

	cfg, err := Read()
	if err != nil || reflect.DeepEqual(cfg, *want) == false || cfg.Server().Addr != addr || cfg.Server().Username != want.Username || cfg.From() != from {
		t.Errorf("got %+v on %q from %q, %v", cfg, cfg.Addr(), cfg.From(), err)
	}
}

func TestRejectsInvalidValues(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"enabled not a boolean": {"SMTP_ENABLED": "1"},

		"enabled in capitals": {"SMTP_ENABLED": "TRUE"},

		"no host": {"SMTP_HOST": ""},

		"host with a port": {"SMTP_HOST": "mail.gokick.local:1025"},

		"host with a scheme": {"SMTP_HOST": "smtp://mail.gokick.local"},

		"port not a number": {"SMTP_PORT": "smtp"},

		"port zero": {"SMTP_PORT": "0"},

		"port of implicit tls": {"SMTP_PORT": "465"},

		"username without a password": {"SMTP_USERNAME": "AKIA"},

		"password without a username": {"SMTP_PASSWORD": "secret"},

		"no sender": {"SMTP_SENDER_EMAIL": ""},

		"sender with a name": {"SMTP_SENDER_EMAIL": "Gin GoKick <noreply@gokick.dev>"},

		"sender without a domain": {"SMTP_SENDER_EMAIL": "noreply"},
	} {
		t.Run(name, func(t *testing.T) {
			emptyEnv(t)
			t.Setenv("SMTP_ENABLED", "true")
			t.Setenv("SMTP_HOST", "mail.gokick.local")
			t.Setenv("SMTP_SENDER_EMAIL", "noreply@gokick.local")

			for key, value := range env {
				t.Setenv(key, value)
			}

			if _, err := Read(); err == nil {
				t.Errorf("%v passed", env)
			}
		})
	}
}
