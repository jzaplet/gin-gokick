package view

import (
	"errors"
	"testing"
	"testing/fstest"

	"gokick/app/core/locale/posix"
)

var mails = fstest.MapFS{
	"mail/welcome.html": {Data: []byte(`<title>{{template "subject" .}}</title><a href="{{.Origin}}/{{.Locale.Language}}">{{t "mail.home"}}</a>{{define "subject"}}{{t "mail.subject"}} {{.Params.Name}}{{end}}`)},
}

func TestMailWritesTheSubjectAndTheBodyInTheLanguageOfTheReader(t *testing.T) {
	subject, html, err := mailRenderer(t).Mail("mail/welcome.html", english(t), "https://example.com", struct{ Name string }{Name: `Jan & "Eva"`})
	if err != nil {
		t.Fatal(err)
	}

	if want := `en mail.subject [] Jan & "Eva"`; subject != want {
		t.Errorf("subject %q, want %q", subject, want)
	}

	if want := `<title>en mail.subject [] Jan &amp; &#34;Eva&#34;</title><a href="https://example.com/en">en mail.home []</a>`; html != want {
		t.Errorf("got  %s\nwant %s", html, want)
	}
}

func TestMailRefusesAnUnknownPageOrLocale(t *testing.T) {
	if _, _, err := mailRenderer(t).Mail("mail/missing.html", english(t), "https://example.com", nil); errors.Is(err, errUnknownPage) == false {
		t.Errorf("unknown page: %v", err)
	}

	german, err := posix.Parse("de_DE")
	if err != nil {
		t.Fatal(err)
	}

	if _, _, err := mailRenderer(t).Mail("mail/welcome.html", german, "https://example.com", nil); errors.Is(err, errUnknownLocale) == false {
		t.Errorf("unknown locale: %v", err)
	}
}

func mailRenderer(t *testing.T) *Renderer {
	t.Helper()

	renderer, err := Parse(mails, nil, texts)
	if err != nil {
		t.Fatal(err)
	}

	return renderer
}

func english(t *testing.T) posix.Locale {
	t.Helper()

	l, err := posix.Parse("en_US")
	if err != nil {
		t.Fatal(err)
	}

	return l
}
