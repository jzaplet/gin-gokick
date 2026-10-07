package mail_test

import (
	"strings"
	"testing"

	"gokick/app/core/mail"
	"gokick/app/core/testkit"
	"gokick/app/internal/shared/routertest"
	usermail "gokick/app/internal/user/mail"
)

func TestWelcomeGreetsTheUserInTheirLanguage(t *testing.T) {
	locales := routertest.Locales(t, "cs_CZ", "cs_CZ", "en_US")
	renderer := routertest.Renderer(t, locales)

	for name, tc := range map[string]struct {
		lang, prefix, home string
	}{
		"cs_CZ": {"cs-CZ", "/cs", "/"},

		"en_US": {"en-US", "/en", "/en"},
	} {
		mailbox := testkit.Mailbox()
		if err := usermail.Welcome(t.Context(), mail.NewMailer(renderer, mailbox, "https://gokick.dev"), locales.Nearest(name), "jan@example.com"); err != nil {
			t.Fatal(err)
		}

		sent := mailbox.Sent()
		if len(sent) != 1 || sent[0].To != "jan@example.com" {
			t.Fatalf("%s: sent %+v", name, sent)
		}

		markup := testkit.MarkupOf(sent[0].HTML)
		for _, want := range []string{
			`<html lang="` + tc.lang + `"`,
			`href="https://gokick.dev` + tc.prefix + `/app/dashboard"`,
			`href="https://gokick.dev` + tc.home + `"`,
			`href="https://gokick.dev` + tc.prefix + `/app/login"`,
			`src="https://gokick.dev/build/assets/mark-test.png"`,
			`src="https://gokick.dev/build/assets/` + name + `-test.png"`,
		} {
			if strings.Contains(markup, want) == false {
				t.Errorf("%s: no %s in\n%s", name, want, markup)
			}
		}
	}
}
