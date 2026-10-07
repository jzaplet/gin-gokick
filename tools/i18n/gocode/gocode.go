package gocode

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"

	"gokick/app/core/i18n/icu"
)

var ErrNotConstant = errors.New("a message key comes from no constant of its type")

var ErrNoText = errors.New("a message key constant has no text in the dictionaries")

var ErrLocalConstant = errors.New("a message key constant lives in a function, where tsgen does not collect it")

var ErrNoKeys = errors.New("no constant has the key type")

type listed struct {
	ImportPath string

	Dir string

	GoFiles []string

	Export string

	DepOnly bool
}

type checker struct {
	dir string

	key reflect.Type

	texts map[string]map[string]icu.Kind

	fset *token.FileSet

	imports types.Importer

	keys []string

	issues []error
}

func Check(dir string, key reflect.Type, texts map[string]map[string]icu.Kind) ([]string, error) {
	root, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	packages, err := list(root)
	if err != nil {
		return nil, err
	}

	c := checker{dir: root, key: key, texts: texts, fset: token.NewFileSet()}

	c.imports = importer.ForCompiler(c.fset, "gc", exports(packages))
	for _, pkg := range packages {
		if pkg.DepOnly {
			continue
		}

		if err := c.check(&pkg); err != nil {
			return nil, fmt.Errorf("%s: %w", pkg.ImportPath, err)
		}
	}

	if len(c.keys) == 0 {
		return nil, fmt.Errorf("%w %s.%s", ErrNoKeys, key.PkgPath(), key.Name())
	}

	return c.keys, errors.Join(c.issues...)
}

func list(dir string) ([]listed, error) {
	cmd := exec.CommandContext(context.Background(), "go", "list", "-export", "-deps", "-json=ImportPath,Dir,GoFiles,Export,DepOnly", "./...")
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w: %s", err, stderr.String())
	}

	var packages []listed
	decoder := json.NewDecoder(bytes.NewReader(out))

	for {
		var pkg listed

		err := decoder.Decode(&pkg)
		if errors.Is(err, io.EOF) {
			return packages, nil
		}

		if err != nil {
			return nil, err
		}

		packages = append(packages, pkg)
	}
}

func exports(packages []listed) importer.Lookup {
	files := map[string]string{}
	for _, pkg := range packages {
		files[pkg.ImportPath] = pkg.Export
	}

	return func(path string) (io.ReadCloser, error) {
		file := files[path]
		if file == "" {
			return nil, fmt.Errorf("go list gave no export data of %s", path)
		}

		return os.DirFS(filepath.Dir(file)).Open(filepath.Base(file))
	}
}

func (c *checker) check(pkg *listed) error {
	files := make([]*ast.File, 0, len(pkg.GoFiles))
	for _, name := range pkg.GoFiles {
		file, err := parser.ParseFile(c.fset, filepath.Join(pkg.Dir, name), nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}

		files = append(files, file)
	}

	info := &types.Info{
		Types: map[ast.Expr]types.TypeAndValue{},

		Defs: map[*ast.Ident]types.Object{},

		Uses: map[*ast.Ident]types.Object{},
	}

	config := types.Config{Importer: c.imports}
	if _, err := config.Check(pkg.ImportPath, c.fset, files, info); err != nil {
		return err
	}

	for _, file := range files {
		c.checkFile(file, info)
	}

	return nil
}

func (c *checker) checkFile(file *ast.File, info *types.Info) {
	declared := c.constants(file, info)
	ast.Inspect(file, func(node ast.Node) bool {
		expr, ok := node.(ast.Expr)
		if ok == false || declared[expr] || c.fromConstant(expr, info) {
			return true
		}

		c.issues = append(c.issues, fmt.Errorf("%s: %w: %s", c.position(expr.Pos()), ErrNotConstant, types.ExprString(expr)))

		return false
	})
}

func (c *checker) constants(file *ast.File, info *types.Info) map[ast.Expr]bool {
	declared := map[ast.Expr]bool{}

	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if ok == false {
			return true
		}

		for i, name := range spec.Names {
			key, ok := info.Defs[name].(*types.Const)
			if ok == false || c.isKey(key.Type()) == false {
				continue
			}

			text := constant.StringVal(key.Val())

			c.keys = append(c.keys, text)
			if key.Parent() != key.Pkg().Scope() {
				c.issues = append(c.issues, fmt.Errorf("%s: %w: %s", c.position(name.Pos()), ErrLocalConstant, name.Name))
			}

			if i < len(spec.Values) {
				declared[spec.Values[i]] = true
			}

			if c.texts[text] == nil {
				c.issues = append(c.issues, fmt.Errorf("%s: %w: %s %q", c.position(name.Pos()), ErrNoText, name.Name, text))
			}
		}

		return true
	})

	return declared
}

func (c *checker) fromConstant(expr ast.Expr, info *types.Info) bool {
	value, typed := info.Types[expr]
	if typed == false || value.IsType() || c.isKey(value.Type) == false {
		return true
	}

	if call, ok := expr.(*ast.CallExpr); ok && info.Types[call.Fun].IsType() {
		return false
	}

	if value.Value == nil {
		return true
	}

	key, ok := info.Uses[identOf(ast.Unparen(expr))].(*types.Const)

	return ok && c.isKey(key.Type())
}

func identOf(expr ast.Expr) *ast.Ident {
	switch e := expr.(type) {
	case *ast.Ident:
		return e
	case *ast.SelectorExpr:
		return e.Sel
	default:
		return nil
	}
}

func (c *checker) isKey(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)

	return ok && named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == c.key.PkgPath() && named.Obj().Name() == c.key.Name()
}

func (c *checker) position(pos token.Pos) string {
	position := c.fset.Position(pos)
	if rel, err := filepath.Rel(c.dir, position.Filename); err == nil {
		position.Filename = filepath.ToSlash(rel)
	}

	return position.String()
}
