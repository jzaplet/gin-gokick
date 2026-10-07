package codegentest

import (
	"bytes"
	"io/fs"
	"maps"
	"os"
	"path"
	"path/filepath"
	"slices"
	"testing"
)

func MatchGolden(t *testing.T, dir string, content map[string][]byte, update bool) {
	t.Helper()

	files := map[string][]byte{}

	for _, rel := range slices.Sorted(maps.Keys(content)) {
		name := path.Base(rel) + ".golden"
		if _, taken := files[name]; taken {
			t.Fatalf("%s and another generated file share the golden file %s", rel, name)
		}

		files[name] = content[rel]
	}

	if update {
		for name, body := range files {
			if err := os.WriteFile(filepath.Join(dir, name), body, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}

	golden := os.DirFS(dir)

	entries, err := fs.Glob(golden, "*.golden")
	if err != nil {
		t.Fatal(err)
	}

	if names := slices.Sorted(maps.Keys(files)); slices.Equal(names, entries) == false {
		t.Fatalf("generated %v, golden files %v", names, entries)
	}

	for name, body := range files {
		want, err := fs.ReadFile(golden, name)
		if err != nil {
			t.Fatal(err)
		}

		if bytes.Equal(body, want) == false {
			t.Errorf("%s:\n%s\nwant:\n%s", name, body, want)
		}
	}
}

func WriteFile(t *testing.T, root, rel, content string) {
	t.Helper()

	target := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(target, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
