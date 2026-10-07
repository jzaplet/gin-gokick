package icu

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseBuildsTheMessage(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		parses(t, "Ahoj", Message{Text("Ahoj")})
	})
	t.Run("nothing", func(t *testing.T) {
		parses(t, "", nil)
	})
	t.Run("an argument with spaces", func(t *testing.T) {
		parses(t, "Ahoj { name }!", Message{Text("Ahoj "), Arg{Name: "name"}, Text("!")})
	})
	t.Run("an apostrophe in a word", func(t *testing.T) {
		parses(t, "don't", Message{Text("don't")})
	})
	t.Run("a doubled apostrophe", func(t *testing.T) {
		parses(t, "isn''t", Message{Text("isn't")})
	})
	t.Run("an apostrophe at the end", func(t *testing.T) {
		parses(t, "a'", Message{Text("a'")})
	})
	t.Run("a quoted brace", func(t *testing.T) {
		parses(t, "'{name}'", Message{Text("{name}")})
	})
	t.Run("a doubled apostrophe inside a quote", func(t *testing.T) {
		parses(t, "'{isn''t}'", Message{Text("{isn't}")})
	})
	t.Run("an apostrophe before a quote", func(t *testing.T) {
		parses(t, "'''{'", Message{Text("'{")})
	})
	t.Run("a pound and an apostrophe before it outside a plural", func(t *testing.T) {
		parses(t, "a '# b #", Message{Text("a '# b #")})
	})
	t.Run("a select", func(t *testing.T) {
		parses(t, "{g, select, female {Ona} other {On}}", Message{Select{
			Name: "g",

			Cases: map[string]Message{"female": {Text("Ona")}, other: {Text("On")}},
		}})
	})
	t.Run("a plural with an offset and exact cases", func(t *testing.T) {
		parses(t, "{n, plural, offset:1 =0 {nikdo} one {# další} other {# dalších}}", Message{Plural{
			Name: "n",

			Offset: 1,

			Exact: map[int]Message{0: {Text("nikdo")}},

			Categories: map[string]Message{"one": {Pound{}, Text(" další")}, other: {Pound{}, Text(" dalších")}},
		}})
	})
	t.Run("an offset with a space after its colon", func(t *testing.T) {
		parses(t, "{n, plural,offset: 1 other {#}}", Message{Plural{
			Name: "n",

			Offset: 1,

			Exact: map[int]Message{},

			Categories: map[string]Message{other: {Pound{}}},
		}})
	})
	t.Run("a quoted pound in a plural", func(t *testing.T) {
		parses(t, "{n, plural, other {'#' x}}", Message{Plural{
			Name: "n",

			Exact: map[int]Message{},

			Categories: map[string]Message{other: {Text("# x")}},
		}})
	})
	t.Run("doubled apostrophes around a pound", func(t *testing.T) {
		parses(t, "{n, plural, other {''#''}}", Message{Plural{
			Name: "n",

			Exact: map[int]Message{},

			Categories: map[string]Message{other: {Text("'"), Pound{}, Text("'")}},
		}})
	})
	t.Run("a quoted closing brace in a plural", func(t *testing.T) {
		parses(t, "{n, plural, other {a '}' b}}", Message{Plural{
			Name: "n",

			Exact: map[int]Message{},

			Categories: map[string]Message{other: {Text("a } b")}},
		}})
	})
	t.Run("a pound in a select inside a plural", func(t *testing.T) {
		parses(t, "{n, plural, other {# {g, select, other {#}}}}", Message{Plural{
			Name: "n",

			Exact: map[int]Message{},

			Categories: map[string]Message{other: {Pound{}, Text(" "), Select{
				Name: "g",

				Cases: map[string]Message{other: {Text("#")}},
			}}},
		}})
	})
}

func TestParseRefusesWhatICU4JAndFormatJSReadDifferently(t *testing.T) {
	for _, source := range []string{"a '<b> c", "a '>' c", "'{a} b", "{n, plural, other {'# x}}", "a } b", "{n, plural, =01 {a} other {b}}", "{g, select, __proto__ {a} other {b}}"} {
		t.Run(source, func(t *testing.T) {
			refuses(t, source, ErrSyntax)
		})
	}
}

func TestParseRefusesBrokenSyntax(t *testing.T) {
	for _, source := range []string{
		"{name",
		"{n, plural, other {x}",
		"{n, plural, other {x",
		"{n, plural other {#}}",
		"{n,}",
		"{n, plural, other x}",
		"{0}",
		"{a-b}",
		"{g, select, a-b {x} other {y}}",
		"{n, plural, one {a} one {b} other {c}}",
		"{n, plural, =1 {a} =1 {b} other {c}}",
		"{g, select, a {x} a {y} other {z}}",
		"{n, plural, one {a}}",
		"{g, select, a {x}}",
		"{n, plural, foo {a} other {b}}",
		"{n, plural, =1.5 {a} other {b}}",
		"{n, plural, =-1 {a} other {b}}",
		"{n, plural, offset:-1 other {#}}",
		"{n, plural, offset:01 other {#}}",
		"{n, plural, =0 {a} =00 {b} other {c}}",
		"{n, plural, =99999999999999999999 {a} other {b}}",
		"{n, plural, =9007199254740992 {a} other {b}}",
		"{n, plural, offset:9007199254740992 other {#}}",
	} {
		t.Run(source, func(t *testing.T) {
			refuses(t, source, ErrSyntax)
		})
	}
}

func TestParseRefusesUnsupportedArgumentTypes(t *testing.T) {
	for _, source := range []string{
		"{n, selectordinal, other {#}}",
		"{n, number}",
		"{n, number, ::currency/EUR}",
		"{d, date}",
		"{d, time}",
		"{n, choice, 0#a|1#b}",
		"{n, spellout}",
	} {
		t.Run(source, func(t *testing.T) {
			refuses(t, source, ErrUnsupported)
		})
	}
}

func TestParseRefusesAnArgumentOfTwoKinds(t *testing.T) {
	for _, source := range []string{
		"{n} {n, plural, other {#}}",
		"{g} {g, select, other {x}}",
		"{g, select, a {{n, plural, other {#}}} other {{n}}}",
	} {
		t.Run(source, func(t *testing.T) {
			refuses(t, source, ErrArgumentKinds)
		})
	}
}

func TestMessageListsItsArguments(t *testing.T) {
	m, err := Parse("{name} má {n, plural, one {# zprávu} other {# zpráv}}{g, select, a { od {sender}} other {}}")
	if err != nil {
		t.Fatal(err)
	}

	kinds, err := m.Args()
	want := map[string]Kind{"name": KindText, "n": KindPlural, "g": KindSelect, "sender": KindText}

	if err != nil || reflect.DeepEqual(kinds, want) == false {
		t.Errorf("args %v, %v", kinds, err)
	}
}

func parses(t *testing.T, source string, want Message) {
	t.Helper()

	got, err := Parse(source)

	if err != nil || reflect.DeepEqual(got, want) == false {
		t.Errorf("Parse(%q) = %#v, %v, want %#v", source, got, err, want)
	}
}

func refuses(t *testing.T, source string, want error) {
	t.Helper()

	got, err := Parse(source)

	if errors.Is(err, want) == false {
		t.Errorf("Parse(%q) = %#v, %v, want %v", source, got, err, want)
	}
}
