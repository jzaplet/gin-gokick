package view

import (
	"html"
	"strings"

	"gokick/app/core/locale/posix"
)

const subject = "subject"

type mailData struct {
	Origin string

	Locale posix.Locale

	Params any
}

func (r *Renderer) Mail(name string, l posix.Locale, origin string, params any) (title, body string, err error) {
	data := &mailData{Origin: origin, Locale: l, Params: params}

	page, err := r.execute(name, name, l, data)
	if err != nil {
		return "", "", err
	}

	head, err := r.execute(name, subject, l, data)
	if err != nil {
		return "", "", err
	}

	return strings.TrimSpace(html.UnescapeString(string(head))), string(page), nil
}
