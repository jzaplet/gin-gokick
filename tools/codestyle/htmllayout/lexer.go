package htmllayout

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

var errUnclosed = errors.New("not closed")

type tokenKind int

const (
	textToken tokenKind = iota
	actionToken
	startTagToken
	endTagToken
	commentToken
	doctypeToken
	rawTextToken
)

type token struct {
	kind tokenKind

	raw string

	name string

	attrs []string

	selfClosing bool

	at int
}

var keywords = slices.Concat(opening, []string{"else", "end"})

var rawTextTags = []string{"script", "style", "textarea", "title", "pre"}

type lexer struct {
	src string

	at int

	tokens []token
}

func lex(src string) ([]token, error) {
	l := &lexer{src: src}
	for l.at < len(src) {
		if err := l.next(); err != nil {
			return nil, fmt.Errorf("line %d: %w", lineOf(src, l.at), err)
		}
	}

	return l.tokens, nil
}

func (l *lexer) next() error {
	rest := l.src[l.at:]
	switch {
	case strings.HasPrefix(rest, "{{"):
		end, err := actionEnd(l.src, l.at)
		if err != nil {
			return err
		}

		l.emit(&token{kind: actionToken, raw: l.src[l.at:end], name: keyword(l.src[l.at:end])}, end)
	case strings.HasPrefix(rest, "<!--"):
		end := strings.Index(rest, "-->")
		if end < 0 {
			return fmt.Errorf("comment %w", errUnclosed)
		}

		l.emit(&token{kind: commentToken, raw: rest[:end+3]}, l.at+end+3)
	case strings.HasPrefix(rest, "<!"):
		end := strings.IndexByte(rest, '>')
		if end < 0 {
			return fmt.Errorf("doctype %w", errUnclosed)
		}

		l.emit(&token{kind: doctypeToken, raw: rest[:end+1]}, l.at+end+1)
	case strings.HasPrefix(rest, "</") && len(rest) > 2 && isLetter(rest[2]):
		end := strings.IndexByte(rest, '>')
		if end < 0 {
			return fmt.Errorf("end tag %w", errUnclosed)
		}

		l.emit(&token{
			kind: endTagToken,

			raw: rest[:end+1],

			name: strings.ToLower(strings.TrimSpace(rest[2:end])),
		}, l.at+end+1)
	case rest[0] == '<' && len(rest) > 1 && isLetter(rest[1]):
		return l.tag()
	default:
		end := l.textEnd()
		l.emit(&token{kind: textToken, raw: rest[:end]}, l.at+end)
	}

	return nil
}

func (l *lexer) emit(t *token, end int) {
	t.at = l.at
	l.tokens = append(l.tokens, *t)
	l.at = end
}

func (l *lexer) textEnd() int {
	rest := l.src[l.at:]
	for i := 1; i < len(rest); i++ {
		if strings.HasPrefix(rest[i:], "{{") || (rest[i] == '<' && i+1 < len(rest) && startsMarkup(rest[i+1])) {
			return i
		}
	}

	return len(rest)
}

func (l *lexer) tag() error {
	j := l.at + 1
	for j < len(l.src) && isNameByte(l.src[j]) {
		j++
	}

	t := token{kind: startTagToken, name: strings.ToLower(l.src[l.at+1 : j])}
	for {
		for j < len(l.src) && isSpace(rune(l.src[j])) {
			j++
		}

		switch {
		case j >= len(l.src):
			return fmt.Errorf("tag <%s> %w", t.name, errUnclosed)
		case strings.HasPrefix(l.src[j:], "/>"):
			t.selfClosing = true
			t.raw = l.src[l.at : j+2]
			l.emit(&t, j+2)

			return nil
		case l.src[j] == '>':
			t.raw = l.src[l.at : j+1]
			l.emit(&t, j+1)

			return l.rawText(t.name)
		}

		end, err := attributeEnd(l.src, j)
		if err != nil {
			return err
		}

		t.attrs = append(t.attrs, l.src[j:end])
		j = end
	}
}

func (l *lexer) rawText(name string) error {
	if slices.Contains(rawTextTags, name) == false {
		return nil
	}

	j := l.at
	for j < len(l.src) {
		if strings.HasPrefix(l.src[j:], "{{") {
			end, err := actionEnd(l.src, j)
			if err != nil {
				return err
			}

			j = end

			continue
		}

		if strings.HasPrefix(l.src[j:], "</") && strings.EqualFold(prefix(l.src[j+2:], len(name)), name) {
			l.emit(&token{kind: rawTextToken, raw: l.src[l.at:j]}, j)

			return nil
		}

		j++
	}

	return fmt.Errorf("<%s> %w", name, errUnclosed)
}

func attributeEnd(src string, j int) (int, error) {
	for j < len(src) && isSpace(rune(src[j])) == false && src[j] != '>' && strings.HasPrefix(src[j:], "/>") == false {
		switch {
		case strings.HasPrefix(src[j:], "{{"):
			end, err := actionEnd(src, j)
			if err != nil {
				return 0, err
			}

			j = end
		case src[j] == '=' && j+1 < len(src) && (src[j+1] == '"' || src[j+1] == '\''):
			end, err := quotedEnd(src, j+1)
			if err != nil {
				return 0, err
			}

			j = end
		default:
			j++
		}
	}

	return j, nil
}

func quotedEnd(src string, j int) (int, error) {
	quote := src[j]
	for j++; j < len(src); {
		switch {
		case strings.HasPrefix(src[j:], "{{"):
			end, err := actionEnd(src, j)
			if err != nil {
				return 0, err
			}

			j = end
		case src[j] == quote:
			return j + 1, nil
		default:
			j++
		}
	}

	return 0, fmt.Errorf("attribute value %w", errUnclosed)
}

func prefix(s string, n int) string {
	return s[:min(n, len(s))]
}

func lineOf(src string, at int) int {
	return strings.Count(src[:at], "\n") + 1
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isNameByte(b byte) bool {
	return isLetter(b) || (b >= '0' && b <= '9') || b == '-' || b == ':'
}

func startsMarkup(b byte) bool {
	return isLetter(b) || b == '/' || b == '!'
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f'
}
