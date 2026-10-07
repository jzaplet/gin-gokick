package htmllayout

import (
	"slices"
	"strings"
)

var blockTags = []string{
	"address", "article", "aside", "base", "blockquote", "body", "caption", "colgroup", "dd", "details", "dialog", "div",
	"dl", "dt", "fieldset", "figcaption", "figure", "footer", "form", "h1", "h2", "h3", "h4", "h5", "h6", "head",
	"header", "hr", "html", "legend", "li", "link", "main", "menu", "meta", "nav", "ol", "p", "pre", "section",
	"summary", "table", "tbody", "td", "tfoot", "th", "thead", "title", "tr", "ul",
}

var insensitiveTags = []string{
	"colgroup", "datalist", "dl", "head", "html", "menu", "ol", "optgroup", "select", "table", "tbody", "tfoot", "thead",
	"tr", "ul",
}

type area struct {
	owner *node

	branch int
}

type context struct {
	area area

	insensitive bool

	open *node

	close *node
}

type laidOut struct {
	lines block

	firstRow int

	startBroken bool

	endBroken bool
}

type printer struct {
	split map[*node]bool

	expanded map[area]bool

	level int
}

func (c context) root() bool {
	return c.area.owner == nil
}

func (c context) definition() bool {
	owner := c.area.owner

	return owner != nil && owner.kind == constructNode && (owner.token.name == "define" || owner.token.name == "block")
}

func (p *printer) node(n *node, insensitive bool) block {
	p.level++
	defer func() { p.level-- }()

	switch n.kind {
	case elementNode:
		return p.element(n)
	case constructNode:
		return p.construct(n, insensitive)
	case leafNode, rootNode:
	}

	return text(n.token.raw)
}

func (p *printer) element(n *node) block {
	start := p.startTag(n)
	switch {
	case n.end == nil:
		return start
	case slices.Contains(rawTextTags, n.token.name):
		for _, raw := range n.children {
			start = glue(start, text(raw.token.raw))
		}

		return glue(start, text(n.end.raw))
	}

	content := p.container(n.children, context{
		area: area{owner: n},

		insensitive: slices.Contains(insensitiveTags, n.token.name),

		open: n,

		close: n,
	})

	return wrap(start, content, text(n.end.raw))
}

func (p *printer) construct(n *node, insensitive bool) block {
	actions := slices.Concat([]token{n.token}, n.separators, []token{*n.end})
	var out block

	broken := false
	for i, branch := range n.branches {
		out = join(out, text(actions[i].raw), broken)
		part := p.container(branch, context{
			area: area{owner: n, branch: i},

			insensitive: insensitive,

			open: &node{kind: leafNode, token: actions[i]},

			close: &node{kind: leafNode, token: actions[i+1]},
		})
		out = content(out, part)
		broken = part.endBroken
	}

	return join(out, text(n.end.raw), broken)
}

func (p *printer) startTag(n *node) block {
	t := n.token
	name := "<" + t.raw[1:1+len(t.name)]

	closing := ">"
	if t.selfClosing {
		closing = "/>"
	}

	if len(t.attrs) == 0 {
		return text(name + closing)
	}

	if p.split[n] == false {
		tag := text(name + " " + strings.Join(t.attrs, " ") + closing)
		tag[0].tags = []*node{n}

		return tag
	}

	tag := text(name)
	for _, attr := range t.attrs {
		tag = append(tag, indent(text(attr), 1)...)
	}

	return append(tag, line{text: closing})
}

func (p *printer) container(children []*node, ctx context) laidOut {
	items, spaces := spaced(children)
	if len(items) == 0 {
		if spaces[0] == "" || ctx.root() {
			return laidOut{}
		}

		return laidOut{lines: block{{text: " "}}, firstRow: 1}
	}

	g := &gaps{items: items, spaces: spaces, ctx: ctx, expanded: p.expanded[ctx.area]}
	blocks := make([]block, len(items))

	multi := g.broken(len(items))
	for i, item := range items {
		blocks[i] = p.node(item, ctx.insensitive)
		multi = multi || len(blocks[i]) > 1 || g.broken(i)
	}

	out := laidOut{startBroken: multi && g.allowed(0), endBroken: multi && g.allowed(len(items))}
	out.lines, out.firstRow = g.assemble(blocks, out.startBroken, out.endBroken, p.level)

	return out
}

func (g *gaps) assemble(blocks []block, startBroken, endBroken bool, level int) (lines block, firstRow int) {
	offer := candidate{area: g.ctx.area, level: level}

	lines = blocks[0]
	if startBroken == false {
		lines = glue(text(g.edgeSpace(0)), lines)
		lines.offer(0, offer, g.offers(0))
	}

	firstRow = -1

	for i := 1; i < len(blocks); i++ {
		if g.broken(i) == false {
			lines.offer(len(lines)-1, offer, g.offers(i))
			lines = glue(lines, glue(text(collapse(g.spaces[i])), blocks[i]))

			continue
		}

		if firstRow < 0 {
			firstRow = len(lines)
		}

		if strings.Count(g.spaces[i], "\n") > 1 {
			lines = append(lines, line{verbatim: true})
		}

		lines = append(lines, blocks[i]...)
	}

	if firstRow < 0 {
		firstRow = len(lines)
	}

	if endBroken == false {
		lines.offer(len(lines)-1, offer, g.offers(len(blocks)))
		lines = glue(lines, text(g.edgeSpace(len(blocks))))
	}

	return lines, firstRow
}

func spaced(children []*node) (items []*node, spaces []string) {
	spaces = []string{""}

	for _, child := range children {
		if child.kind != leafNode || child.token.kind != textToken {
			items = append(items, child)
			spaces = append(spaces, "")

			continue
		}

		raw := child.token.raw

		core := strings.TrimFunc(raw, isSpace)
		if core == "" {
			spaces[len(spaces)-1] += raw

			continue
		}

		start := strings.Index(raw, core)
		spaces[len(spaces)-1] += raw[:start]
		items = append(items, &node{kind: leafNode, token: token{kind: textToken, raw: core}})
		spaces = append(spaces, raw[start+len(core):])
	}

	return items, spaces
}

func collapse(space string) string {
	if strings.ContainsAny(space, "\n\r") {
		return " "
	}

	return space
}

func wrap(start block, inner laidOut, end block) block {
	return join(content(start, inner), end, inner.endBroken)
}

func content(start block, inner laidOut) block {
	lines := indent(inner.lines, 1)
	if inner.startBroken {
		return slices.Concat(start, lines)
	}

	return slices.Concat(glue(start, lines[:inner.firstRow]), lines[inner.firstRow:])
}

func join(a, b block, broken bool) block {
	if broken {
		return slices.Concat(a, b)
	}

	return glue(a, b)
}
