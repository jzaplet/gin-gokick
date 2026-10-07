package typescript_test

import (
	"errors"
	"flag"
	"maps"
	"regexp"
	"slices"
	"testing"

	"gokick/app/core/i18n/icu"
	"gokick/app/core/locale/posix"
	"gokick/locale"
	"gokick/tools/codegen/codegentest"
	"gokick/tools/i18n/dictionaries"
	"gokick/tools/i18n/typescript"
)

var update = flag.Bool("update", false, "rewrite testdata from the fixture dictionaries")

var shownCategory = regexp.MustCompile(`(?m)^ {4}\[.+, '([a-z]+) .*'\],$`)

var fixture = map[string]locale.Dictionary{
	"cs_CZ": {
		"files.count": "{n, plural, offset:1 =0 {nikdo} =1 {jen vy} one {vy a # další} few {vy a # další} many {vy a # dalšího} other {vy a # dalších}}",

		"language.name": "Čeština",

		"login.title": "Přihlášení",

		"profile.written": "{g, select, female {Napsala} 1st {Napsal} other {Napsalo}} {name}",

		"quotes.mixed": "it''s '{literal}'\u00a0a\\b",
	},

	"en_US": {
		"files.count": "{n, plural, offset:1 =0 {nobody} one {you and # other} other {you and # others}}",

		"language.name": "English",

		"login.title": "Sign in",

		"profile.written": "{g, select, female {She wrote} 1st {He wrote} other {They wrote}} {name}",

		"quotes.mixed": "",
	},
}

func TestRenderMatchesTheGoldenFiles(t *testing.T) {
	c, err := dictionaries.Build(fixture)
	if err != nil {
		t.Fatal(err)
	}

	files, err := typescript.Render(c)
	if err != nil {
		t.Fatal(err)
	}

	codegentest.MatchGolden(t, "testdata", files.Content, *update)
}

func TestRenderWantsTheNameOfEveryLanguageAsPlainText(t *testing.T) {
	for name, source := range map[string]string{
		"no name": "",

		"a name with an argument": "{name}",
	} {
		t.Run(name, func(t *testing.T) {
			c, err := dictionaries.Build(map[string]locale.Dictionary{"cs_CZ": {"language.name": source}})
			if err != nil {
				t.Fatal(err)
			}

			if _, err := typescript.Render(c); errors.Is(err, typescript.ErrLanguageName) == false {
				t.Errorf("Render = %v", err)
			}
		})
	}
}

func TestNumbersReachEveryCategoryOfALanguage(t *testing.T) {
	names := []string{"ar_EG", "cs_CZ", "cy_GB", "en_US", "fr_FR", "he_IL", "lt_LT", "lv_LV", "pl_PL", "ru_RU", "sl_SI"}

	sources := map[string]locale.Dictionary{}
	for _, name := range names {
		sources[name] = locale.Dictionary{"a.b": "x"}
	}

	c, err := dictionaries.Build(sources)
	if err != nil {
		t.Fatal(err)
	}

	files, err := typescript.Numbers(c)
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range names {
		parsed, err := posix.Parse(name)
		if err != nil {
			t.Fatal(err)
		}

		found := map[string]bool{}
		for _, match := range shownCategory.FindAllSubmatch(files.Content[typescript.NumbersDir+"/"+name+".ts"], -1) {
			found[string(match[1])] = true
		}

		if want := icu.NewFormatter(parsed.Tag()).Categories(); slices.Equal(slices.Sorted(maps.Keys(found)), slices.Sorted(slices.Values(want))) == false {
			t.Errorf("%s shows %v, its language has %v", name, slices.Sorted(maps.Keys(found)), want)
		}
	}
}
