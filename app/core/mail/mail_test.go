package mail

import (
	"bytes"
	"io"
	"mime"
	"mime/quotedprintable"
	netmail "net/mail"
	"strings"
	"testing"
	"time"
)

func TestComposeEncodesTheSubjectAndTheBody(t *testing.T) {
	html := "<p>Vítejte " + strings.Repeat("v Gin GoKick ", 12) + "</p>\n<p>a=b</p>"
	msg := read(t, Message{Subject: "Vítejte v Gin GoKick", HTML: html})

	var decoder mime.WordDecoder

	subject, err := decoder.DecodeHeader(msg.Header.Get("Subject"))
	if err != nil || subject != "Vítejte v Gin GoKick" {
		t.Errorf("subject %q %v", subject, err)
	}

	for key, want := range map[string]string{
		"From": `"Gin GoKick" <noreply@example.com>`,

		"To": "<jan@example.com>",

		"Date": "Mon, 05 Oct 2026 12:00:00 +0000",

		"Content-Type": "text/html; charset=utf-8",

		"Content-Transfer-Encoding": "quoted-printable",
	} {
		if got := msg.Header.Get(key); got != want {
			t.Errorf("%s %q, want %q", key, got, want)
		}
	}

	body, err := io.ReadAll(quotedprintable.NewReader(msg.Body))
	if err != nil || string(body) != strings.ReplaceAll(html, "\n", "\r\n") {
		t.Errorf("body %q %v", body, err)
	}
}

func TestComposeKeepsLinesShortAndHeadersWhole(t *testing.T) {
	from, to := addresses(t)

	raw, err := compose(from, to, Message{
		Subject: "Ahoj\r\nBcc: eva@example.com",

		HTML: strings.Repeat("<td>", 400),
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	for line := range strings.SplitSeq(string(raw), "\r\n") {
		if len(line) > 78 {
			t.Errorf("line of %d characters: %s", len(line), line)
		}
	}

	msg, err := netmail.ReadMessage(bytes.NewReader(raw))
	if err != nil || msg.Header.Get("Bcc") != "" {
		t.Errorf("headers %v %v", msg.Header, err)
	}
}

func read(t *testing.T, m Message) *netmail.Message {
	t.Helper()
	from, to := addresses(t)

	raw, err := compose(from, to, m, time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}

	msg, err := netmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	return msg
}

func addresses(t *testing.T) (from, to *netmail.Address) {
	t.Helper()

	from, err := netmail.ParseAddress("Gin GoKick <noreply@example.com>")
	if err != nil {
		t.Fatal(err)
	}

	to, err = netmail.ParseAddress("jan@example.com")
	if err != nil {
		t.Fatal(err)
	}

	return from, to
}
