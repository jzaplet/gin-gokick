package dictionaries

import (
	"errors"
	"fmt"
	"regexp"

	"gokick/app/core/i18n/icu"
)

var ErrPreposition = errors.New("a vocalized preposition stands right before #, and only the number decides its form (s 5, se 7): rephrase or write it into =N cases")

var vocalized = regexp.MustCompile(`(?i)(?:^|[^\p{L}])(s|se|k|ke|v|ve|z|ze)[\s\p{Zs}]+$`)

func czechPrepositions(m icu.Message, before string, counted bool) []error {
	var issues []error

	for _, part := range m {
		switch part := part.(type) {
		case icu.Text:
			before = string(part)

			continue
		case icu.Pound:
			if match := vocalized.FindStringSubmatch(before); counted && match != nil {
				issues = append(issues, fmt.Errorf("%w: %q", ErrPreposition, match[1]))
			}
		case icu.Plural:
			for _, c := range casesOf(part.Categories) {
				issues = append(issues, czechPrepositions(c, before, true)...)
			}

			for _, c := range casesOf(part.Exact) {
				issues = append(issues, czechPrepositions(c, before, false)...)
			}
		case icu.Select:
			for _, c := range casesOf(part.Cases) {
				issues = append(issues, czechPrepositions(c, before, counted)...)
			}
		}

		before = ""
	}

	return issues
}
