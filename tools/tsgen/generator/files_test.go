package generator_test

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gokick/tools/codegen/codegentest"
	"gokick/tools/tsgen/generator"
	"gokick/tools/tsgen/typescript"
)

func TestGenerateWritesTheTypesAndRemovesOrphans(t *testing.T) {
	root := t.TempDir()
	codegentest.WriteFile(t, root, "app/x/x.go", "package x\n\n//tsgen:assets/x/X.ts X\ntype X struct {\n\tA string `json:\"a\"`\n}\n")
	codegentest.WriteFile(t, root, "assets/old/Old.ts", typescript.Header+"\nexport type Old = { a: string };\n")
	codegentest.WriteFile(t, root, "assets/hand/Hand.ts", "export type Hand = { a: string };\n")

	changed, err := generator.Generate(root)
	if err != nil {
		t.Fatal(err)
	}

	if want := []string{"wrote assets/x/X.ts", "removed assets/old/Old.ts"}; slices.Equal(changed, want) == false {
		t.Errorf("changed %v, want %v", changed, want)
	}

	if _, err := os.Stat(filepath.Join(root, "assets", "hand", "Hand.ts")); err != nil {
		t.Errorf("a hand-written file is gone: %v", err)
	}

	if changed, err := generator.Generate(root); err != nil || len(changed) > 0 {
		t.Errorf("a second run changed %v, %v", changed, err)
	}
}
