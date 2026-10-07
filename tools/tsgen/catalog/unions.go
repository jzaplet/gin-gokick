package catalog

import (
	"errors"
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"gokick/tools/tsgen/gosource"
)

type Member struct {
	Name string

	Value string
}

func membersOf(sources []gosource.Source, u gosource.TypeRef) ([]Member, error) {
	var members []Member
	owners := map[string]string{}

	for _, src := range sources {
		for spec := range src.ConstSpecs() {
			found, err := constantsOf(src, spec, u)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", src.Path, err)
			}

			for _, m := range found {
				name, owner := memberName(m.Name, u.Name), src.Pkg+"."+m.Name
				if prev, taken := owners[name]; taken {
					return nil, fmt.Errorf("%s and %s both become %s", prev, owner, name)
				}

				owners[name] = owner
				members = append(members, Member{Name: name, Value: m.Value})
			}
		}
	}

	if len(members) == 0 {
		return nil, errors.New("the type has no constants, its guard would refuse every value")
	}

	return members, nil
}

func constantsOf(src gosource.Source, spec *ast.ValueSpec, u gosource.TypeRef) ([]Member, error) {
	if spec.Type == nil {
		for _, value := range spec.Values {
			if call, ok := value.(*ast.CallExpr); ok && refers(src, call.Fun, u) {
				return nil, fmt.Errorf("declare %s as const %s %s = \"…\"", spec.Names[0].Name, spec.Names[0].Name, u.Name)
			}
		}

		return nil, nil
	}

	if refers(src, spec.Type, u) == false {
		return nil, nil
	}

	if len(spec.Values) != len(spec.Names) {
		return nil, fmt.Errorf("const %s has no value of its own", spec.Names[0].Name)
	}

	constants := make([]Member, 0, len(spec.Names))
	for i, name := range spec.Names {
		value, ok := stringLiteral(spec.Values[i])
		if ok == false {
			return nil, fmt.Errorf("const %s is not a string literal", name.Name)
		}

		constants = append(constants, Member{Name: name.Name, Value: value})
	}

	return constants, nil
}

func refers(src gosource.Source, expr ast.Expr, u gosource.TypeRef) bool {
	ref, ok := src.Ref(expr)

	return ok && ref == u
}

func stringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if ok == false || lit.Kind != token.STRING {
		return "", false
	}

	value, err := strconv.Unquote(lit.Value)

	return value, err == nil
}

func memberName(constName, typeName string) string {
	rest, ok := strings.CutPrefix(constName, typeName)
	if first, _ := utf8.DecodeRuneInString(rest); ok && unicode.IsUpper(first) {
		return rest
	}

	return constName
}
