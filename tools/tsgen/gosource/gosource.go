package gosource

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"iter"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

type Source struct {
	Path string

	Pkg string

	File *ast.File

	Imports map[string]string
}

type TypeRef struct {
	Pkg string

	Name string
}

func (r TypeRef) String() string {
	return r.Pkg + "." + r.Name
}

func Parse(dir, basePkg string) ([]Source, error) {
	fset := token.NewFileSet()
	var sources []Source
	err := filepath.WalkDir(dir, func(file string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || strings.HasSuffix(file, ".go") == false || strings.HasSuffix(file, "_test.go") {
			return err
		}

		parsed, err := parser.ParseFile(fset, file, nil, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(dir, filepath.Dir(file))
		if err != nil {
			return err
		}

		sources = append(sources, Source{
			Path: file,

			Pkg: path.Join(basePkg, filepath.ToSlash(rel)),

			File: parsed,

			Imports: importsOf(parsed),
		})

		return nil
	})

	return sources, err
}

func importsOf(file *ast.File) map[string]string {
	imports := map[string]string{}

	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			continue
		}

		name := path.Base(importPath)
		if spec.Name != nil {
			name = spec.Name.Name
		}

		imports[name] = importPath
	}

	return imports
}

func (s Source) Ref(expr ast.Expr) (TypeRef, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		return TypeRef{Pkg: s.Pkg, Name: t.Name}, true
	case *ast.SelectorExpr:
		x, ok := t.X.(*ast.Ident)
		if ok == false {
			return TypeRef{}, false
		}

		pkg, ok := s.Imports[x.Name]

		return TypeRef{Pkg: pkg, Name: t.Sel.Name}, ok
	default:
		return TypeRef{}, false
	}
}

func (s Source) TypeSpecs() iter.Seq2[*ast.TypeSpec, *ast.CommentGroup] {
	return func(yield func(*ast.TypeSpec, *ast.CommentGroup) bool) {
		for decl := range s.genDecls(token.TYPE) {
			for _, spec := range decl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if ok == false {
					continue
				}

				doc := typeSpec.Doc
				if doc == nil && decl.Lparen.IsValid() == false {
					doc = decl.Doc
				}

				if yield(typeSpec, doc) == false {
					return
				}
			}
		}
	}
}

func (s Source) ConstSpecs() iter.Seq[*ast.ValueSpec] {
	return func(yield func(*ast.ValueSpec) bool) {
		for decl := range s.genDecls(token.CONST) {
			for _, spec := range decl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if ok && yield(valueSpec) == false {
					return
				}
			}
		}
	}
}

func (s Source) genDecls(tok token.Token) iter.Seq[*ast.GenDecl] {
	return func(yield func(*ast.GenDecl) bool) {
		for _, decl := range s.File.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if ok && genDecl.Tok == tok && yield(genDecl) == false {
				return
			}
		}
	}
}
