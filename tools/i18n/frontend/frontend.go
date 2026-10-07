package frontend

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

var ErrUnreadable = errors.New("the check cannot read keys in this kind of script")

type Literal struct {
	Pos string

	Text string

	Built bool
}

var scripts = []string{".ts", ".mts", ".cts", ".js", ".mjs", ".cjs"}

var unreadable = []string{".tsx", ".jsx"}

var boundPrefixes = []string{":", "@", "#", ".", "v-"}

func Find(fsys fs.FS, root string) ([]Literal, error) {
	var literals []Literal
	err := fs.WalkDir(fsys, root, func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}

		ext := path.Ext(name)
		if slices.Contains(unreadable, ext) {
			return fmt.Errorf("%s: %w", name, ErrUnreadable)
		}

		if ext != ".vue" && slices.Contains(scripts, ext) == false {
			return nil
		}

		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}

		src := string(content)
		if strings.HasPrefix(src, "// Code generated") {
			return nil
		}

		pieces := lex(src)
		if ext == ".vue" {
			pieces = component(src)
		}

		slices.SortFunc(pieces, func(a, b piece) int { return cmp.Compare(a.at, b.at) })

		for _, f := range pieces {
			literals = append(literals, Literal{Pos: position(name, src, f.at), Text: f.text, Built: f.built})
		}

		return nil
	})

	return literals, err
}

func component(src string) []piece {
	masked, all := mustaches(src)
	z := html.NewTokenizer(strings.NewReader(masked))
	at, script := 0, false

	for {
		kind := z.Next()
		raw := string(z.Raw())
		start := at
		at += len(raw)

		switch kind {
		case html.ErrorToken:
			return all
		case html.TextToken:
			if script {
				all = append(all, shifted(lex(raw), start)...)
			}

			script = false
		case html.StartTagToken, html.SelfClosingTagToken:
			name, more := z.TagName()
			all = append(all, attributes(z, raw, start, more)...)
			script = kind == html.StartTagToken && string(name) == "script"
		case html.EndTagToken, html.CommentToken, html.DoctypeToken:
			script = false
		}
	}
}

func attributes(z *html.Tokenizer, raw string, at int, more bool) []piece {
	var all []piece

	for more {
		var key, value []byte
		key, value, more = z.TagAttr()

		offset := at + max(strings.Index(raw, string(value)), 0)
		switch {
		case bound(string(key)):
			all = append(all, shifted(lex(string(value)), offset)...)
		case len(value) > 0:
			all = append(all, piece{at: offset, text: string(value)})
		}
	}

	return all
}

func bound(attribute string) bool {
	return slices.ContainsFunc(boundPrefixes, func(prefix string) bool {
		return strings.HasPrefix(attribute, prefix)
	})
}

func mustaches(src string) (string, []piece) {
	masked := []byte(src)
	var all []piece

	at, end := strings.Index(src, "<template"), strings.LastIndex(src, "</template>")
	for at >= 0 && at < end {
		open := strings.Index(src[at:end], "{{")
		if open < 0 {
			break
		}

		start := at + open + 2

		length := strings.Index(src[start:end], "}}")
		if length < 0 {
			break
		}

		all = append(all, shifted(lex(src[start:start+length]), start)...)

		at = start + length + 2
		for i := start - 2; i < at; i++ {
			if masked[i] != '\n' {
				masked[i] = ' '
			}
		}
	}

	return string(masked), all
}

func shifted(pieces []piece, by int) []piece {
	for i := range pieces {
		pieces[i].at += by
	}

	return pieces
}

func position(name, src string, at int) string {
	before := src[:at]
	line := 1 + strings.Count(before, "\n")
	column := 1 + len([]rune(before[strings.LastIndex(before, "\n")+1:]))

	return fmt.Sprintf("%s:%d:%d", name, line, column)
}
