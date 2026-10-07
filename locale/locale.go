package locale

import (
	"errors"
	"fmt"
	"maps"
)

var ErrNoDictionary = errors.New("the locale has no dictionary")

const TypeScriptDir = "assets/shared/I18n/Dictionary"

const NameKey = "language.name"

type Dictionary map[string]string

var dictionaries = map[string]Dictionary{}

func register(name string, d Dictionary) {
	dictionaries[name] = d
}

func All() map[string]Dictionary {
	return maps.Clone(dictionaries)
}

func Require(locales []string) error {
	for _, name := range locales {
		if _, ok := dictionaries[name]; ok == false {
			return fmt.Errorf("%w: %s", ErrNoDictionary, name)
		}
	}

	return nil
}

func Module(name string) string {
	return TypeScriptDir + "/Locales/" + name + ".ts"
}
