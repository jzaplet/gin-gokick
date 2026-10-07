package events

import (
	"go/ast"
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"testing"
)

func TestAdsConversionsListsEveryConversion(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "events.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	var declared []AdsConversion

	for node := range ast.Preorder(file) {
		spec, ok := node.(*ast.ValueSpec)
		if ok == false || isAdsConversion(spec.Type) == false {
			continue
		}

		for _, expr := range spec.Values {
			lit, ok := expr.(*ast.BasicLit)
			if ok == false || lit.Kind != token.STRING {
				t.Fatalf("%s is no string literal", spec.Names[0].Name)
			}

			value, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatal(err)
			}

			declared = append(declared, AdsConversion(value))
		}
	}

	if len(declared) == 0 || slices.Equal(slices.Sorted(slices.Values(declared)), slices.Sorted(slices.Values(AdsConversions))) == false {
		t.Errorf("constants %v, AdsConversions %v", declared, AdsConversions)
	}
}

func isAdsConversion(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)

	return ok && ident.Name == "AdsConversion"
}
