package csp

import (
	"slices"
	"strings"
	"testing"
)

func TestAnEmptyPolicyAllowsOnlyTheSiteAndTheNonce(t *testing.T) {
	policy := Policy{}

	want := "default-src 'none'; script-src 'nonce-abc' 'strict-dynamic'; style-src 'self' 'nonce-abc'; img-src 'self' data: https:; font-src 'self'; connect-src 'self' https:; manifest-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'; require-trusted-types-for 'script'; trusted-types 'none'"
	if got := policy.Header().Value("abc"); got != want {
		t.Errorf("got  %s\nwant %s", got, want)
	}
}

func TestServicesAddTheirSourcesOnce(t *testing.T) {
	analytics := Service{ScriptSrc: []string{"https://a.example"}, TrustedTypes: []string{"a", "'allow-duplicates'"}}
	widget := Service{
		ScriptSrc: []string{"https://a.example", "https://w.example"},

		FrameSrc: []string{"https://w.example"},

		TrustedTypes: []string{"w", "'allow-duplicates'"},
	}

	empty := directives(Policy{}.Header().Value("abc"))

	got := directives(Policy{Services: []Service{analytics, widget}}.Header().Value("abc"))
	for name, want := range map[string][]string{
		"script-src": {"https://a.example", "https://w.example"},

		"frame-src": {"https://w.example"},

		"trusted-types": {"'allow-duplicates'", "a", "w"},
	} {
		added := slices.DeleteFunc(got[name], func(source string) bool { return slices.Contains(empty[name], source) })
		slices.Sort(added)

		if slices.Equal(added, want) == false {
			t.Errorf("%s adds %v, want %v", name, added, want)
		}
	}
}

func directives(header string) map[string][]string {
	all := map[string][]string{}

	for directive := range strings.SplitSeq(header, "; ") {
		fields := strings.Fields(directive)
		all[fields[0]] = fields[1:]
	}

	return all
}
