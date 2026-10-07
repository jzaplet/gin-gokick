package texts

import (
	"errors"
	"testing"

	"golang.org/x/text/language"

	"gokick/app/core/i18n/icu"
)

func TestTextFillsTheParamsOfItsKey(t *testing.T) {
	texts := czech(t, map[string]string{
		"home.title": "Úvod",

		"files.count": "{n, plural, one {# soubor} few {# soubory} many {# souboru} other {# souborů}} od {name}",
	})

	t.Run("a plain text", func(t *testing.T) {
		shows(t, texts, "Úvod", "home.title")
	})
	t.Run("a plural and a text", func(t *testing.T) {
		shows(t, texts, "3 soubory od Jany", "files.count", "n", 3, "name", "Jany")
	})
	t.Run("the params in any order", func(t *testing.T) {
		shows(t, texts, "1,5 souboru od Jany", "files.count", "name", "Jany", "n", 1.5)
	})
}

func TestTextRefusesWhatItCannotFill(t *testing.T) {
	texts := czech(t, map[string]string{"home.title": "Úvod", "user.greeting": "Ahoj {name}"})

	t.Run("an unknown key", func(t *testing.T) {
		refuses(t, texts, ErrUnknownKey, "home.titel")
	})
	t.Run("a name without its value", func(t *testing.T) {
		refuses(t, texts, ErrPairs, "user.greeting", "name")
	})
	t.Run("a name that is no string", func(t *testing.T) {
		refuses(t, texts, ErrPairs, "user.greeting", 1, "Jano")
	})
	t.Run("a param of the wrong type", func(t *testing.T) {
		refuses(t, texts, icu.ErrArgumentType, "user.greeting", "name", 1)
	})
}

func czech(t *testing.T, dictionary map[string]string) *Texts {
	t.Helper()

	texts, err := New(language.MustParse("cs-CZ"), dictionary)
	if err != nil {
		t.Fatal(err)
	}

	return texts
}

func shows(t *testing.T, texts *Texts, want, key string, pairs ...any) {
	t.Helper()

	if got, err := texts.Text(key, pairs...); err != nil || got != want {
		t.Errorf("Text(%s) = %q, %v, want %q", key, got, err, want)
	}
}

func refuses(t *testing.T, texts *Texts, want error, key string, pairs ...any) {
	t.Helper()

	if got, err := texts.Text(key, pairs...); errors.Is(err, want) == false {
		t.Errorf("Text(%s) = %q, %v, want %v", key, got, err, want)
	}
}
