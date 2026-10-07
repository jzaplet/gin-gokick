package dialect

import (
	"golang.org/x/text/language"

	"gokick/app/core/locale/posix"
)

type Matcher struct {
	dialects []posix.Locale

	matcher language.Matcher
}

func New(dialects []posix.Locale) *Matcher {
	tags := make([]language.Tag, len(dialects))
	for i, l := range dialects {
		tags[i] = l.Tag()
	}

	return &Matcher{dialects: dialects, matcher: language.NewMatcher(tags)}
}

func (m *Matcher) Match(acceptLanguage string) posix.Locale {
	tags, _, err := language.ParseAcceptLanguage(acceptLanguage)
	if err != nil {
		return m.dialects[0]
	}

	_, index, _ := m.matcher.Match(tags...)

	return m.dialects[index]
}
