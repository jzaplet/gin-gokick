package icu

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"

	"golang.org/x/text/language"
)

type count int64

type gender string

const englishFiles = "{n, plural, one {# file} other {# files}}"

const czechFiles = "{n, plural, one {# soubor} few {# soubory} other {# souborů}}"

const polishFiles = "{n, plural, one {# plik} few {# pliki} many {# plików} other {# pliku}}"

const anyPlural = "{n, plural, other {#}}"

const czechOthers = "{n, plural, offset:1 =0 {nikdo} =1 {jen vy} one {vy a # další} few {vy a # další} other {vy a # dalších}}"

func TestFormatterFillsTheArguments(t *testing.T) {
	t.Run("a text", func(t *testing.T) {
		formats(t, "cs-CZ", "Ahoj {name}!", map[string]any{"name": "Jano"}, "Ahoj Jano!")
	})
	t.Run("a text of a named string type", func(t *testing.T) {
		formats(t, "cs-CZ", "{g}", map[string]any{"g": gender("female")}, "female")
	})
	t.Run("a select case", func(t *testing.T) {
		formats(t, "cs-CZ", "{g, select, female {Napsala} other {Napsal}}", map[string]any{"g": "female"}, "Napsala")
	})
	t.Run("the other case of a select", func(t *testing.T) {
		formats(t, "cs-CZ", "{g, select, female {Napsala} other {Napsal}}", map[string]any{"g": "unknown"}, "Napsal")
	})
	t.Run("a pound in a select inside a plural", func(t *testing.T) {
		formats(t, "cs-CZ", "{n, plural, other {# {g, select, other {#}}}}", map[string]any{"n": 3, "g": "a"}, "3 #")
	})
	t.Run("a pound of the inner plural", func(t *testing.T) {
		formats(t, "cs-CZ", "{n, plural, other {# {g, select, other {{m, plural, other {#}}}}}}", map[string]any{
			"n": 3,

			"g": "a",

			"m": 7,
		}, "3 7")
	})
}

func TestFormatterPicksThePluralCaseOfTheLanguage(t *testing.T) {
	for _, tc := range []struct {
		tag string

		source string

		n any

		want string
	}{
		{tag: "cs-CZ", source: czechFiles, n: 0, want: "0 souborů"},
		{tag: "cs-CZ", source: czechFiles, n: 1, want: "1 soubor"},
		{tag: "cs-CZ", source: czechFiles, n: 2, want: "2 soubory"},
		{tag: "cs-CZ", source: czechFiles, n: 4, want: "4 soubory"},
		{tag: "cs-CZ", source: czechFiles, n: 5, want: "5 souborů"},
		{tag: "cs-CZ", source: czechFiles, n: 22, want: "22 souborů"},
		{tag: "cs-CZ", source: czechFiles, n: -2, want: "-2 soubory"},
		{tag: "cs-CZ", source: czechFiles, n: int64(1000), want: "1\u00a0000 souborů"},
		{tag: "cs-CZ", source: czechFiles, n: count(5), want: "5 souborů"},
		{tag: "en-US", source: englishFiles, n: 1, want: "1 file"},
		{tag: "en-US", source: englishFiles, n: 1000, want: "1,000 files"},
		{tag: "en-US", source: englishFiles, n: -1, want: "-1 file"},
		{tag: "pl-PL", source: polishFiles, n: 22, want: "22 pliki"},
		{tag: "pl-PL", source: polishFiles, n: 25, want: "25 plików"},
		{tag: "cs-CZ", source: "{n, plural, =1 {jeden} one {#} other {#}}", n: 1, want: "jeden"},
		{tag: "cs-CZ", source: "{n, plural, other {#}}", n: 1, want: "1"},
		{tag: "cs-CZ", source: czechOthers, n: 0, want: "nikdo"},
		{tag: "cs-CZ", source: czechOthers, n: 1, want: "jen vy"},
		{tag: "cs-CZ", source: czechOthers, n: 2, want: "vy a 1 další"},
		{tag: "cs-CZ", source: czechOthers, n: 3, want: "vy a 2 další"},
		{tag: "cs-CZ", source: czechOthers, n: 6, want: "vy a 5 dalších"},
	} {
		t.Run(tc.tag+" "+tc.want, func(t *testing.T) {
			formats(t, tc.tag, tc.source, map[string]any{"n": tc.n}, tc.want)
		})
	}
}

