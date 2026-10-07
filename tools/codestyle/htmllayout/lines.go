package htmllayout

import (
	"slices"
	"strings"

	"gokick/tools/codestyle/gofiles"
)

const indentWidth = 4

type line struct {
	depth int

	text string

	verbatim bool

	tags []*node

	candidates []candidate
}

type candidate struct {
	area area

	level int
}

type block []line

type gaps struct {
	items []*node

	spaces []string

	ctx context

	expanded bool
}

func text(s string) block {
	parts := strings.Split(s, "\n")

	lines := block{{text: parts[0]}}
	for _, part := range parts[1:] {
		lines = append(lines, line{text: part, verbatim: true})
	}

	return lines
}

func glue(a, b block) block {
	if len(b) == 0 {
		return a
	}

	if len(a) == 0 {
		return b
	}

	out := slices.Clone(a)
	last := out[len(out)-1]
	last.text += b[0].text
	last.tags = slices.Concat(last.tags, b[0].tags)
	last.candidates = slices.Concat(last.candidates, b[0].candidates)

	out[len(out)-1] = last
	for _, next := range b[1:] {
		next.depth += last.depth - b[0].depth
		out = append(out, next)
	}

	return out
}

func indent(b block, levels int) block {
	out := slices.Clone(b)
	for i := range out {
		out[i].depth += levels
	}

	return out
}

func (b block) String() string {
	rendered := make([]string, len(b))
	for i, l := range b {
		rendered[i] = l.text
		if l.verbatim == false {
			rendered[i] = strings.Repeat(" ", l.depth*indentWidth) + l.text
		}
	}

	return strings.Join(rendered, "\n")
}

func (b block) offer(at int, c candidate, allowed bool) {
	if allowed {
		b[at].candidates = append(b[at].candidates, c)
	}
}

func (b block) overflowing() (*area, *node) {
	for i := range b {
		l := &b[i]
		if l.width() <= gofiles.MaxWidth {
			continue
		}

		var outermost *candidate
		for j := range l.candidates {
			if outermost == nil || l.candidates[j].level < outermost.level {
				outermost = &l.candidates[j]
			}
		}

		if outermost != nil {
			return &outermost.area, nil
		}

		if len(l.tags) > 0 {
			return nil, l.tags[0]
		}
	}

	return nil, nil
}

func (l *line) width() int {
	if l.verbatim {
		return gofiles.Width(l.text)
	}

	return l.depth*indentWidth + gofiles.Width(l.text)
}

func (g *gaps) allowed(i int) bool {
	if g.ctx.root() && (i == 0 || i == len(g.items)) {
		return false
	}

	left, right := g.bounded(i)

	return g.spaces[i] != "" || g.ctx.insensitive || isBlock(left) || isBlock(right) || trimsAfter(left) || trimsBefore(right)
}

func (g *gaps) offers(i int) bool {
	return g.ctx.definition() == false && g.allowed(i)
}

func (g *gaps) broken(i int) bool {
	left, right := g.neighbours(i)
	forced := (g.ctx.root() && strings.Contains(g.spaces[i], "\n")) || g.ctx.insensitive || isBlock(left) || isBlock(right)

	return g.allowed(i) && (forced || g.expanded)
}

func (g *gaps) edgeSpace(i int) string {
	left, right := g.bounded(i)

	boundary := right
	if i == 0 {
		boundary = left
	}

	if boundary == nil || isBlock(boundary) {
		return ""
	}

	return collapse(g.spaces[i])
}

func (g *gaps) bounded(i int) (left, right *node) {
	left, right = g.neighbours(i)
	if i == 0 {
		left = g.ctx.open
	}

	if i == len(g.items) {
		right = g.ctx.close
	}

	return left, right
}

func (g *gaps) neighbours(i int) (left, right *node) {
	if i > 0 {
		left = g.items[i-1]
	}

	if i < len(g.items) {
		right = g.items[i]
	}

	return left, right
}

func isBlock(n *node) bool {
	switch {
	case n == nil:
		return false
	case n.kind == elementNode:
		return slices.Contains(blockTags, n.token.name)
	default:
		return n.kind == leafNode && n.token.kind == doctypeToken
	}
}

func trimsAfter(n *node) bool {
	switch {
	case n == nil:
		return false
	case n.kind == constructNode:
		return trimsRight(n.end.raw)
	default:
		return n.kind == leafNode && n.token.kind == actionToken && trimsRight(n.token.raw)
	}
}

func trimsBefore(n *node) bool {
	switch {
	case n == nil:
		return false
	case n.kind == constructNode:
		return trimsLeft(n.token.raw)
	default:
		return n.kind == leafNode && n.token.kind == actionToken && trimsLeft(n.token.raw)
	}
}
