package keys

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"gokick/app/core/i18n/icu"
	"gokick/tools/i18n/templates/baretext"
)

var args = map[string]map[string]icu.Kind{
	"home.title": {},

	"home.intro": {},

	"footer.copyright": {"year": icu.KindText},

	"files.count": {"n": icu.KindPlural, "name": icu.KindText},
}

func TestCheckAcceptsTextsWithTheirParamsAndReturnsTheirKeys(t *testing.T) {
	keys, err := Check(fstest.MapFS{
		"shared/layout.html": {Data: []byte(`<title>{{block "title" .}}{{t "home.title"}}{{end}}</title>`)},

		"home/home.html": {Data: []byte(`{{define "content"}}{{if .}}<p>{{t "footer.copyright" "year" year}}</p>{{end}}{{t "files.count" "name" .Name "n" .N}}{{end}}`)},

		"home/components/intro.html": {Data: []byte(`<p>{{t "home.intro"}}</p>`)},

		"notes.txt": {Data: []byte(`{{t "unknown.key"}}`)},
	}, args)
	if want := []string{"home.intro", "footer.copyright", "files.count", "home.title"}; err != nil || slices.Equal(keys, want) == false {
		t.Errorf("Check = %v, %v, want %v", keys, err, want)
	}
}

func TestCheckRefusesTextsTheDictionariesCannotFill(t *testing.T) {
	t.Run("a key in a variable", func(t *testing.T) {
		refuses(t, `{{t .Key}}`, ErrKey)
	})
	t.Run("no key", func(t *testing.T) {
		refuses(t, `{{t}}`, ErrKey)
	})
	t.Run("t as the argument of another func", func(t *testing.T) {
		refuses(t, `{{printf "%s" t}}`, ErrKey)
	})
	t.Run("an unknown key", func(t *testing.T) {
		refuses(t, `{{define "title"}}{{t "home.titel"}}{{end}}`, ErrUnknownKey)
	})
	t.Run("a missing param", func(t *testing.T) {
		refuses(t, `{{t "files.count" "n" 3}}`, ErrParams)
	})
	t.Run("an extra param", func(t *testing.T) {
		refuses(t, `{{t "home.title" "n" 3}}`, ErrParams)
	})
	t.Run("a name without its value", func(t *testing.T) {
		refuses(t, `{{t "footer.copyright" "year"}}`, ErrParams)
	})
	t.Run("a name in a variable", func(t *testing.T) {
		refuses(t, `{{t "footer.copyright" .Name year}}`, ErrParams)
	})
	t.Run("a param twice", func(t *testing.T) {
		refuses(t, `{{t "files.count" "n" 1 "n" 2}}`, ErrParams)
	})
	t.Run("text outside t", func(t *testing.T) {
		refuses(t, `<p>{{t "home.title"}} Ahoj</p>`, baretext.ErrBareText)
	})
	t.Run("text outside t in a component", func(t *testing.T) {
		_, err := Check(fstest.MapFS{"home/components/intro.html": {Data: []byte("<p>\nAhoj</p>")}}, args)
		if errors.Is(err, baretext.ErrBareText) == false || strings.HasPrefix(err.Error(), "home/components/intro.html:2:") == false {
			t.Errorf("Check = %v", err)
		}
	})
	t.Run("a broken template", func(t *testing.T) {
		refuses(t, `{{t "home.title"`, ErrTemplate)
	})
}

func TestCheckNamesTheFileAndLineOfAText(t *testing.T) {
	_, err := Check(fstest.MapFS{"errors/not_found.html": {Data: []byte("<p>\n{{t \"home.titel\"}}</p>")}}, args)

	if err == nil || strings.HasPrefix(err.Error(), "errors/not_found.html:2:") == false {
		t.Errorf("Check = %v", err)
	}
}

func refuses(t *testing.T, content string, want error) {
	t.Helper()

	if _, err := Check(fstest.MapFS{"page.html": {Data: []byte(content)}}, args); errors.Is(err, want) == false {
		t.Errorf("Check(%s) = %v, want %v", content, err, want)
	}
}
