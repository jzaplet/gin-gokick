package unused

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"gokick/tools/i18n/frontend"
)

var ErrUnused = errors.New("no template, Go constant or script uses the key")

var ErrBuilt = errors.New("a script builds a key at runtime")

func Check(keys, used []string, literals []frontend.Literal) error {
	seen := map[string]bool{}
	for _, key := range used {
		seen[key] = true
	}

	var issues []error

	for _, literal := range literals {
		if literal.Built == false {
			seen[literal.Text] = true
		}

		if startsKey(literal.Text, keys) {
			issues = append(issues, fmt.Errorf("%s: %w: %q", literal.Pos, ErrBuilt, literal.Text))
		}
	}

	for _, key := range keys {
		if seen[key] == false {
			issues = append(issues, fmt.Errorf("%w: %s", ErrUnused, key))
		}
	}

	return errors.Join(issues...)
}

func startsKey(text string, keys []string) bool {
	return strings.HasSuffix(text, ".") && slices.ContainsFunc(keys, func(key string) bool {
		return strings.HasPrefix(key, text)
	})
}
