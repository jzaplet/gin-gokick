package icu

import (
	"slices"
	"testing"

	"golang.org/x/text/language"
)

func TestFormatterListsTheCategoriesOfItsLanguage(t *testing.T) {
	for tag, want := range map[string][]string{
		"cs-CZ": {"one", "few", "many", "other"},

		"en-US": {"one", "other"},

		"pl-PL": {"one", "few", "many", "other"},

		"ar-EG": {"zero", "one", "two", "few", "many", "other"},

		"ja-JP": {"other"},
	} {
		t.Run(tag, func(t *testing.T) {
			got := NewFormatter(language.MustParse(tag)).Categories()

			if slices.Equal(got, want) == false {
				t.Errorf("Categories = %v, want %v", got, want)
			}
		})
	}
}
