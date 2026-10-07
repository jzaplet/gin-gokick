package dictionaries

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"

	"gokick/app/core/i18n/icu"
	"gokick/app/core/locale/posix"
	"gokick/locale"
)

var ErrNoDictionaries = errors.New("no dictionaries")

var ErrLocale = errors.New("not a locale")

var ErrKey = errors.New("a key is not words joined by dots")

var ErrMissingKey = errors.New("a dictionary lacks a key the others have")

var ErrMessage = errors.New("a message is no ICU MessageFormat of the supported subset")

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:_[a-z0-9]+)*(?:\.[a-z][a-z0-9]*(?:_[a-z0-9]+)*)+$`)

type Locale struct {
	Name string

	Messages map[string]icu.Message
}

type Catalog struct {
	Locales []Locale

	Keys []string

	Args map[string]map[string]icu.Kind
}

func Build(dictionaries map[string]locale.Dictionary) (Catalog, error) {
	if len(dictionaries) == 0 {
		return Catalog{}, ErrNoDictionaries
	}

	var issues []error
	keys := map[string]bool{}

	for _, d := range dictionaries {
		for key := range d {
			keys[key] = true
		}
	}

	c := Catalog{Keys: slices.Sorted(maps.Keys(keys))}
	for _, name := range slices.Sorted(maps.Keys(dictionaries)) {
		l, found := parse(name, dictionaries[name], c.Keys)
		c.Locales = append(c.Locales, l)
		issues = append(issues, found...)
	}

	for _, key := range c.Keys {
		if keyPattern.MatchString(key) == false {
			issues = append(issues, fmt.Errorf("%w: %q", ErrKey, key))
		}
	}

	if len(issues) == 0 {
		c.Args, issues = sameArgs(c)
	}

	if len(issues) > 0 {
		return Catalog{}, errors.Join(issues...)
	}

	return c, nil
}

func parse(name string, d locale.Dictionary, keys []string) (Locale, []error) {
	parsed, err := posix.Parse(name)
	if err != nil {
		return Locale{}, []error{fmt.Errorf("%w: %w", ErrLocale, err)}
	}

	var issues []error
	l := Locale{Name: name, Messages: map[string]icu.Message{}}
	categories := categoriesOf(parsed.Tag())

	for _, key := range keys {
		source, ok := d[key]
		if ok == false {
			issues = append(issues, fmt.Errorf("%w: %s lacks %s", ErrMissingKey, name, key))

			continue
		}

		m, err := icu.Parse(source)
		if err != nil {
			issues = append(issues, fmt.Errorf("%w: %s %s: %w", ErrMessage, name, key, err))

			continue
		}

		l.Messages[key] = m

		found := categoryIssues(m, categories)
		if parsed.Language() == "cs" {
			found = append(found, czechPrepositions(m, "", false)...)
		}

		for _, issue := range found {
			issues = append(issues, fmt.Errorf("%s %s: %w", name, key, issue))
		}
	}

	return l, issues
}
