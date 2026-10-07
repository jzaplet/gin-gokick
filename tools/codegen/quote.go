package codegen

import (
	"fmt"
	"strings"
	"unicode"
)

func Quote(s string) string {
	var b strings.Builder
	b.WriteByte('\'')

	for _, r := range s {
		switch {
		case r == '\'' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case unicode.IsPrint(r):
			b.WriteRune(r)
		case r > 0xffff:
			fmt.Fprintf(&b, `\u{%x}`, r)
		default:
			fmt.Fprintf(&b, `\u%04x`, r)
		}
	}

	b.WriteByte('\'')

	return b.String()
}