func TestFormatterRefusesTheArguments(t *testing.T) {
	for _, tc := range []struct {
		name string

		source string

		args map[string]any

		want error
	}{
		{name: "a missing text", source: "Ahoj {name}", args: map[string]any{}, want: ErrMissingArgument},
		{
			name: "a missing text in a case not taken",

			source: "{g, select, a {{name}} other {}}",

			args: map[string]any{"g": "b"},

			want: ErrMissingArgument,
		},
		{name: "an unknown argument", source: "Ahoj", args: map[string]any{"name": "Jano"}, want: ErrUnknownArgument},
		{name: "a number as a text", source: "{name}", args: map[string]any{"name": 1}, want: ErrArgumentType},
		{name: "a nil text", source: "{name}", args: map[string]any{"name": nil}, want: ErrArgumentType},
		{
			name: "a number as a select",

			source: "{g, select, other {x}}",

			args: map[string]any{"g": 1},

			want: ErrArgumentType,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewFormatter(language.Czech).Format(parse(t, tc.source), tc.args)

			if errors.Is(err, tc.want) == false {
				t.Errorf("Format = %q, %v, want %v", got, err, tc.want)
			}
		})
	}
}

func TestFormatterMatchesAnExactCaseOnTheValueBeforeRounding(t *testing.T) {
	for _, tc := range []struct {
		n any

		want string
	}{
		{n: 1.0, want: "jeden"},
		{n: Decimal{Value: 1, Digits: 1}, want: "jeden"},
		{n: 1.0004, want: "1"},
	} {
		t.Run(fmt.Sprint(tc.n), func(t *testing.T) {
			formats(t, "cs-CZ", "{n, plural, =1 {jeden} other {#}}", map[string]any{"n": tc.n}, tc.want)
		})
	}
}

func TestFormatterSubtractsTheOffsetBeforeRounding(t *testing.T) {
	formats(t, "cs-CZ", "{n, plural, offset:1 other {#}}", map[string]any{"n": 2.5}, "1,5")
	formats(t, "cs-CZ", "{n, plural, offset:1 other {#}}", map[string]any{"n": 1.0005}, "0")
}

func TestFormatterRefusesAPluralValue(t *testing.T) {
	for _, tc := range []struct {
		name string

		n any

		want error
	}{
		{name: "a text", n: "1", want: ErrArgumentType},
		{name: "nil", n: nil, want: ErrArgumentType},
		{name: "an unsigned number", n: uint(1), want: ErrArgumentType},
		{name: "a float32", n: float32(1.5), want: ErrArgumentType},
		{name: "NaN", n: math.NaN(), want: ErrArgumentRange},
		{name: "infinity", n: math.Inf(1), want: ErrArgumentRange},
		{name: "a whole number beyond 2^53", n: int64(1 << 53), want: ErrArgumentRange},
		{name: "a decimal beyond 2^53", n: 1e16, want: ErrArgumentRange},
		{name: "a decimal with too many digits", n: Decimal{Value: 1, Digits: 16}, want: ErrArgumentRange},
		{name: "a decimal with negative digits", n: Decimal{Value: 1, Digits: -1}, want: ErrArgumentRange},
		{name: "a fixed decimal beyond 2^53", n: Decimal{Value: 1e16, Digits: 0}, want: ErrArgumentRange},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewFormatter(language.Czech).Format(parse(t, anyPlural), map[string]any{"n": tc.n})

			if errors.Is(err, tc.want) == false {
				t.Errorf("Format = %q, %v, want %v", got, err, tc.want)
			}
		})
	}
}

func TestFormatterFormatsFromManyGoroutines(t *testing.T) {
	formatter := NewFormatter(language.MustParse("cs-CZ"))
	m := parse(t, czechFiles)
	var wg sync.WaitGroup

	for n := range 16 {
		wg.Go(func() {
			for range 50 {
				got, err := formatter.Format(m, map[string]any{"n": 1000 + n})
				if err != nil || got != fmt.Sprintf("1\u00a0%03d souborů", n) {
					t.Errorf("Format(%d) = %q, %v", 1000+n, got, err)

					return
				}
			}
		})
	}

	wg.Wait()
}

func formats(t *testing.T, tag, source string, args map[string]any, want string) {
	t.Helper()
	got, err := NewFormatter(language.MustParse(tag)).Format(parse(t, source), args)

	if err != nil || got != want {
		t.Errorf("Format(%q, %v) = %q, %v, want %q", source, args, got, err, want)
	}
}

func parse(t *testing.T, source string) Message {
	t.Helper()

	m, err := Parse(source)
	if err != nil {
		t.Fatal(err)
	}

	return m
}
