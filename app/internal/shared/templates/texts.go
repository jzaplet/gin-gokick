package templates

import (
	"fmt"

	"gokick/app/core/i18n/texts"
	"gokick/app/core/locale/posix"
	"gokick/app/core/view"
	dictionaries "gokick/locale"
)

func translations(locales []posix.Locale) (map[string]view.Translate, error) {
	all := dictionaries.All()

	byLocale := make(map[string]view.Translate, len(locales))
	for _, l := range locales {
		dictionary, ok := all[l.String()]
		if ok == false {
			return nil, fmt.Errorf("%w: %s", dictionaries.ErrNoDictionary, l)
		}

		localized, err := texts.New(l.Tag(), dictionary)
		if err != nil {
			return nil, err
		}

		byLocale[l.String()] = localized.Text
	}

	return byLocale, nil
}

func languageName(byLocale map[string]view.Translate) func(posix.Locale) (string, error) {
	return func(l posix.Locale) (string, error) {
		translate, ok := byLocale[l.String()]
		if ok == false {
			return "", fmt.Errorf("%w: %s", dictionaries.ErrNoDictionary, l)
		}

		return translate(dictionaries.NameKey)
	}
}

func dictionaryModules(locales []posix.Locale) []string {
	modules := make([]string, len(locales))
	for i, l := range locales {
		modules[i] = dictionaries.Module(l.String())
	}

	return modules
}
