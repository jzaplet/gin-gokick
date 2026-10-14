package mail

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gokick/app/core/business"
	"gokick/app/core/locale/posix"
	"gokick/app/core/logging"
)

const EventSent business.Event = "mail.sent"

const EventFailed business.Event = "mail.failed"

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

	events *business.Recorder
}

func NewMailer(renderer Renderer, sender Sender, origin string, events *business.Recorder) *Mailer {
	return &Mailer{renderer: renderer, sender: sender, origin: origin, events: events}
}

func (m *Mailer) Send(ctx context.Context, t *Template) error {
	err := m.send(ctx, t)

	event := EventSent
	if err != nil {
		event = EventFailed
	}

	m.events.Record(ctx, event, slog.String(logging.KeyMail, t.Name))

	return err
}

func (m *Mailer) send(ctx context.Context, t *Template) error {
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
