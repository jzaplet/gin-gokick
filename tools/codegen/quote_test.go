package codegen_test

import (
	"testing"

	"gokick/tools/codegen"
)

func TestQuoteWritesASingleQuotedTypeScriptString(t *testing.T) {
	for in, want := range map[string]string{
		"Přihlášení": `'Přihlášení'`,

		"it's": `'it\'s'`,

		`a\b`: `'a\\b'`,

		"a\nb\rc": `'a\nb\rc'`,

		"a\u00a0b": `'a\u00a0b'`,

		"a\u2028b": `'a\u2028b'`,

		"a\U000E0001b": `'a\u{e0001}b'`,
	} {
		if got := codegen.Quote(in); got != want {
			t.Errorf("Quote(%q) = %s, want %s", in, got, want)
		}
	}
}
