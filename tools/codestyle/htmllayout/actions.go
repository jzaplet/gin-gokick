package htmllayout

import (
	"fmt"
	"slices"
	"strings"
)

func actionEnd(src string, at int) (int, error) {
	j := at + 2
	for j < len(src) {
		var err error

		switch {
		case strings.HasPrefix(src[j:], "}}"):
			return j + 2, nil
		case strings.HasPrefix(src[j:], "/*"):
			end := strings.Index(src[j+2:], "*/")
			if end < 0 {
				return 0, fmt.Errorf("template comment %w", errUnclosed)
			}

			j += 2 + end + 2
		case src[j] == '"' || src[j] == '\'':
			j, err = literalEnd(src, j)
		case src[j] == '`':
			end := strings.IndexByte(src[j+1:], '`')
			if end < 0 {
				return 0, fmt.Errorf("raw string %w", errUnclosed)
			}

			j += 1 + end + 1
		default:
			j++
		}

		if err != nil {
			return 0, err
		}
	}

	return 0, fmt.Errorf("action %w", errUnclosed)
}

func literalEnd(src string, j int) (int, error) {
	quote := src[j]
	for j++; j < len(src) && src[j] != '\n'; j++ {
		switch src[j] {
		case '\\':
			j++
		case quote:
			return j + 1, nil
		}
	}

	return 0, fmt.Errorf("string %w", errUnclosed)
}

func keyword(action string) string {
	inner := strings.TrimLeft(strings.TrimPrefix(action[2:], "-"), " \t\r\n")

	end := 0
	for end < len(inner) && isLetter(inner[end]) {
		end++
	}

	if word := inner[:end]; slices.Contains(keywords, word) {
		return word
	}

	return ""
}

func trimsLeft(action string) bool {
	return len(action) > 3 && action[2] == '-' && isSpace(rune(action[3]))
}

func trimsRight(action string) bool {
	return len(action) > 4 && action[len(action)-3] == '-' && isSpace(rune(action[len(action)-4]))
}
