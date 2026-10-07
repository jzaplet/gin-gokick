package locale

import (
	"errors"
	"slices"
	"testing"
)

func TestNewRefusesInvalidLocales(t *testing.T) {
	tests := []struct {
		name, fallback string

		locales []string
	}{
		{"no locales", "cs_CZ", nil},
		{"an invalid locale", "cs_CZ", []string{"cs_CZ", "en-US"}},
		{"an invalid default", "cs", []string{"cs_CZ"}},
		{"a locale twice", "cs_CZ", []string{"cs_CZ", "en_US", "en_US"}},
		{"a default outside the locales", "cs_CZ", []string{"en_US"}},
		{"a second dialect of the default language", "cs_CZ", []string{"cs_CZ", "en_US", "cs_SK"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New(tt.fallback, tt.locales); err == nil {
				t.Errorf("%s with %q passed", tt.fallback, tt.locales)
			}
		})
	}

	if _, err := New("cs_CZ", nil); errors.Is(err, errNoLocales) == false {
		t.Errorf("no locales: %v", err)
	}
}

func TestLocalesAreEveryLocaleInItsOrder(t *testing.T) {
	set, err := New("cs_CZ", []string{"en_GB", "cs_CZ", "en_US"})
	if err != nil {
		t.Fatal(err)
	}

	var names []string
	for _, l := range set.Locales() {
		names = append(names, l.String())
	}

	if slices.Equal(names, []string{"en_GB", "cs_CZ", "en_US"}) == false {
		t.Errorf("locales %v", names)
	}
}

func TestOneLanguageHasNoSwitcher(t *testing.T) {
	set, err := New("cs_CZ", []string{"cs_CZ"})
	if err != nil {
		t.Fatal(err)
	}

	if languages := set.Languages(); languages != nil {
		t.Errorf("languages %v", languages)
	}
}

func TestNearestFindsTheLocaleOrItsClosestOne(t *testing.T) {
	set, err := New("cs_CZ", []string{"cs_CZ", "en_GB", "en_US"})
	if err != nil {
		t.Fatal(err)
	}

	for name, want := range map[string]string{
		"en_US": "en_US",

		"en_GB": "en_GB",

		"en_AU": "en_GB",

		"de_DE": "cs_CZ",

		"": "cs_CZ",

		"en": "cs_CZ",

		"en-US": "cs_CZ",
	} {
		if got := set.Nearest(name).String(); got != want {
			t.Errorf("Nearest(%q) = %s, want %s", name, got, want)
		}
	}
}
