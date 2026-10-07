package htmllayout

import (
	"errors"
	"fmt"
	gotoken "go/token"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"text/template/parse"

	"gokick/tools/codestyle/gofiles"
)

const dir = "views"

var errChanged = errors.New("the layout would change what the template writes")

func Issues(root string) ([]gofiles.Issue, error) {
	var issues []gofiles.Issue
	err := walk(root, func(file, src, formatted string) error {
		if formatted != src {
			issues = append(issues, gofiles.Issue{
				Pos: gotoken.Position{Filename: file, Line: firstDifference(src, formatted)},

				Msg: "template not laid out",
			})
		}

		return nil
	})

	return issues, err
}

func Rewrite(root string) ([]string, error) {
	var changed []string
	err := walk(root, func(file, src, formatted string) error {
		if formatted == src {
			return nil
		}

		changed = append(changed, file)

		return os.WriteFile(file, []byte(formatted), 0o600)
	})

	return changed, err
}

func Format(src string) (string, error) {
	tokens, err := lex(src)
	if err != nil {
		return "", err
	}

	root, err := build(src, tokens)
	if err != nil {
		return "", err
	}

	formatted := src[:len(src)-len(strings.TrimLeftFunc(src, isSpace))] + narrow(root).String()
	if strings.TrimRightFunc(src, isSpace) != src {
		formatted += "\n"
	}

	if formatted == src {
		return src, nil
	}

	if err := sameContent(src, formatted); err != nil {
		return "", err
	}

	return formatted, nil
}

func narrow(root *node) block {
	p := &printer{split: map[*node]bool{}, expanded: map[area]bool{}}
	lay := func() block { return p.container(root.children, context{}).lines }
	var split []*node

	for {
		wide, tag := lay().overflowing()
		switch {
		case wide != nil:
			p.expanded[*wide] = true
		case tag != nil:
			p.split[tag] = true
			split = append(split, tag)
		default:
			return unsplit(p, lay, split)
		}
	}
}

func unsplit(p *printer, lay func() block, split []*node) block {
	for _, tag := range slices.Backward(split) {
		delete(p.split, tag)

		if wide, wider := lay().overflowing(); wide != nil || wider != nil {
			p.split[tag] = true
		}
	}

	return lay()
}

func sameContent(before, after string) error {
	want, err := skeleton(before)
	if err != nil {
		return err
	}

	got, err := skeleton(after)
	if err != nil || got != want {
		return errors.Join(errChanged, err)
	}

	return nil
}

func skeleton(src string) (string, error) {
	tree := parse.New("template")
	tree.Mode = parse.SkipFuncCheck | parse.ParseComments

	trees := map[string]*parse.Tree{}
	if _, err := tree.Parse(src, "", "", trees); err != nil {
		return "", err
	}

	var all strings.Builder
	for _, name := range slices.Sorted(maps.Keys(trees)) {
		all.WriteString(name + ":" + strings.Join(strings.FieldsFunc(trees[name].Root.String(), isSpace), "") + "\n")
	}

	return all.String(), nil
}

func firstDifference(a, b string) int {
	for i := range min(len(a), len(b)) {
		if a[i] != b[i] {
			return lineOf(a, i)
		}
	}

	return lineOf(a, min(len(a), len(b)))
}

func walk(root string, visit func(file, src, formatted string) error) error {
	fsys := os.DirFS(filepath.Join(root, dir))

	return fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || path.Ext(name) != ".html" {
			return err
		}

		src, err := fs.ReadFile(fsys, name)
		if err != nil {
			return err
		}

		file := filepath.Join(root, dir, name)

		formatted, err := Format(string(src))
		if err != nil {
			return fmt.Errorf("%s: %w", file, err)
		}

		return visit(file, string(src), formatted)
	})
}
