package catalog

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"gokick/tools/tsgen/gosource"
)

type Kind int

const (
	String Kind = iota
	Number
	Boolean
	Record
	Named
)

type Shape struct {
	Kind Kind

	Ref gosource.TypeRef

	Array bool

	Nullable bool
}

type Field struct {
	Name string

	Optional bool

	Dive bool

	Shape Shape
}

var textTypes = []gosource.TypeRef{
	{Pkg: "time", Name: "Time"},
	{Pkg: "github.com/google/uuid", Name: "UUID"},
	{Pkg: "net/netip", Name: "Addr"},
}

func fieldsOf(src gosource.Source, st *ast.StructType) ([]Field, error) {
	var fields []Field
	seen := map[string]bool{}

	for _, f := range st.Fields.List {
		if len(f.Names) == 0 {
			return nil, errors.New("embedded fields are not supported")
		}

		for _, name := range f.Names {
			if name.IsExported() == false {
				continue
			}

			fl, ok, err := fieldOf(src, f)
			if err != nil {
				return nil, fmt.Errorf("field %s: %w", name.Name, err)
			}

			if ok && seen[fl.Name] {
				return nil, fmt.Errorf("field %s: the json name %q is taken", name.Name, fl.Name)
			}

			if ok {
				seen[fl.Name] = true
				fields = append(fields, fl)
			}
		}
	}

	return fields, nil
}

func fieldOf(src gosource.Source, f *ast.Field) (Field, bool, error) {
	tag := ""
	if f.Tag != nil {
		tag, _ = strconv.Unquote(f.Tag.Value)
	}

	jsonTag, ok := reflect.StructTag(tag).Lookup("json")
	if ok == false {
		return Field{}, false, errors.New("the field has no json tag")
	}

	name, opts, _ := strings.Cut(jsonTag, ",")
	if name == "-" && opts == "" {
		return Field{}, false, nil
	}

	if token.IsIdentifier(name) == false && token.IsKeyword(name) == false {
		return Field{}, false, fmt.Errorf("the json name %q is not a TypeScript identifier", name)
	}

	shape, err := shapeOf(src, f.Type)
	if err != nil {
		return Field{}, false, err
	}

	optional := slices.ContainsFunc(strings.Split(opts, ","), func(opt string) bool {
		return opt == "omitempty" || opt == "omitzero"
	})
	dive := slices.Contains(strings.Split(reflect.StructTag(tag).Get("binding"), ","), "dive")

	return Field{Name: name, Optional: optional, Dive: dive, Shape: shape}, true, nil
}

func shapeOf(src gosource.Source, expr ast.Expr) (Shape, error) {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return containerOf(src, expr, t.X, false)
	case *ast.ArrayType:
		if elem, ok := t.Elt.(*ast.Ident); ok && (elem.Name == "byte" || elem.Name == "uint8") {
			return Shape{}, errors.New("[]byte marshals to a base64 string, use a string field")
		}

		return containerOf(src, expr, t.Elt, true)
	}

	if leaf, ok := leafOf(src, expr); ok {
		return leaf, nil
	}

	return Shape{}, fmt.Errorf("unmapped Go type %s", types.ExprString(expr))
}

func containerOf(src gosource.Source, expr, elem ast.Expr, array bool) (Shape, error) {
	inner, err := shapeOf(src, elem)
	if err != nil {
		return Shape{}, err
	}

	if inner.Nullable || array && inner.Array {
		return Shape{}, fmt.Errorf("a guard cannot check %s, flatten the field", types.ExprString(expr))
	}

	if array {
		inner.Array = true
	} else {
		inner.Nullable = true
	}

	return inner, nil
}

func leafOf(src gosource.Source, expr ast.Expr) (Shape, bool) {
	switch t := expr.(type) {
	case *ast.MapType:
		return Shape{Kind: Record}, types.ExprString(t) == "map[string]any"
	case *ast.Ident:
		if k, ok := scalar(t.Name); ok {
			return Shape{Kind: k}, true
		}

		ref, _ := src.Ref(t)

		return Shape{Kind: Named, Ref: ref}, types.Universe.Lookup(t.Name) == nil
	case *ast.SelectorExpr:
		ref, ok := src.Ref(t)
		if slices.Contains(textTypes, ref) {
			return Shape{Kind: String}, ok
		}

		return Shape{Kind: Named, Ref: ref}, ok
	default:
		return Shape{}, false
	}
}

func scalar(name string) (Kind, bool) {
	switch name {
	case "string":
		return String, true
	case "bool":
		return Boolean, true
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return Number, true
	default:
		return 0, false
	}
}
