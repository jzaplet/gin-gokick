package catalog

import (
	"errors"
	"fmt"
	"go/ast"

	"gokick/tools/tsgen/gosource"
)

type DTO struct {
	Ref gosource.TypeRef

	Dir Directive

	Fields []Field

	ErrorPaths []string
}

type Union struct {
	Ref gosource.TypeRef

	Dir Directive

	Members []Member
}

type Catalog struct {
	DTOs []DTO

	Unions []Union

	Directives map[gosource.TypeRef]Directive
}

func Collect(sources []gosource.Source) (Catalog, error) {
	c := Catalog{Directives: map[gosource.TypeRef]Directive{}}

	for _, src := range sources {
		for spec, doc := range src.TypeSpecs() {
			dir, ok, err := directiveOf(doc)
			if err == nil && ok {
				err = c.add(src, spec, dir)
			}

			if err != nil {
				return Catalog{}, fmt.Errorf("%s: type %s: %w", src.Path, spec.Name.Name, err)
			}
		}
	}

	for i := range c.Unions {
		members, err := membersOf(sources, c.Unions[i].Ref)
		if err != nil {
			return Catalog{}, fmt.Errorf("union %s: %w", c.Unions[i].Ref, err)
		}

		c.Unions[i].Members = members
	}

	if err := c.addErrorPaths(); err != nil {
		return Catalog{}, err
	}

	return c, nil
}

func (c *Catalog) add(src gosource.Source, spec *ast.TypeSpec, dir Directive) error {
	ref := gosource.TypeRef{Pkg: src.Pkg, Name: spec.Name.Name}

	c.Directives[ref] = dir
	if dir.Union {
		if base, ok := spec.Type.(*ast.Ident); ok == false || base.Name != "string" {
			return errors.New("a union must be a named string type")
		}

		c.Unions = append(c.Unions, Union{Ref: ref, Dir: dir})

		return nil
	}

	st, ok := spec.Type.(*ast.StructType)
	if ok == false || spec.TypeParams != nil {
		return errors.New("the directive needs a struct without type parameters, or a named string type with union")
	}

	fields, err := fieldsOf(src, st)
	if err != nil {
		return err
	}

	if len(fields) == 0 {
		return errors.New("the struct has no JSON fields")
	}

	c.DTOs = append(c.DTOs, DTO{Ref: ref, Dir: dir, Fields: fields})

	return nil
}
