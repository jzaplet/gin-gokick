package generator_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gokick/tools/tsgen/generator"
)

func TestBuildRefusesADirectiveItCannotFollow(t *testing.T) {
	t.Run("missing name", func(t *testing.T) { expectRefusal(t, "malformed directive", "//tsgen:assets/X.ts\ntype X struct{}") })
	t.Run("unknown option", func(t *testing.T) {
		expectRefusal(t, `unknown option "maybe"`, "//tsgen:assets/X.ts X maybe\ntype X struct{}")
	})
	t.Run("path outside assets", func(t *testing.T) {
		expectRefusal(t, "must be a .ts file under assets/", "//tsgen:src/X.ts X\ntype X struct{}")
	})
	t.Run("path leaving assets", func(t *testing.T) {
		expectRefusal(t, "must be a .ts file under assets/", "//tsgen:assets/../X.ts X\ntype X struct{}")
	})
	t.Run("name", func(t *testing.T) {
		expectRefusal(t, "is not an identifier", "//tsgen:assets/X.ts X-Y\ntype X struct{}")
	})
	t.Run("named string type", func(t *testing.T) { expectRefusal(t, "needs a struct", "//tsgen:assets/X.ts X\ntype X string") })
	t.Run("union of numbers", func(t *testing.T) {
		expectRefusal(t, "a union must be a named string type", "//tsgen:assets/X.ts X union\ntype X int")
	})
	t.Run("same name twice", func(t *testing.T) {
		expectRefusal(t, "assets/X.ts: X is defined twice", "//tsgen:assets/X.ts X\ntype A struct{ A string `json:\"a\"` }\n\n//tsgen:assets/X.ts X\ntype B struct{ B string `json:\"b\"` }")
	})
}

func TestBuildRefusesAFieldItCannotType(t *testing.T) {
	t.Run("no fields", func(t *testing.T) { expectRefusal(t, "no JSON fields", structWith("hidden string `json:\"hidden\"`")) })
	t.Run("json tag", func(t *testing.T) { expectRefusal(t, "field A: the field has no json tag", structWith("A string")) })
	t.Run("json name", func(t *testing.T) {
		expectRefusal(t, `"first-name" is not a TypeScript identifier`, structWith("A string `json:\"first-name\"`"))
	})
	t.Run("json name with a dot", func(t *testing.T) {
		expectRefusal(t, `"a.b" is not a TypeScript identifier`, structWith("A string `json:\"a.b\"`"))
	})
	t.Run("json name with a space", func(t *testing.T) {
		expectRefusal(t, `"a b" is not a TypeScript identifier`, structWith("A string `json:\"a b\"`"))
	})
	t.Run("json name twice", func(t *testing.T) {
		expectRefusal(t, `field B: the json name "a" is taken`, structWith("A string `json:\"a\"`\nB string `json:\"a\"`"))
	})
	t.Run("embedded", func(t *testing.T) { expectRefusal(t, "embedded fields are not supported", structWith("Other")) })
	t.Run("channel", func(t *testing.T) {
		expectRefusal(t, "unmapped Go type chan int", structWith("A chan int `json:\"a\"`"))
	})
	t.Run("any", func(t *testing.T) { expectRefusal(t, "unmapped Go type any", structWith("A any `json:\"a\"`")) })
	t.Run("typed map", func(t *testing.T) {
		expectRefusal(t, "unmapped Go type map[string]string", structWith("A map[string]string `json:\"a\"`"))
	})
	t.Run("bytes", func(t *testing.T) { expectRefusal(t, "base64", structWith("A []byte `json:\"a\"`")) })
	t.Run("nested slice", func(t *testing.T) {
		expectRefusal(t, "a guard cannot check [][]string", structWith("A [][]string `json:\"a\"`"))
	})
	t.Run("slice of pointers", func(t *testing.T) {
		expectRefusal(t, "a guard cannot check []*string", structWith("A []*string `json:\"a\"`"))
	})
	t.Run("pointer to pointer", func(t *testing.T) {
		expectRefusal(t, "a guard cannot check **string", structWith("A **string `json:\"a\"`"))
	})
	t.Run("type without directive", func(t *testing.T) {
		expectRefusal(t, "fixture/p0.Other has no //tsgen: directive", structWith("A Other `json:\"a\"`")+"\n\ntype Other struct{}")
	})
	t.Run("type without guard", func(t *testing.T) {
		expectRefusal(t, "Form has no guard (noguard or request)", structWith("A Form `json:\"a\"`")+"\n\n//tsgen:assets/Form.ts Form noguard\ntype Form struct{ A string `json:\"a\"` }")
	})
}

