package typescript

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"gokick/app/core/i18n/icu"
	"gokick/tools/codegen"
)

var bareKey = regexp.MustCompile(`^(?:[A-Za-z_$][\w$]*|0|[1-9]\d*)$`)

func message(m icu.Message) string {
	if text, ok := plainText(m); ok {
		return codegen.Quote(text)
	}

	return list(m)
}

func list(m icu.Message) string {
	parts := make([]string, 0, len(m))
	for _, part := range m {
		parts = append(parts, partOf(part))
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

func partOf(part icu.Part) string {
	switch part := part.(type) {
	case icu.Text:
		return codegen.Quote(string(part))
	case icu.Arg:
		return fmt.Sprintf("{ type: 'arg', name: %s }", codegen.Quote(part.Name))
	case icu.Pound:
		return "{ type: 'pound' }"
	case icu.Plural:
		return fmt.Sprintf("{ type: 'plural', name: %s, offset: %d, exact: %s, cases: %s }", codegen.Quote(part.Name), part.Offset, exact(part.Exact), categories(part.Categories))
	case icu.Select:
		return fmt.Sprintf("{ type: 'select', name: %s, cases: %s }", codegen.Quote(part.Name), selectCases(part.Cases))
	default:
		panic(fmt.Sprintf("an ICU part of the type %T", part))
	}
}

func exact(cases map[int]icu.Message) string {
	entries := make([]string, 0, len(cases))
	for _, n := range slices.Sorted(maps.Keys(cases)) {
		entries = append(entries, strconv.Itoa(n)+": "+message(cases[n]))
	}

	return object(entries)
}

func categories(cases map[string]icu.Message) string {
	var entries []string

	for _, category := range icu.CategoryOrder() {
		if m, ok := cases[category]; ok {
			entries = append(entries, category+": "+message(m))
		}
	}

	return object(entries)
}

func selectCases(cases map[string]icu.Message) string {
	keys := slices.DeleteFunc(slices.Sorted(maps.Keys(cases)), func(key string) bool { return key == "other" })
	keys = append(keys, "other")
	quoted := slices.ContainsFunc(keys, func(key string) bool { return bareKey.MatchString(key) == false })

	entries := make([]string, 0, len(keys))
	for _, key := range keys {
		name := key
		if quoted {
			name = codegen.Quote(key)
		}

		entries = append(entries, name+": "+message(cases[key]))
	}

	return object(entries)
}

func object(entries []string) string {
	if len(entries) == 0 {
		return "{}"
	}

	return "{ " + strings.Join(entries, ", ") + " }"
}
