package htmllayout

import (
	"errors"
	"fmt"
	"slices"
)

var errNesting = errors.New("elements and template blocks do not nest")

type nodeKind int

const (
	rootNode nodeKind = iota
	leafNode
	elementNode
	constructNode
)

type node struct {
	kind nodeKind

	token token

	children []*node

	branches [][]*node

	separators []token

	end *token
}

var voidTags = []string{"area", "base", "br", "col", "embed", "hr", "img", "input", "link", "meta", "source", "track", "wbr"}

var opening = []string{"if", "range", "with", "define", "block"}

type builder struct {
	src string

	stack []*node
}

func build(src string, tokens []token) (*node, error) {
	b := &builder{src: src, stack: []*node{{kind: rootNode}}}
	for i := range tokens {
		if err := b.take(&tokens[i]); err != nil {
			return nil, err
		}
	}

	if len(b.stack) > 1 {
		return nil, nestingError(src, &b.stack[len(b.stack)-1].token)
	}

	return b.stack[0], nil
}

func (b *builder) take(t *token) error {
	top := b.stack[len(b.stack)-1]

	switch {
	case t.kind == startTagToken:
		element := &node{kind: elementNode, token: *t}
		top.add(element)

		if t.selfClosing == false && slices.Contains(voidTags, t.name) == false {
			b.stack = append(b.stack, element)
		}
	case t.kind == endTagToken:
		if top.kind != elementNode || top.token.name != t.name {
			return nestingError(b.src, t)
		}

		top.end = t
		b.stack = b.stack[:len(b.stack)-1]
	case t.kind == actionToken && slices.Contains(opening, t.name):
		construct := &node{kind: constructNode, token: *t, branches: [][]*node{nil}}
		top.add(construct)
		b.stack = append(b.stack, construct)
	case t.kind == actionToken && (t.name == "else" || t.name == "end"):
		return b.close(top, t)
	default:
		top.add(&node{kind: leafNode, token: *t})
	}

	return nil
}

func (b *builder) close(top *node, t *token) error {
	if top.kind != constructNode {
		return nestingError(b.src, t)
	}

	if t.name == "else" {
		top.separators = append(top.separators, *t)
		top.branches = append(top.branches, nil)

		return nil
	}

	top.end = t
	b.stack = b.stack[:len(b.stack)-1]

	return nil
}

func (n *node) add(child *node) {
	if n.kind == constructNode {
		n.branches[len(n.branches)-1] = append(n.branches[len(n.branches)-1], child)

		return
	}

	n.children = append(n.children, child)
}

func nestingError(src string, t *token) error {
	return fmt.Errorf("line %d: %w at %s", lineOf(src, t.at), errNesting, t.raw)
}