func TestBuildRefusesAUnionItCannotResolve(t *testing.T) {
	union := "//tsgen:assets/Key.ts Key union\ntype Key string\n\n"

	t.Run("no constants", func(t *testing.T) { expectRefusal(t, "the type has no constants", union) })
	t.Run("expression", func(t *testing.T) {
		expectRefusal(t, "const KeyA is not a string literal", union+`const KeyA Key = "a" + "b"`)
	})
	t.Run("conversion", func(t *testing.T) { expectRefusal(t, "declare KeyA as const KeyA Key", union+`const KeyA = Key("a")`) })
	t.Run("same member twice", func(t *testing.T) {
		expectRefusal(t, "fixture/p0.KeyA and fixture/p1.KeyA both become A", union+`const KeyA Key = "a"`, "import \"fixture/p0\"\n\nconst KeyA p0.Key = \"b\"")
	})
}

func TestBuildRefusesARequestItCannotCheck(t *testing.T) {
	item := "\n\n//tsgen:assets/Item.ts Item noguard\ntype Item struct{ Name string `json:\"name\" binding:\"required\"` }"

	t.Run("list of structs without dive", func(t *testing.T) {
		expectRefusal(t, "items: a list of fixture/p0.Item needs dive", requestWith("Items []Item `json:\"items\" binding:\"required\"`")+item)
	})
	t.Run("pointer to a list of structs without dive", func(t *testing.T) {
		expectRefusal(t, "items: a list of fixture/p0.Item needs dive", requestWith("Items *[]Item `json:\"items\"`")+item)
	})
	t.Run("nested list of structs without dive", func(t *testing.T) {
		box := "\n\n//tsgen:assets/Box.ts Box noguard\ntype Box struct{ Items []Item `json:\"items\"` }"
		expectRefusal(t, "box.items: a list of fixture/p0.Item needs dive", requestWith("Box *Box `json:\"box\"`")+item+box)
	})
	t.Run("itself", func(t *testing.T) {
		expectRefusal(t, "fixture/p0.X contains itself", requestWith("Next *X `json:\"next\"`"))
	})
	t.Run("nested struct that contains itself", func(t *testing.T) {
		node := "\n\n//tsgen:assets/Node.ts Node noguard\ntype Node struct{ Children []Node `json:\"children\" binding:\"dive\"` }"
		expectRefusal(t, "fixture/p0.Node contains itself", requestWith("Root Node `json:\"root\"`")+node)
	})
	t.Run("general", func(t *testing.T) {
		expectRefusal(t, `the json name "general" is kept`, requestWith("General string `json:\"general\"`"))
	})
	t.Run("errors name taken", func(t *testing.T) {
		other := "\n\n//tsgen:assets/X.ts XErrors\ntype Other struct{ A string `json:\"a\"` }"
		expectRefusal(t, "assets/X.ts: XErrors is defined twice", requestWith("A string `json:\"a\"`")+other)
	})
	t.Run("named string type", func(t *testing.T) { expectRefusal(t, "needs a struct", "//tsgen:assets/X.ts X request\ntype X string") })
	t.Run("inside a guarded type", func(t *testing.T) {
		form := "\n\n//tsgen:assets/Form.ts Form request\ntype Form struct{ A string `json:\"a\"` }"
		expectRefusal(t, "Form has no guard (noguard or request)", structWith("A Form `json:\"a\"`")+form)
	})
}

func requestWith(fields string) string {
	return "//tsgen:assets/X.ts X request\ntype X struct {\n" + fields + "\n}"
}

func structWith(fields string) string {
	return "//tsgen:assets/X.ts X\ntype X struct {\n" + fields + "\n}"
}

func expectRefusal(t *testing.T, want string, packages ...string) {
	t.Helper()

	dir := t.TempDir()
	for i, src := range packages {
		pkgDir := filepath.Join(dir, fmt.Sprintf("p%d", i))
		if err := os.MkdirAll(pkgDir, 0o750); err != nil {
			t.Fatal(err)
		}

		code := fmt.Sprintf("package p%d\n\n%s\n", i, src)
		if err := os.WriteFile(filepath.Join(pkgDir, "x.go"), []byte(code), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := generator.Build(dir, "fixture"); err == nil || strings.Contains(err.Error(), want) == false {
		t.Errorf("error %v, want %q", err, want)
	}
}
