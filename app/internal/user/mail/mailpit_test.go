//go:build mailpit

package mail_test

import (
	"maps"
	"slices"
	"testing"

	"gokick/app/core/locale"
	"gokick/app/core/mail"
	"gokick/app/core/testkit"
	"gokick/app/core/vite"
	"gokick/app/internal/shared/config"
	"gokick/app/internal/shared/config/server"
	"gokick/app/internal/shared/config/smtp"
	"gokick/app/internal/shared/templates"
	usermail "gokick/app/internal/user/mail"
	dictionaries "gokick/locale"
	"gokick/public"
)

func TestSendSamplesToMailpit(t *testing.T) {
	testkit.LoadEnv(t)

	site, err := server.Read()
	if err != nil {
		t.Fatal(err)
	}

	smtpConfig, err := smtp.Read()
	if err != nil || smtpConfig.Enabled == false {
		t.Fatalf("SMTP_ENABLED is not true or the section is invalid, see .env.example: %v", err)
	}

	sender, err := mail.New(smtpConfig.Server(), smtpConfig.From())
	if err != nil {
		t.Fatal(err)
	}

	names := slices.Sorted(maps.Keys(dictionaries.All()))

	locales, err := locale.New(names[0], names)
	if err != nil {
		t.Fatal(err)
	}

	assets, err := vite.Load(public.FS, testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	renderer, err := templates.Parse(&config.Config{}, assets, nil, locales)
	if err != nil {
		t.Fatal(err)
	}

	mailer := mail.NewMailer(renderer, sender, site.URL)

	for _, l := range locales.Locales() {
		if err := usermail.Welcome(t.Context(), mailer, l, "jan@example.com"); err != nil {
			t.Fatal(err)
		}

		t.Logf("%s sent to %s", l, smtpConfig.Server().Addr)
	}
}
