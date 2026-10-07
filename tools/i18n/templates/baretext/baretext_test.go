package baretext

import (
	"errors"
	"strings"
	"testing"
	"text/template/parse"
)

func TestCheckAcceptsTextsFromTheDictionaries(t *testing.T) {
	for _, content := range []string{
		`<!doctype html><html lang="{{.Lang}}"><!-- a note --><title>{{t "home.title"}} | {{t "brand.name"}}</title></html>`,
		`<p>{{t "home.lead"}}&nbsp;– ({{.Count}})</p>`,
		`<script nonce="{{.Nonce}}">const greeting = "Ahoj";</script><style>p::after { content: "Ahoj" }</style>`,
		`<nav aria-labelledby="features" aria-hidden="true" class="flex"><img src="logo.svg" alt=""></nav>`,
		`<meta name="robots" content="noindex"><meta name="description" content="{{template "description" .}}">`,
		`<button title="{{t "a"}}" aria-label="{{if .Open}}{{t "a"}}{{else}}{{t "b"}}{{end}}">{{range .Items}}<li>{{.}}</li>{{end}}</button>`,
		`{{define "title"}}{{t "not_found.title"}} | {{t "brand.name"}}{{end}}`,
		`<li><span aria-hidden="true">01</span>{{t "home.feature"}}</li><p>{{.Count}} / 404</p>`,
		`<div>{{t "mail.preheader"}}&#8203;&#8203;&shy;&zwnj;</div>`,
	} {
		if issues := check(t, content); len(issues) > 0 {
			t.Errorf("Check(%s) = %v", content, issues)
		}
	}
}

func TestCheckRefusesTextOutsideT(t *testing.T) {
	t.Run("text between tags", func(t *testing.T) {
		refuses(t, `<p>Ahoj</p>`, `"Ahoj"`)
	})
	t.Run("text next to an action", func(t *testing.T) {
		refuses(t, `<p>{{t "a"}} světe</p>`, `"{{…}} světe"`)
	})
	t.Run("a number with its word", func(t *testing.T) {
		refuses(t, `<p>3 {{t "files"}}</p><p>4 soubory</p>`, `"4 soubory"`)
	})
	t.Run("text in an else branch", func(t *testing.T) {
		refuses(t, `{{if .Open}}{{t "a"}}{{else}}<p>Zavřeno</p>{{end}}`, `"Zavřeno"`)
	})
	t.Run("text in a define", func(t *testing.T) {
		refuses(t, `{{define "title"}}Úvod | {{t "brand.name"}}{{end}}`, `"Úvod | {{…}}"`)
	})
	t.Run("text in the title", func(t *testing.T) {
		refuses(t, `<title>Úvod</title>`, `"Úvod"`)
	})
	t.Run("the description", func(t *testing.T) {
		refuses(t, `<meta name="description" content="Kostra pro Golang">`, `content="Kostra pro Golang"`)
	})
	t.Run("an attribute next to an action", func(t *testing.T) {
		refuses(t, `<img src="logo.svg" alt="{{t "brand.name"}} logo">`, `alt="{{…}} logo"`)
	})

	for _, name := range textAttributes {
		t.Run(name, func(t *testing.T) {
			refuses(t, `<input `+name+`="Ahoj">`, name+`="Ahoj"`)
		})
	}
}

func TestCheckNamesTheFileLineAndColumnOfAText(t *testing.T) {
	for content, want := range map[string]string{
		"<p>\n  <span>{{t \"a\"}}</span>\n  Ahoj</p>": "page.html:3:2: ",

		"{{if .Open}}\n{{else}}\n<nav class=\"flex\"\n aria-label=\"GitHub\">{{end}}": "page.html:3:0: ",

		"{{define \"content\"}}\n<p>{{t \"a\"}}</p>{{end}}{{define \"title\"}}<b>Úvod</b>{{end}}": "page.html:2:44: ",
	} {
		issues := check(t, content)
		if len(issues) != 1 || strings.HasPrefix(issues[0].Error(), want) == false {
			t.Errorf("Check(%q) = %v, want %s", content, issues, want)
		}
	}
}

func refuses(t *testing.T, content, want string) {
	t.Helper()

	issues := check(t, content)
	if len(issues) != 1 || errors.Is(issues[0], ErrBareText) == false || strings.HasSuffix(issues[0].Error(), ": "+want) == false {
		t.Errorf("Check(%s) = %v, want %s", content, issues, want)
	}
}

func check(t *testing.T, content string) []error {
	t.Helper()

	file := parse.New("page.html")
	file.Mode = parse.SkipFuncCheck

	trees := map[string]*parse.Tree{}
	if _, err := file.Parse(content, "", "", trees); err != nil {
		t.Fatal(err)
	}

	var issues []error
	for _, tree := range trees {
		issues = append(issues, Check(tree)...)
	}

	return issues
}
