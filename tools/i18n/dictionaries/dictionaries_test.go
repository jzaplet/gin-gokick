package dictionaries_test

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"gokick/app/core/i18n/icu"
	"gokick/locale"
	"gokick/tools/i18n/dictionaries"
)

const czechFiles = "{n, plural, one {# soubor} few {# soubory} many {# souboru} other {# souborů}}"

const englishFiles = "{n, plural, one {# file} other {# files}}"

func TestBuildParsesEveryDictionary(t *testing.T) {
	c, err := dictionaries.Build(map[string]locale.Dictionary{
		"en_US": {"files.count": englishFiles, "login.title": "Sign in"},

		"cs_CZ": {"files.count": czechFiles, "login.title": "Přihlášení"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if names := []string{c.Locales[0].Name, c.Locales[1].Name}; slices.Equal(names, []string{"cs_CZ", "en_US"}) == false {
		t.Errorf("locales %v", names)
	}

	if slices.Equal(c.Keys, []string{"files.count", "login.title"}) == false {
		t.Errorf("keys %v", c.Keys)
	}

	if want := map[string]map[string]icu.Kind{
		"files.count": {"n": icu.KindPlural},

		"login.title": {},
	}; reflect.DeepEqual(c.Args, want) == false {
		t.Errorf("args %v", c.Args)
	}

	if got := c.Locales[0].Messages["login.title"]; reflect.DeepEqual(got, icu.Message{icu.Text("Přihlášení")}) == false {
		t.Errorf("message %#v", got)
	}
}

func TestBuildRefusesWhatIsNoDictionary(t *testing.T) {
	t.Run("no dictionaries", func(t *testing.T) {
		refuses(t, map[string]locale.Dictionary{}, dictionaries.ErrNoDictionaries)
	})
	t.Run("a name that is no locale", func(t *testing.T) {
		refuses(t, map[string]locale.Dictionary{"cs": {"a.b": "x"}}, dictionaries.ErrLocale)
	})
}

func TestBuildRefusesAKeyOfTheWrongShape(t *testing.T) {
	for _, key := range []string{"login", "Login.title", "login.Title", "login..title", "login.title-x", "login._title", "1a.title", "login.title_", "login.title."} {
		t.Run(key, func(t *testing.T) {
			refuses(t, map[string]locale.Dictionary{"cs_CZ": {key: "x"}}, dictionaries.ErrKey)
		})
	}
}

func TestBuildRefusesDictionariesThatDiffer(t *testing.T) {
	t.Run("a missing key", func(t *testing.T) {
		refuses(t, map[string]locale.Dictionary{
			"cs_CZ": {"a.b": "x", "a.c": "y"},

			"en_US": {"a.b": "x"},
		}, dictionaries.ErrMissingKey)
	})
	t.Run("an argument of another kind", func(t *testing.T) {
		refuses(t, map[string]locale.Dictionary{
			"cs_CZ": {"a.b": "{n}"},

			"en_US": {"a.b": englishFiles},
		}, dictionaries.ErrArgs)
	})
	t.Run("an argument of another name", func(t *testing.T) {
		refuses(t, map[string]locale.Dictionary{
			"cs_CZ": {"a.b": "{name}"},

			"en_US": {"a.b": "{user}"},
		}, dictionaries.ErrArgs)
	})
	t.Run("a missing argument", func(t *testing.T) {
		refuses(t, map[string]locale.Dictionary{
			"cs_CZ": {"a.b": "{name}"},

			"en_US": {"a.b": "Hi"},
		}, dictionaries.ErrArgs)
	})
}

func TestBuildRefusesAMessageOutsideTheSubset(t *testing.T) {
	refuses(t, map[string]locale.Dictionary{"cs_CZ": {"a.b": "{n, number}"}}, dictionaries.ErrMessage)
}

func TestBuildWantsEveryCategoryOfTheLanguage(t *testing.T) {
	for name, tc := range map[string]struct {
		dictionaries map[string]locale.Dictionary

		want error
	}{
		"a Czech plural without many": {
			dictionaries: map[string]locale.Dictionary{"cs_CZ": {"a.b": "{n, plural, one {#} few {#} other {#}}"}},

			want: dictionaries.ErrMissingCategory,
		},
		"an English plural with few": {
			dictionaries: map[string]locale.Dictionary{"en_US": {"a.b": "{n, plural, one {#} few {#} other {#}}"}},

			want: dictionaries.ErrUnusedCategory,
		},
		"a plural inside a select": {
			dictionaries: map[string]locale.Dictionary{"cs_CZ": {"a.b": "{g, select, a {{n, plural, one {#} other {#}}} other {x}}"}},

			want: dictionaries.ErrMissingCategory,
		},
		"a plural inside an exact case": {
			dictionaries: map[string]locale.Dictionary{"en_US": {"a.b": "{n, plural, =0 {{m, plural, other {#}}} one {#} other {#}}"}},

			want: dictionaries.ErrMissingCategory,
		},
	} {
		t.Run(name, func(t *testing.T) {
			refuses(t, tc.dictionaries, tc.want)
		})
	}
}

func TestBuildRefusesAVocalizedPrepositionBeforeACountedNumber(t *testing.T) {
	for _, source := range []string{
		"s {n, plural, one {# souborem} few {# soubory} many {# souboru} other {# soubory}}",
		"{n, plural, one {se # souborem} few {se # soubory} many {se # souboru} other {se # soubory}}",
		"{n, plural, one {s\u00a0# souborem} few {#} many {#} other {#}}",
		"{n, plural, one {K #} few {#} many {#} other {#}}",
		"{n, plural, one {v #} few {#} many {#} other {#}}",
		"{n, plural, one {ze #} few {#} many {#} other {#}}",
		"{n, plural, =1 {jeden} one {#} few {#} many {#} other {(s #)}}",
	} {
		t.Run(source, func(t *testing.T) {
			refuses(t, map[string]locale.Dictionary{"cs_CZ": {"a.b": source}}, dictionaries.ErrPreposition)
		})
	}
}

func TestBuildLetsAPrepositionStandWhereTheNumberIsFixed(t *testing.T) {
	for _, source := range []string{
		"{n, plural, =7 {se 7 soubory} one {# soubor} few {# soubory} many {# souboru} other {# souborů}}",
		"{n, plural, =1 {s #} one {jeden} few {#} many {#} other {#}}",
		"{n, plural, one {kus #} few {#} many {#} other {#}}",
		"{n, plural, one {s {name} #} few {#} many {#} other {#}}",
		"se {name}",
	} {
		t.Run(source, func(t *testing.T) {
			if _, err := dictionaries.Build(map[string]locale.Dictionary{"cs_CZ": {"a.b": source}}); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestBuildChecksPrepositionsOnlyInCzech(t *testing.T) {
	if _, err := dictionaries.Build(map[string]locale.Dictionary{"en_US": {"a.b": "{n, plural, one {s #} other {s #}}"}}); err != nil {
		t.Error(err)
	}
}

func TestBuildReportsEveryIssue(t *testing.T) {
	_, err := dictionaries.Build(map[string]locale.Dictionary{"cs_CZ": {"Bad": "x", "a.b": "{n, plural, other {#}}"}})

	if errors.Is(err, dictionaries.ErrKey) == false || errors.Is(err, dictionaries.ErrMissingCategory) == false {
		t.Errorf("Build = %v", err)
	}
}

func refuses(t *testing.T, sources map[string]locale.Dictionary, want error) {
	t.Helper()

	c, err := dictionaries.Build(sources)

	if errors.Is(err, want) == false {
		t.Errorf("Build = %+v, %v, want %v", c, err, want)
	}
}
