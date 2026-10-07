package baretext

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"text/template/parse"
	"unicode"

	"golang.org/x/net/html"
)

var ErrBareText = errors.New("a template writes text outside t")

const action = "\uE000"

const digits = "0123456789"

const punctuation = "(),.&+-=*/#%!?:[]{}<>|\u00b7\u2022\u2010\u2013\u2014\u2212"

var textAttributes = []string{"title", "alt", "label", "placeholder", "aria-label", "aria-placeholder", "aria-roledescription", "aria-valuetext"}

var rawTextTags = []string{"script", "style"}

type segment struct {
	at int

	pos parse.Pos

	source bool
}

type page struct {
	html strings.Builder

	segments []segment
}

type found struct {
	at int

	in string

	text string
}

func Check(tree *parse.Tree) []error {
	var p page
	p.walk(tree.Root)
	var issues []error

	for _, bare := range p.bare() {
		location, _ := tree.ErrorContext(&parse.TextNode{NodeType: parse.NodeText, Pos: p.pos(bare.at)})
		issues = append(issues, fmt.Errorf("%s: %w: %s", location, ErrBareText, bare))
	}

	return issues
}

func (p *page) walk(node parse.Node) {
	switch n := node.(type) {
	case *parse.ListNode:
		if n == nil {
			return
		}

		for _, child := range n.Nodes {
			p.walk(child)
		}
	case *parse.TextNode:
		p.add(n.Pos, string(n.Text), true)
	case *parse.IfNode:
		p.branch(&n.BranchNode)
	case *parse.RangeNode:
		p.branch(&n.BranchNode)
	case *parse.WithNode:
		p.branch(&n.BranchNode)
	default:
		p.add(n.Position(), action, false)
	}
}

func (p *page) branch(branch *parse.BranchNode) {
	p.add(branch.Pos, action, false)
	p.walk(branch.List)
	p.add(branch.Pos, action, false)
	p.walk(branch.ElseList)
	p.add(branch.Pos, action, false)
}

func (p *page) add(pos parse.Pos, text string, source bool) {
	p.segments = append(p.segments, segment{at: p.html.Len(), pos: pos, source: source})
	p.html.WriteString(text)
}

func (p *page) pos(at int) parse.Pos {
	var last segment

	for _, s := range p.segments {
		if s.at > at {
			break
		}

		last = s
	}

	if last.source == false {
		return last.pos
	}

	return last.pos + parse.Pos(at-last.at)
}

func (p *page) bare() []found {
	z := html.NewTokenizer(strings.NewReader(p.html.String()))
	var all []found
	at, rawText := 0, false

	for {
		kind := z.Next()
		raw := string(z.Raw())
		start := at
		at += len(raw)

		switch kind {
		case html.ErrorToken:
			return all
		case html.TextToken:
			if text := string(z.Text()); rawText == false && isBare(text) {
				all = append(all, found{
					at: start + len(raw) - len(strings.TrimLeftFunc(raw, unicode.IsSpace)),

					text: text,
				})
			}

			rawText = false
		case html.StartTagToken, html.SelfClosingTagToken:
			name, more := z.TagName()
			tag := string(name)
			all = append(all, attributes(z, start, tag, more)...)
			rawText = kind == html.StartTagToken && slices.Contains(rawTextTags, tag)
		case html.EndTagToken, html.CommentToken, html.DoctypeToken:
			rawText = false
		}
	}
}

func attributes(z *html.Tokenizer, at int, tag string, more bool) []found {
	values := map[string]string{}

	for more {
		var key, value []byte
		key, value, more = z.TagAttr()
		values[string(key)] = string(value)
	}

	var all []found

	for _, name := range textAttributes {
		if text, ok := values[name]; ok && isBare(text) {
			all = append(all, found{at: at, in: name, text: text})
		}
	}

	if content := values["content"]; tag == "meta" && values["name"] == "description" && isBare(content) {
		all = append(all, found{at: at, in: "content", text: content})
	}

	return all
}

func isBare(text string) bool {
	return strings.ContainsFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) == false && unicode.Is(unicode.Cf, r) == false && strings.ContainsRune(action+digits+punctuation, r) == false
	})
}

func (f found) String() string {
	text := strconv.Quote(strings.ReplaceAll(strings.TrimSpace(f.text), action, "{{…}}"))
	if f.in == "" {
		return text
	}

	return f.in + "=" + text
}
