package mailbox

import (
	"context"
	"slices"
	"sync"

	"gokick/app/core/mail"
)

type Mailbox struct {
	Err error

	mu sync.Mutex

	sent []mail.Message
}

func (m *Mailbox) Send(_ context.Context, msg mail.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Err != nil {
		return m.Err
	}

	m.sent = append(m.sent, msg)

	return nil
}

func (m *Mailbox) Sent() []mail.Message {
	m.mu.Lock()
	defer m.mu.Unlock()

	return slices.Clone(m.sent)
}
