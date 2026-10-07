package splitliterals

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

func Issues(root string) ([]gofiles.Issue, error) {
	var issues []gofiles.Issue

	err := gofiles.Walk(root, func(path string, src []byte) error {
		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, lit := range longLiterals(fset.File(file.Pos()), file, src) {
			issues = append(issues, gofiles.Issue{
				Pos: fset.Position(lit.Lbrace),

				Msg: fmt.Sprintf("keyed literal on a line wider than %d columns", gofiles.MaxWidth),
			})
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
		formatted, err := splitLong(src)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}

		if bytes.Equal(formatted, src) {
			return nil
		}

		changed = append(changed, path)

		return os.WriteFile(path, formatted, 0o600)
	})
	if err != nil {
		return nil, err
	}

	return changed, nil
}

func splitLong(src []byte) ([]byte, error) {
	for {
		fset := token.NewFileSet()

		file, err := parser.ParseFile(fset, "", src, parser.ParseComments)
		if err != nil {
			return nil, err
		}

		tokens := fset.File(file.Pos())

		lits := longLiterals(tokens, file, src)
		if len(lits) == 0 {
			return src, nil
		}

		for _, lit := range slices.Backward(lits) {
			src = splitLiteral(src, tokens, lit)
		}

		if src, err = format.Source(src); err != nil {
			return nil, err
		}
	}
}

func longLiterals(tokens *token.File, file *ast.File, src []byte) []*ast.CompositeLit {
	var found []*ast.CompositeLit
	lines := map[int]bool{}

	ast.Inspect(file, func(node ast.Node) bool {
		lit, ok := node.(*ast.CompositeLit)
		if ok == false || splittable(tokens, lit, file.Comments) == false {
			return true
		}

		line := tokens.Line(lit.Lbrace)
		if lines[line] || gofiles.Width(string(lineAt(tokens, src, line))) <= gofiles.MaxWidth {
			return true
		}

		lines[line] = true

		found = append(found, lit)

		return false
	})

	return found
}

func splittable(tokens *token.File, lit *ast.CompositeLit, comments []*ast.CommentGroup) bool {
	if len(lit.Elts) < 2 || tokens.Line(lit.Lbrace) != tokens.Line(lit.Rbrace) {
		return false
	}

	for _, elt := range lit.Elts {
		if _, keyed := elt.(*ast.KeyValueExpr); keyed == false {
			return false
		}
	}

	for _, comment := range comments {
		if comment.Pos() > lit.Lbrace && comment.End() < lit.Rbrace {
			return false
		}
	}

	return true
}

func splitLiteral(src []byte, tokens *token.File, lit *ast.CompositeLit) []byte {
	var out bytes.Buffer
	out.Write(src[:tokens.Offset(lit.Lbrace)+1])

	for i, elt := range lit.Elts {
		if i > 0 {
			out.WriteString("\n")
		}

		out.WriteString("\n")
		out.Write(src[tokens.Offset(elt.Pos()):tokens.Offset(elt.End())])
		out.WriteString(",")
	}

	out.WriteString("\n")
	out.Write(src[tokens.Offset(lit.Rbrace):])

	return out.Bytes()
}

func lineAt(tokens *token.File, src []byte, line int) []byte {
	start := tokens.Offset(tokens.LineStart(line))

	end := len(src)
	if line < tokens.LineCount() {
		end = tokens.Offset(tokens.LineStart(line + 1))
	}

	return bytes.TrimRight(src[start:end], "\n")
}
