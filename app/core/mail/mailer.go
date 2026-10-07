package mail

import (
	"context"
	"fmt"
	"time"

	"gokick/app/core/locale/posix"
)

const sendTimeout = 3 * time.Second

type Renderer interface {
	Mail(name string, l posix.Locale, origin string, params any) (subject, html string, err error)
}

type Template struct {
	Name string

	To string

	Locale posix.Locale

	Params any
}

type Mailer struct {
	renderer Renderer

	sender Sender

	origin string
}

func NewMailer(renderer Renderer, sender Sender, origin string) *Mailer {
	return &Mailer{renderer: renderer, sender: sender, origin: origin}
}

func (m *Mailer) Send(ctx context.Context, t *Template) error {
	subject, html, err := m.renderer.Mail(t.Name, t.Locale, m.origin, t.Params)
	if err != nil {
		return fmt.Errorf("mail %s: %w", t.Name, err)
	}

	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()

	if err := m.sender.Send(ctx, Message{To: t.To, Subject: subject, HTML: html}); err != nil {
		return fmt.Errorf("mail %s: %w", t.Name, err)
	}

	return nil
}
