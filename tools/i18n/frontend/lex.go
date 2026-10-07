package frontend

import "strings"

type piece struct {
	at int

	text string

	built bool
}

type lexer struct {
	src string

	pos int

	regex bool

	pieces []piece
}

var braces = map[byte]int{'{': 1, '}': -1}

var regexAfter = map[string]bool{
	"return": true,

	"typeof": true,

	"case": true,

	"do": true,

	"else": true,

	"in": true,

	"of": true,

	"new": true,

	"delete": true,

	"void": true,

	"throw": true,

	"instanceof": true,

	"yield": true,

	"await": true,
}

func lex(src string) []piece {
	l := lexer{src: src, regex: true}
	l.code(false)

	return l.pieces
}

func (l *lexer) code(nested bool) {
	depth := 0

	for l.pos < len(l.src) {
		c := l.src[l.pos]
		switch {
		case c == '\'' || c == '"':
			l.quoted(c)
		case c == '`':
			l.template()
		case strings.HasPrefix(l.src[l.pos:], "//"):
			l.skipPast("\n")
		case strings.HasPrefix(l.src[l.pos:], "/*"):
			l.skipPast("*/")
		case c == '/' && l.regex:
			l.regexLiteral()
		case c == '}' && nested && depth == 0:
			l.pos++

			return
		case isWord(c):
			l.word()
		default:
			depth += braces[c]
			l.punctuation(c)
		}
	}
}

func (l *lexer) punctuation(c byte) {
	l.pos++

	switch c {
	case ' ', '\t', '\n', '\r':
	case ')', ']':
		l.regex = false
	default:
		l.regex = true
	}
}

func (l *lexer) word() {
	start := l.pos
	for l.pos < len(l.src) && isWord(l.src[l.pos]) {
		l.pos++
	}

	l.regex = regexAfter[l.src[start:l.pos]]
}

func isWord(c byte) bool {
	return c == '_' || c == '$' || c >= 0x80 || ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9')
}

func (l *lexer) skipPast(end string) {
	if i := strings.Index(l.src[l.pos:], end); i >= 0 {
		l.pos += i + len(end)

		return
	}

	l.pos = len(l.src)
}

func (l *lexer) quoted(quote byte) {
	start := l.pos

	l.pos++
	for l.pos < len(l.src) && l.src[l.pos] != quote && l.src[l.pos] != '\n' {
		if l.src[l.pos] == '\\' {
			l.pos++
		}

		l.pos++
	}

	l.pieces = append(l.pieces, piece{at: start, text: l.src[start+1 : min(l.pos, len(l.src))]})
	l.pos++
	l.regex = false
}

func (l *lexer) template() {
	start := l.pos
	l.pos++
	head, built := "", false

	for l.pos < len(l.src) && l.src[l.pos] != '`' {
		switch {
		case l.src[l.pos] == '\\':
			l.pos += 2
		case strings.HasPrefix(l.src[l.pos:], "${"):
			if built == false {
				head, built = l.src[start+1:l.pos], true
			}

			l.pos += 2
			l.regex = true
			l.code(true)
		default:
			l.pos++
		}
	}

	if built == false {
		head = l.src[start+1 : min(l.pos, len(l.src))]
	}

	l.pieces = append(l.pieces, piece{at: start, text: head, built: built})
	l.pos++
	l.regex = false
}

func (l *lexer) regexLiteral() {
	start := l.pos
	l.pos++
	class := false

	for l.pos < len(l.src) && l.src[l.pos] != '\n' {
		switch c := l.src[l.pos]; {
		case c == '\\':
			l.pos++
		case c == '[':
			class = true
		case c == ']':
			class = false
		case c == '/' && class == false:
			l.pos++
			l.word()
			l.regex = false

			return
		}

		l.pos++
	}

	l.pos = start + 1
	l.regex = true
}
