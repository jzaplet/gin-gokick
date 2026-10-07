package nonegation

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"slices"

	"gokick/tools/codestyle/gofiles"
)

type negation struct {
	expr *ast.UnaryExpr

	parent ast.Node
}

type edit struct {
	start, end int

	text string
}

var inverse = map[token.Token]token.Token{token.EQL: token.NEQ, token.NEQ: token.EQL}

func Issues(root string) ([]gofiles.Issue, error) {
	var issues []gofiles.Issue

	err := gofiles.Walk(root, func(path string, src []byte) error {
		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, found := range negations(file) {
			issues = append(issues, gofiles.Issue{Pos: fset.Position(found.expr.OpPos), Msg: "negation with !"})
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return issues, nil
}

func Rewrite(root string) ([]string, error) {
	var changed []string

	err := gofiles.Walk(root, func(path string, src []byte) error {
		rewritten, err := compareExplicitly(src)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		if bytes.Equal(rewritten, src) {
			return nil
		}

		changed = append(changed, path)

		return os.WriteFile(path, rewritten, 0o600)
	})
	if err != nil {
		return nil, err
	}

	return changed, nil
}

func compareExplicitly(src []byte) ([]byte, error) {
	for {
		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		innermost := slices.DeleteFunc(negations(file), func(found negation) bool { return holdsNegation(found.expr.X) })
		if len(innermost) == 0 {
			return src, nil
		}

		tokens := fset.File(file.Pos())

		edits := make([]edit, 0, len(innermost))
		for _, found := range innermost {
			edits = append(edits, editOf(src, tokens, found))
		}

		slices.SortFunc(edits, func(a, b edit) int { return a.start - b.start })

		for _, change := range slices.Backward(edits) {
			src = slices.Concat(src[:change.start], []byte(change.text), src[change.end:])
		}

		if src, err = format.Source(src); err != nil {
			return nil, err
		}
	}
}

func negations(file *ast.File) []negation {
	var found []negation
	var parents []ast.Node

	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			parents = parents[:len(parents)-1]

			return true
		}

		if expr, ok := node.(*ast.UnaryExpr); ok && expr.Op == token.NOT {
			found = append(found, negation{expr: expr, parent: parents[len(parents)-1]})
		}

		parents = append(parents, node)

		return true
	})

	return found
}

func holdsNegation(node ast.Node) bool {
	held := false

	ast.Inspect(node, func(child ast.Node) bool {
		if expr, ok := child.(*ast.UnaryExpr); ok && expr.Op == token.NOT {
			held = true
		}

		return held == false
	})

	return held
}

func editOf(src []byte, tokens *token.File, found negation) edit {
	text := func(node ast.Node) string { return string(src[tokens.Offset(node.Pos()):tokens.Offset(node.End())]) }
	span := func(node ast.Node, replaced string) edit {
		return edit{start: tokens.Offset(node.Pos()), end: tokens.Offset(node.End()), text: replaced}
	}

	if parent, ok := found.parent.(*ast.BinaryExpr); ok && inverse[parent.Op] != token.ILLEGAL {
		switch {
		case parent.X == found.expr && holdsNegation(parent.Y) == false:
			return span(parent, text(found.expr.X)+" "+inverse[parent.Op].String()+" "+text(parent.Y))
		case parent.Y == found.expr && holdsNegation(parent.X) == false:
			return span(parent, text(parent.X)+" "+inverse[parent.Op].String()+" "+text(found.expr.X))
		}
	}

	return span(found.expr, replacement(text, found))
}

func replacement(text func(ast.Node) string, found negation) string {
	compared := text(found.expr.X) + " == false"
	if paren, ok := found.expr.X.(*ast.ParenExpr); ok {
		if binary, ok := paren.X.(*ast.BinaryExpr); ok && inverse[binary.Op] != token.ILLEGAL {
			compared = text(binary.X) + " " + inverse[binary.Op].String() + " " + text(binary.Y)
		}
	}

	if bindsTighter(found.parent) {
		return "(" + compared + ")"
	}

	return compared
}

func bindsTighter(parent ast.Node) bool {
	switch parent := parent.(type) {
	case *ast.BinaryExpr:
		return parent.Op.Precedence() >= token.EQL.Precedence()
	case *ast.UnaryExpr, *ast.StarExpr, *ast.SelectorExpr:
		return true
	default:
		return false
	}
}
