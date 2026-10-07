package generator_test

import (
	"flag"
	"path/filepath"
	"testing"

	"gokick/tools/codegen/codegentest"
	"gokick/tools/tsgen/generator"
)

var update = flag.Bool("update", false, "rewrite testdata/golden from testdata/fixture")

func TestBuildMatchesTheGoldenFiles(t *testing.T) {
	files, err := generator.Build(filepath.Join("testdata", "fixture"), "fixture")
	if err != nil {
		t.Fatal(err)
	}

	codegentest.MatchGolden(t, filepath.Join("testdata", "golden"), files.Content, *update)
}
