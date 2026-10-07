package mail

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"mime/quotedprintable"
	netmail "net/mail"
	"time"
)

var Discard Sender = discard{}

type Sender interface {
	Send(ctx context.Context, m Message) error
}

type discard struct{}

type Message struct {
	To string

	Subject string

	HTML string
}

func (discard) Send(context.Context, Message) error {
	return nil
}

func compose(from, to *netmail.Address, m Message, at time.Time) ([]byte, error) {
	var out bytes.Buffer
	fmt.Fprintf(&out, "From: %s\r\n", from)
	fmt.Fprintf(&out, "To: %s\r\n", to)
	fmt.Fprintf(&out, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", m.Subject))
	fmt.Fprintf(&out, "Date: %s\r\n", at.Format(time.RFC1123Z))
	out.WriteString("MIME-Version: 1.0\r\n")
	out.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	out.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")

	body := quotedprintable.NewWriter(&out)
	if _, err := body.Write([]byte(m.HTML)); err != nil {
		return nil, err
	}

	if err := body.Close(); err != nil {
		return nil, err
	}

	return out.Bytes(), nil
}
