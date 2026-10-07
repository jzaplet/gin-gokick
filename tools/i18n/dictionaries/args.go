package dictionaries

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"gokick/app/core/i18n/icu"
)

var ErrArgs = errors.New("a message takes other arguments than in the first dictionary")

func sameArgs(c Catalog) (map[string]map[string]icu.Kind, []error) {
	args := map[string]map[string]icu.Kind{}
	var issues []error

	for _, key := range c.Keys {
		for i, l := range c.Locales {
			kinds, err := l.Messages[key].Args()
			switch {
			case err != nil:
				issues = append(issues, fmt.Errorf("%w: %s %s: %w", ErrMessage, l.Name, key, err))
			case i == 0:
				args[key] = kinds
			case maps.Equal(kinds, args[key]) == false:
				issues = append(issues, fmt.Errorf("%w: %s %s takes %s, %s takes %s", ErrArgs, c.Locales[0].Name, key, describe(args[key]), l.Name, describe(kinds)))
			}
		}
	}

	return args, issues
}

func describe(kinds map[string]icu.Kind) string {
	parts := make([]string, 0, len(kinds))
	for _, name := range slices.Sorted(maps.Keys(kinds)) {
		parts = append(parts, name+" "+string(kinds[name]))
	}

	return "{" + strings.Join(parts, ", ") + "}"
}
