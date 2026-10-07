package tags

import (
	"html/template"
	"slices"
	"strings"
)

type Set struct {
	styles, preloads, scripts []string
}

func (s *Set) Add(href string) {
	if strings.HasSuffix(href, ".css") {
		s.Style(href)
	} else {
		push(&s.scripts, href)
	}
}

func (s *Set) Style(href string) {
	push(&s.styles, href)
}

func (s *Set) Preload(href string) bool {
	return push(&s.preloads, href)
}

func (s *Set) Merge(other Set) {
	for _, href := range other.styles {
		push(&s.styles, href)
	}

	for _, href := range other.preloads {
		push(&s.preloads, href)
	}

	for _, href := range other.scripts {
		push(&s.scripts, href)
	}
}

func (s *Set) HTML(nonce string) template.HTML {
	nonce = template.HTMLEscapeString(nonce)

	var out strings.Builder
	for _, href := range s.styles {
		out.WriteString(`<link rel="stylesheet" href="` + template.HTMLEscapeString(href) + "\">\n")
	}

	for _, href := range s.preloads {
		out.WriteString(`<link rel="modulepreload" nonce="` + nonce + `" href="` + template.HTMLEscapeString(href) + "\">\n")
	}

	for _, href := range s.scripts {
		out.WriteString(`<script type="module" nonce="` + nonce + `" src="` + template.HTMLEscapeString(href) + "\"></script>\n")
	}

	return template.HTML(out.String())
}

func push(list *[]string, href string) bool {
	if slices.Contains(*list, href) {
		return false
	}

	*list = append(*list, href)

	return true
}
