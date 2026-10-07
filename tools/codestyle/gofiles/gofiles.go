package gofiles

import (
	"fmt"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const MaxWidth = 120

const tabWidth = 4

var generated = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.$`)

type Issue struct {
	Pos token.Position

	Msg string
}

func (i Issue) String() string {
	return fmt.Sprintf("%s: %s", i.Pos, i.Msg)
}

func Walk(root string, visit func(path string, src []byte) error) error {
	fsys := os.DirFS(root)

	return fs.WalkDir(fsys, ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if path != "." && skipDir(entry.Name()) {
				return fs.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(path, ".go") == false {
			return nil
		}

		src, err := fs.ReadFile(fsys, path)
		if err != nil || generated.Match(src) {
			return err
		}

		return visit(filepath.Join(root, path), src)
	})
}

func skipDir(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") ||
		name == "testdata" || name == "node_modules"
}

func Width(line string) int {
	columns := 0

	for _, r := range line {
		if r == '\t' {
			columns += tabWidth
		} else {
			columns++
		}
	}

	return columns
}
