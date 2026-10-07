package dialect

import (
	"testing"

	"gokick/app/core/locale/posix"
)

func TestMatchPicksTheDialectOfTheBrowser(t *testing.T) {
	matcher := New(locales(t, "en_GB", "en_US"))
	for header, want := range map[string]string{
		"": "en_GB",

		"de": "en_GB",

		"en-US;q=invalid": "en_GB",

		"en-US": "en_US",

		"en": "en_US",

		"en-AU": "en_GB",

		"cs, en-US;q=0.5": "en_US",

		"en-GB;q=0.5, en-US;q=0.9": "en_US",

		"en-US;q=0, en-GB;q=0.1": "en_GB",
	} {
		t.Run(header, func(t *testing.T) {
			if got := matcher.Match(header).String(); got != want {
				t.Errorf("got %s, want %s", got, want)
			}
		})
	}
}

func locales(t *testing.T, names ...string) []posix.Locale {
	t.Helper()

	parsed := make([]posix.Locale, len(names))
	for i, name := range names {
		l, err := posix.Parse(name)
		if err != nil {
			t.Fatal(err)
		}

		parsed[i] = l
	}

	return parsed
}
