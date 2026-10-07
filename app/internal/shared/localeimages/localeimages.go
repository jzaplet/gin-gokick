package localeimages

import (
	"strings"

	"gokick/app/core/locale/posix"
)

func Flag(l posix.Locale) string {
	return "assets/img/flags/" + strings.ToLower(l.Region()) + ".svg"
}

func Flags(languages []posix.Locale) []string {
	return each(languages, Flag)
}

func OGImage(l posix.Locale) string {
	return "assets/img/og/" + l.String() + ".png"
}

func OGImages(locales []posix.Locale) []string {
	return each(locales, OGImage)
}

func each(locales []posix.Locale, file func(posix.Locale) string) []string {
	files := make([]string, len(locales))
	for i, l := range locales {
		files[i] = file(l)
	}

	return files
}
