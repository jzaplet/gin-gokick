package posix

import (
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

var pattern = regexp.MustCompile(`^[a-z]{2}_[A-Z]{2}$`)

type Locale struct {
	name string

	tag language.Tag
}

func Parse(name string) (Locale, error) {
	bcp47 := strings.Replace(name, "_", "-", 1)
	tag, err := language.Parse(bcp47)

	region, _ := tag.Region()
	if pattern.MatchString(name) == false || err != nil || tag.String() != bcp47 || region.IsCountry() == false {
		return Locale{}, fmt.Errorf("%q is not a locale xx_YY of a known language and country", name)
	}

	return Locale{name: name, tag: tag}, nil
}

func (l Locale) String() string {
	return l.name
}

func (l Locale) Tag() language.Tag {
	return l.tag
}

func (l Locale) Language() string {
	lang, _, _ := strings.Cut(l.name, "_")

	return lang
}

func (l Locale) Region() string {
	_, region, _ := strings.Cut(l.name, "_")

	return region
}
