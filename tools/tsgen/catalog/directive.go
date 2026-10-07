package catalog

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"path"
	"strings"
)

const Prefix = "//tsgen:"

type Directive struct {
	Path string

	Name string

	NoGuard bool

	Union bool

	Request bool
}

func (d Directive) Module() string {
	return "@/" + strings.TrimSuffix(strings.TrimPrefix(d.Path, "assets/"), ".ts")
}

func directiveOf(doc *ast.CommentGroup) (Directive, bool, error) {
	if doc == nil {
		return Directive{}, false, nil
	}

	for _, comment := range doc.List {
		rest, ok := strings.CutPrefix(comment.Text, Prefix)
		if ok == false {
			continue
		}

		dir, err := parseDirective(strings.Fields(rest))
		if err != nil {
			return Directive{}, false, fmt.Errorf("%s: %w", comment.Text, err)
		}

		return dir, true, nil
	}

	return Directive{}, false, nil
}

func parseDirective(words []string) (Directive, error) {
	if len(words) < 2 || len(words) > 3 {
		return Directive{}, errors.New("malformed directive, want " + Prefix + "<path> <Name> [noguard|union|request]")
	}

	dir := Directive{Path: words[0], Name: words[1]}
	if len(words) == 3 {
		switch words[2] {
		case "noguard":
			dir.NoGuard = true
		case "union":
			dir.Union = true
		case "request":
			dir.Request = true
			dir.NoGuard = true
		default:
			return Directive{}, fmt.Errorf("malformed directive, unknown option %q", words[2])
		}
	}

	if path.Clean(dir.Path) != dir.Path || strings.HasPrefix(dir.Path, "assets/") == false || strings.HasSuffix(dir.Path, ".ts") == false {
		return Directive{}, fmt.Errorf("the path %q must be a .ts file under assets/", dir.Path)
	}

	if token.IsIdentifier(dir.Name) == false {
		return Directive{}, fmt.Errorf("the name %q is not an identifier", dir.Name)
	}

	return dir, nil
}
