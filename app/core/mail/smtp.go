package mail

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	netmail "net/mail"
	"net/smtp"
	"time"
)

type SMTPServer struct {
	Addr string

	Username string

	Password string
}

type smtpSender struct {
	server SMTPServer

	host string

	tls *tls.Config

	from *netmail.Address
}

func New(server SMTPServer, from string) (Sender, error) {
	if server.Addr == "" {
		return Discard, nil
	}

	return newSMTP(server, from)
}

func newSMTP(server SMTPServer, from string) (*smtpSender, error) {
	host, _, err := net.SplitHostPort(server.Addr)
	if err != nil {
		return nil, fmt.Errorf("mail: address %q: %w", server.Addr, err)
	}

	sender, err := netmail.ParseAddress(from)
	if err != nil {
		return nil, fmt.Errorf("mail: sender %q: %w", from, err)
	}

	return &smtpSender{
		server: server,

		host: host,

		tls: &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12},

		from: sender,
	}, nil
}

func (s *smtpSender) Send(ctx context.Context, m Message) error {
	to, err := netmail.ParseAddress(m.To)
	if err != nil {
		return fmt.Errorf("mail: recipient: %w", err)
	}

	body, err := compose(s.from, to, m, time.Now())
	if err != nil {
		return fmt.Errorf("mail: compose: %w", err)
	}

	client, err := s.connect(ctx)
	if err != nil {
		return err
	}

	if err := s.signIn(client); err != nil {
		return errors.Join(err, client.Close())
	}

	if err := deliver(client, s.from.Address, to.Address, body); err != nil {
		return errors.Join(err, client.Close())
	}

	return client.Quit()
}

func (s *smtpSender) connect(ctx context.Context) (*smtp.Client, error) {
	var dialer net.Dialer

	conn, err := dialer.DialContext(ctx, "tcp", s.server.Addr)
	if err != nil {
		return nil, fmt.Errorf("mail: dial %s: %w", s.server.Addr, err)
	}

	deadline, _ := ctx.Deadline()

	err = conn.SetDeadline(deadline)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("mail: deadline: %w", err), conn.Close())
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("mail: greeting of %s: %w", s.server.Addr, err), conn.Close())
	}

	return client, nil
}

func (s *smtpSender) signIn(client *smtp.Client) error {
	if offered, _ := client.Extension("STARTTLS"); offered {
		if err := client.StartTLS(s.tls); err != nil {
			return fmt.Errorf("mail: starttls: %w", err)
		}
	}

	if s.server.Username == "" {
		return nil
	}

	if err := client.Auth(smtp.PlainAuth("", s.server.Username, s.server.Password, s.host)); err != nil {
		return fmt.Errorf("mail: sign in: %w", err)
	}

	return nil
}

func deliver(client *smtp.Client, from, to string, body []byte) error {
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("mail: sender: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("mail: recipient: %w", err)
	}

	data, err := client.Data()
	if err != nil {
		return fmt.Errorf("mail: data: %w", err)
	}

	if _, err := data.Write(body); err != nil {
		return fmt.Errorf("mail: body: %w", err)
	}

	if err := data.Close(); err != nil {
		return fmt.Errorf("mail: body: %w", err)
	}

	return nil
}
