package frontend

import (
	"cmp"
	"slices"
	"testing"
)

func TestLexFindsTheStringsOfAScript(t *testing.T) {
	for src, want := range map[string][]piece{
		`t('login.title'); const label = "form.email";`: {{text: "login.title"}, {text: "form.email"}},

		`const quote = 'it\'s'; const other = "a\"b";`: {{text: `it\'s`}, {text: `a\"b`}},

		"// t('commented')\n/* t(\"blocked\") */ t('kept')": {{text: "kept"}},

		`const apostrophe = /'/g; return /"[']"/.test(x) ? 'yes' : 'no';`: {{text: "yes"}, {text: "no"}},

		"const half = width / 2; const third = (a) / 3; t('after.division')": {{text: "after.division"}},

		"const plain = `home.title`;": {{text: "home.title"}},

		"t(`tracking.${tool}`)": {{text: "tracking.", built: true}},

		"`${t('inner.key')} and ${ {a: 'object'}.a }` + 'tail'": {{
			text: "",

			built: true,
		}, {text: "inner.key"}, {text: "object"}, {text: "tail"}},

		"const broken = 'no end\nt('next.line')": {{text: "no end"}, {text: "next.line"}},

		"const last = count++ / 2; t('last.line')": {{text: "last.line"}},
	} {
		got := lex(src)
		slices.SortFunc(got, func(a, b piece) int { return cmp.Compare(a.at, b.at) })

		for i := range got {
			got[i].at = 0
		}

		if slices.Equal(got, want) == false {
			t.Errorf("lex(%s) = %v, want %v", src, got, want)
		}
	}
}
