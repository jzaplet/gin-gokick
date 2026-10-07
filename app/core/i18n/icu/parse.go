package icu

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

const other = "other"

var ErrSyntax = errors.New("icu: syntax error")

var ErrUnsupported = errors.New("icu: unsupported argument type")

type parser struct {
	source []rune

	pos int
}

func Parse(source string) (Message, error) {
	p := &parser{source: []rune(source)}

	m, err := p.message(false, false)
	if err != nil {
		return nil, err
	}

	if _, err := m.Args(); err != nil {
		return nil, err
	}

	return m, nil
}

func (p *parser) message(nested, plural bool) (Message, error) {
	var parts Message
	var text strings.Builder

	for p.done() == false && (nested == false || p.peek() != '}') {
		var part Part
		var err error

		switch r := p.peek(); {
		case r == '}':
			err = p.fail("a } closes nothing")
		case r == '{':
			part, err = p.argument()
		case r == '#' && plural:
			p.pos++
			part = Pound{}
		case r == '\'':
			err = p.apostrophe(&text, plural)
		default:
			p.pos++

			text.WriteRune(r)
		}

		if err != nil {
			return nil, err
		}

		if part != nil {
			parts = appendText(parts, &text)
			parts = append(parts, part)
		}
	}

	return appendText(parts, &text), nil
}

func appendText(parts Message, text *strings.Builder) Message {
	if text.Len() == 0 {
		return parts
	}

	defer text.Reset()

	return append(parts, Text(text.String()))
}

func (p *parser) apostrophe(text *strings.Builder, plural bool) error {
	start := p.pos

	p.pos++
	if p.done() {
		text.WriteByte('\'')

		return nil
	}

	switch next := p.peek(); {
	case next == '\'':
		p.pos++

		text.WriteByte('\'')
	case next == '<' || next == '>':
		return p.failAt(start, "an apostrophe before < or > starts a quote in FormatJS but not in ICU4J, write '' for the apostrophe")
	case next == '{' || next == '}' || (next == '#' && plural):
		return p.quote(start, text)
	default:
		text.WriteByte('\'')
	}

	return nil
}

func (p *parser) quote(start int, text *strings.Builder) error {
	for p.done() == false {
		r := p.source[p.pos]
		p.pos++

		if r != '\'' {
			text.WriteRune(r)

			continue
		}

		if p.done() || p.peek() != '\'' {
			return nil
		}

		p.pos++

		text.WriteByte('\'')
	}

	return p.failAt(start, "the quote never ends")
}

func (p *parser) argument() (Part, error) {
	start := p.pos
	p.pos++
	p.space()

	name := p.take(isNameRune)
	if name == "" || isLetter(rune(name[0])) == false {
		return nil, p.fail("an argument name starts with a letter")
	}

	p.space()

	if p.eat('}') {
		return Arg{Name: name}, nil
	}

	if p.eat(',') == false {
		return nil, p.fail("expected , or } after the argument name")
	}

	p.space()

	switch kind := p.take(isLetter); kind {
	case "plural":
		return p.plural(start, name)
	case "select":
		return p.choice(start, name)
	case "":
		return nil, p.fail("expected the type of the argument")
	default:
		return nil, fmt.Errorf("%w at %d: %s", ErrUnsupported, start, kind)
	}
}

func (p *parser) plural(start int, name string) (Part, error) {
	part := Plural{Name: name, Exact: map[int]Message{}, Categories: map[string]Message{}}

	if err := p.expect(','); err != nil {
		return nil, err
	}

	p.space()

	if p.prefix("offset:") {
		p.space()

		offset, err := p.integer()
		if err != nil {
			return nil, err
		}

		part.Offset = offset
	}

	for p.space(); p.eat('}') == false; p.space() {
		var err error
		if p.eat('=') {
			err = p.exact(part.Exact)
		} else {
			err = p.category(part.Categories)
		}

		if err != nil {
			return nil, err
		}
	}

	if _, ok := part.Categories[other]; ok == false {
		return nil, p.failAt(start, "a plural needs the case other")
	}

	return part, nil
}

func (p *parser) exact(cases map[int]Message) error {
	value, err := p.integer()
	if err != nil {
		return err
	}

	return addCase(p, cases, value, true)
}

func (p *parser) category(cases map[string]Message) error {
	key := p.take(isNameRune)
	if key == "" {
		return p.fail("expected a plural case or }")
	}

	if slices.Contains(categories[:], key) == false {
		return p.fail("a plural case is =N or a CLDR category, not " + key)
	}

	return addCase(p, cases, key, true)
}

func (p *parser) choice(start int, name string) (Part, error) {
	part := Select{Name: name, Cases: map[string]Message{}}

	if err := p.expect(','); err != nil {
		return nil, err
	}

	for p.space(); p.eat('}') == false; p.space() {
		key := p.take(isNameRune)
		if key == "" {
			return nil, p.fail("expected a select case or }")
		}

		if key == "__proto__" {
			return nil, p.fail("FormatJS can't hold the select case __proto__")
		}

		if err := addCase(p, part.Cases, key, false); err != nil {
			return nil, err
		}
	}

	if _, ok := part.Cases[other]; ok == false {
		return nil, p.failAt(start, "a select needs the case other")
	}

	return part, nil
}

func addCase[K comparable](p *parser, cases map[K]Message, key K, plural bool) error {
	if _, ok := cases[key]; ok {
		return p.fail(fmt.Sprintf("the case %v repeats", key))
	}

	if err := p.expect('{'); err != nil {
		return err
	}

	m, err := p.message(true, plural)
	if err != nil {
		return err
	}

	if p.eat('}') == false {
		return p.fail("a { never closes")
	}

	cases[key] = m

	return nil
}

func (p *parser) integer() (int, error) {
	digits := p.take(isDigit)
	if digits == "" {
		return 0, p.fail("expected a whole number without a sign")
	}

	if len(digits) > 1 && digits[0] == '0' {
		return 0, p.fail("FormatJS never matches a number with a leading zero, write " + strings.TrimLeft(digits, "0"))
	}

	value, err := strconv.Atoi(digits)
	if err != nil || value > maxSafe {
		return 0, p.fail("the number " + digits + " is beyond 2^53")
	}

	return value, nil
}

func (p *parser) expect(r rune) error {
	p.space()

	if p.eat(r) == false {
		return p.fail(fmt.Sprintf("expected %c", r))
	}

	return nil
}

func (p *parser) prefix(word string) bool {
	if strings.HasPrefix(string(p.source[p.pos:]), word) == false {
		return false
	}

	p.pos += len([]rune(word))

	return true
}

func (p *parser) take(accept func(rune) bool) string {
	start := p.pos
	for p.done() == false && accept(p.peek()) {
		p.pos++
	}

	return string(p.source[start:p.pos])
}

func (p *parser) space() {
	p.take(isSpace)
}

func (p *parser) eat(r rune) bool {
	if p.done() || p.peek() != r {
		return false
	}

	p.pos++

	return true
}

func (p *parser) peek() rune {
	return p.source[p.pos]
}

func (p *parser) done() bool {
	return p.pos >= len(p.source)
}

func (p *parser) fail(reason string) error {
	return p.failAt(p.pos, reason)
}

func (p *parser) failAt(pos int, reason string) error {
	return fmt.Errorf("%w at %d: %s", ErrSyntax, pos, reason)
}

func isLetter(r rune) bool {
	return ('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z')
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isNameRune(r rune) bool {
	return isLetter(r) || isDigit(r) || r == '_'
}

func isSpace(r rune) bool {
	return unicode.Is(unicode.Pattern_White_Space, r)
}
