package mail

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"io"
	"mime/quotedprintable"
	"net"
	"net/http"
	"net/http/httptest"
	netmail "net/mail"
	"net/textproto"
	"strings"
	"testing"
)

type envelope struct {
	from, to, data, credentials string

	encrypted bool
}

func TestSendDeliversTheMessage(t *testing.T) {
	addr, received := server(t, nil)

	sender, err := newSMTP(SMTPServer{Addr: addr}, "Gin GoKick <noreply@example.com>")
	if err != nil {
		t.Fatal(err)
	}

	err = sender.Send(t.Context(), Message{To: "Jan <jan@example.com>", Subject: "Ahoj", HTML: "<p>.a</p>\n.\n"})
	if err != nil {
		t.Fatal(err)
	}

	got := <-received
	if got.from != "FROM:<noreply@example.com>" || got.to != "TO:<jan@example.com>" || got.encrypted || got.credentials != "" {
		t.Errorf("envelope %+v", got)
	}

	msg, err := netmail.ReadMessage(strings.NewReader(got.data))
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(quotedprintable.NewReader(msg.Body))
	if err != nil || msg.Header.Get("Subject") != "Ahoj" || string(body) != "<p>.a</p>\n.\n" {
		t.Errorf("subject %q, body %q %v", msg.Header.Get("Subject"), body, err)
	}
}

func TestSendSignsInOverStartTLS(t *testing.T) {
	certificate := httptest.NewTLSServer(http.NotFoundHandler())
	t.Cleanup(certificate.Close)
	addr, received := server(t, certificate.TLS)

	sender, err := newSMTP(SMTPServer{Addr: addr, Username: "user", Password: "secret"}, "noreply@example.com")
	if err != nil {
		t.Fatal(err)
	}

	sender.tls.RootCAs = x509.NewCertPool()
	sender.tls.RootCAs.AddCert(certificate.Certificate())

	if err := sender.Send(t.Context(), Message{To: "jan@example.com", Subject: "Ahoj", HTML: "<p>a</p>"}); err != nil {
		t.Fatal(err)
	}

	if got := <-received; got.encrypted == false || got.credentials != "\x00user\x00secret" || got.to != "TO:<jan@example.com>" {
		t.Errorf("envelope %+v", got)
	}
}

func TestSendFailsWithoutAServer(t *testing.T) {
	sender, err := newSMTP(SMTPServer{Addr: "127.0.0.1:1"}, "noreply@example.com")
	if err != nil {
		t.Fatal(err)
	}

	if err := sender.Send(t.Context(), Message{To: "jan@example.com", Subject: "Ahoj", HTML: "<p>a</p>"}); err == nil {
		t.Error("sent without a server")
	}
}

func TestNewRefusesABadAddressOrSender(t *testing.T) {
	if _, err := New(SMTPServer{Addr: "localhost"}, "noreply@example.com"); err == nil {
		t.Error("took an address without a port")
	}

	if _, err := New(SMTPServer{Addr: "localhost:1025"}, "noreply"); err == nil {
		t.Error("took a sender without a domain")
	}
}

func server(t *testing.T, encryption *tls.Config) (addr string, received <-chan envelope) {
	t.Helper()
	var config net.ListenConfig

	listener, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = listener.Close() })

	envelopes := make(chan envelope, 1)

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}

		session := &smtpSession{conn: conn, text: textproto.NewConn(conn), encryption: encryption}
		defer func() { _ = session.conn.Close() }()

		envelopes <- session.converse()
	}()

	return listener.Addr().String(), envelopes
}

type smtpSession struct {
	conn net.Conn

	text *textproto.Conn

	encryption *tls.Config

	envelope envelope
}

func (s *smtpSession) reply(line string) {
	_ = s.text.PrintfLine("%s", line)
}

func (s *smtpSession) converse() envelope {
	s.reply("220 test")

	for {
		line, err := s.text.ReadLine()
		if err != nil {
			return s.envelope
		}

		verb, arg, _ := strings.Cut(line, " ")
		if s.answer(verb, arg) == false {
			return s.envelope
		}
	}
}

func (s *smtpSession) answer(verb, arg string) bool {
	switch verb {
	case "EHLO", "HELO":
		if s.encryption != nil && s.envelope.encrypted == false {
			s.reply("250-test")
			s.reply("250 STARTTLS")
		} else {
			s.reply("250-test")
			s.reply("250 AUTH PLAIN")
		}
	case "STARTTLS":
		s.reply("220 go on")
		s.conn = tls.Server(s.conn, s.encryption)
		s.text = textproto.NewConn(s.conn)
		s.envelope.encrypted = true
	case "AUTH":
		credentials, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(arg, "PLAIN "))
		s.envelope.credentials = string(credentials)
		s.reply("235 ok")
	case "MAIL":
		s.envelope.from = arg
		s.reply("250 ok")
	case "RCPT":
		s.envelope.to = arg
		s.reply("250 ok")
	case "DATA":
		s.reply("354 go on")
		data, _ := s.text.ReadDotBytes()
		s.envelope.data = string(data)
		s.reply("250 ok")
	case "QUIT":
		s.reply("221 bye")

		return false
	default:
		s.reply("502 unknown")
	}

	return true
}
