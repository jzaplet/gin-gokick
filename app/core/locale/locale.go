package locale

import (
	"errors"
	"fmt"
	"slices"

	"gokick/app/core/locale/dialect"
	"gokick/app/core/locale/hreflang"
	"gokick/app/core/locale/posix"
)

var errNoLocales = errors.New("no locales")

type grouped struct {
	locales []posix.Locale

	languages []string

	byLanguage map[string][]posix.Locale
}

type Set struct {
	fallback posix.Locale

	locales []posix.Locale

	languages []string

	dialects map[string]*dialect.Matcher

	switcher []posix.Locale

	homes []Home
}

func New(defaultLocale string, locales []string) (*Set, error) {
	if len(locales) == 0 {
		return nil, errNoLocales
	}

	fallback, err := posix.Parse(defaultLocale)
	if err != nil {
		return nil, err
	}

	g, err := group(locales)
	if err != nil {
		return nil, err
	}

	defaults := g.byLanguage[fallback.Language()]
	switch {
	case slices.Contains(defaults, fallback) == false:
		return nil, fmt.Errorf("the default locale %s is not among the locales", fallback)
	case len(defaults) > 1:
		return nil, fmt.Errorf("the default language %s has more than one locale", fallback.Language())
	}

	dialects := make(map[string]*dialect.Matcher, len(g.byLanguage))
	for lang, parsed := range g.byLanguage {
		dialects[lang] = dialect.New(parsed)
	}

	return &Set{
		fallback: fallback,

		locales: g.locales,

		languages: g.languages,

		dialects: dialects,

		switcher: g.firstDialects(),

		homes: homesOf(g.languages, fallback.Language()),
	}, nil
}

func (s *Set) Locales() []posix.Locale {
	return s.locales
}

func (s *Set) Languages() []posix.Locale {
	return s.switcher
}

func (s *Set) OtherLanguages() []string {
	return slices.DeleteFunc(slices.Clone(s.languages), func(lang string) bool {
		return lang == s.DefaultLanguage()
	})
}

func (s *Set) DefaultLanguage() string {
	return s.fallback.Language()
}

func (s *Set) Nearest(name string) posix.Locale {
	wanted, err := posix.Parse(name)
	if err != nil {
		return s.fallback
	}

	if slices.Contains(s.locales, wanted) {
		return wanted
	}

	if i := slices.IndexFunc(s.locales, func(l posix.Locale) bool { return l.Language() == wanted.Language() }); i >= 0 {
		return s.locales[i]
	}

	return s.fallback
}

func homesOf(languages []string, defaultLanguage string) []Home {
	homes := make([]Home, len(languages))
	for i, lang := range languages {
		homes[i] = Home{Language: lang, Path: hreflang.Path(lang, defaultLanguage, "")}
	}

	return homes
}

func (g grouped) firstDialects() []posix.Locale {
	if len(g.languages) < 2 {
		return nil
	}

	firsts := make([]posix.Locale, len(g.languages))
	for i, lang := range g.languages {
		firsts[i] = g.byLanguage[lang][0]
	}

	return firsts
}

func group(names []string) (grouped, error) {
	g := grouped{byLanguage: map[string][]posix.Locale{}}

	for _, name := range names {
		l, err := posix.Parse(name)
		if err != nil {
			return grouped{}, err
		}

		lang := l.Language()
		if slices.Contains(g.byLanguage[lang], l) {
			return grouped{}, fmt.Errorf("the locale %s twice", l)
		}

		if len(g.byLanguage[lang]) == 0 {
			g.languages = append(g.languages, lang)
		}

		g.byLanguage[lang] = append(g.byLanguage[lang], l)
		g.locales = append(g.locales, l)
	}

	return g, nil
}
