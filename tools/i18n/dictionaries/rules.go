package dictionaries

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sync"

	"golang.org/x/text/language"

	"gokick/app/core/i18n/icu"
)

var ErrMissingCategory = errors.New("a plural lacks a CLDR category of its language")

var ErrUnusedCategory = errors.New("a plural has a CLDR category its language never picks")

var categoryCache = struct {
	sync.Mutex

	byTag map[language.Tag][]string
}{byTag: map[language.Tag][]string{}}

func categoriesOf(tag language.Tag) []string {
	categoryCache.Lock()
	defer categoryCache.Unlock()

	found, ok := categoryCache.byTag[tag]
	if ok == false {
		found = icu.NewFormatter(tag).Categories()
		categoryCache.byTag[tag] = found
	}

	return found
}

func categoryIssues(m icu.Message, want []string) []error {
	var issues []error

	eachPlural(m, func(p icu.Plural) {
		for _, category := range want {
			if _, ok := p.Categories[category]; ok == false {
				issues = append(issues, fmt.Errorf("%w: %s lacks %s", ErrMissingCategory, p.Name, category))
			}
		}

		for _, category := range slices.Sorted(maps.Keys(p.Categories)) {
			if slices.Contains(want, category) == false {
				issues = append(issues, fmt.Errorf("%w: %s has %s", ErrUnusedCategory, p.Name, category))
			}
		}
	})

	return issues
}

func eachPlural(m icu.Message, visit func(icu.Plural)) {
	for _, part := range m {
		switch part := part.(type) {
		case icu.Plural:
			visit(part)

			for _, c := range casesOf(part.Exact) {
				eachPlural(c, visit)
			}

			for _, c := range casesOf(part.Categories) {
				eachPlural(c, visit)
			}
		case icu.Select:
			for _, c := range casesOf(part.Cases) {
				eachPlural(c, visit)
			}
		}
	}
}

func casesOf[K cmp.Ordered](cases map[K]icu.Message) []icu.Message {
	messages := make([]icu.Message, 0, len(cases))
	for _, key := range slices.Sorted(maps.Keys(cases)) {
		messages = append(messages, cases[key])
	}

	return messages
}
