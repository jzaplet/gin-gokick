package mail

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"gokick/app/core/locale/posix"
)

type renderer struct {
	name, origin string

	locale posix.Locale

	params any

	err error
}

type outbox struct {
	sent []Message

	deadline time.Time

	err error
}

func (r *renderer) Mail(name string, l posix.Locale, origin string, params any) (subject, html string, err error) {
	r.name, r.locale, r.origin, r.params = name, l, origin, params

	return "Ahoj", "<p>Ahoj</p>", r.err
}

func (o *outbox) Send(ctx context.Context, m Message) error {
	o.sent = append(o.sent, m)
	o.deadline, _ = ctx.Deadline()

	return o.err
}

func TestSendRendersTheTemplateInItsLocaleAndSendsIt(t *testing.T) {
	english, err := posix.Parse("en_US")
	if err != nil {
		t.Fatal(err)
	}

	templates, sent := &renderer{}, &outbox{}
	start := time.Now()

	err = NewMailer(templates, sent, "https://gokick.dev").Send(t.Context(), &Template{
		Name: "mail/welcome.html",

		To: "jan@example.com",

		Locale: english,

		Params: 42,
	})
	if err != nil {
		t.Fatal(err)
	}

	if templates.name != "mail/welcome.html" || templates.locale != english || templates.origin != "https://gokick.dev" || templates.params != 42 {
		t.Errorf("rendered %+v", templates)
	}

	if len(sent.sent) != 1 || sent.sent[0] != (Message{To: "jan@example.com", Subject: "Ahoj", HTML: "<p>Ahoj</p>"}) {
		t.Errorf("sent %+v", sent.sent)
	}

	if limit := sent.deadline.Sub(start); limit < sendTimeout || limit > sendTimeout+time.Second {
		t.Errorf("the sender had %v", limit)
	}
}

func TestSendNamesTheTemplateThatFailed(t *testing.T) {
	broken := errors.New("broken")
	for name, mailer := range map[string]*Mailer{
		"rendering": NewMailer(&renderer{err: broken}, &outbox{}, ""),

		"sending": NewMailer(&renderer{}, &outbox{err: broken}, ""),
	} {
		err := mailer.Send(t.Context(), &Template{Name: "mail/welcome.html", To: "jan@example.com"})
		if errors.Is(err, broken) == false || strings.Contains(err.Error(), "mail/welcome.html") == false {
			t.Errorf("%s: error %v", name, err)
		}
	}
}
