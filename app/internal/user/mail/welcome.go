package mail

import (
	"context"

	"gokick/app/core/locale/posix"
	coremail "gokick/app/core/mail"
)

const welcomeTemplate = "mail/welcome.html"

type WelcomeParams struct {
	Email string
}

func Welcome(ctx context.Context, mailer *coremail.Mailer, l posix.Locale, email string) error {
	return mailer.Send(ctx, &coremail.Template{
		Name: welcomeTemplate,

		To: email,

		Locale: l,

		Params: WelcomeParams{Email: email},
	})
}
