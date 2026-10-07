package webroot

import (
	"encoding/json"
	"io/fs"
	"path"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"text/template/parse"

	"gokick/app/core/view/literals"
)

func Public(tb testing.TB, names ...string) fstest.MapFS {
	tb.Helper()

	manifest := map[string]map[string]any{}
	public := fstest.MapFS{}

	for _, name := range names {
		ext, entry := path.Ext(name), true
		switch ext {
		case ".css":
		case ".ts", ".js":
			ext = ".js"
		default:
			entry = false
		}

		file := "assets/" + strings.TrimSuffix(path.Base(name), path.Ext(name)) + "-test" + ext
		manifest[name] = map[string]any{"file": file, "isEntry": entry}
		public["build/"+file] = &fstest.MapFile{Data: []byte(name)}
	}

	data, err := json.Marshal(manifest)
	if err != nil {
		tb.Fatal(err)
	}

	public["build/manifest.json"] = &fstest.MapFile{Data: data}

	return public
}

func TemplateLiterals(tb testing.TB, templates fs.FS, funcs ...string) []string {
	tb.Helper()
	var found []string

	err := fs.WalkDir(templates, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || path.Ext(name) != ".html" {
			return err
		}

		content, err := fs.ReadFile(templates, name)
		if err != nil {
			return err
		}

		tree := parse.New(name)
		tree.Mode = parse.SkipFuncCheck

		trees := map[string]*parse.Tree{}
		if _, err := tree.Parse(string(content), "", "", trees); err != nil {
			return err
		}

		for _, parsed := range trees {
			for _, fn := range funcs {
				found = append(found, literals.Of(parsed.Root, fn)...)
			}
		}

		return nil
	})
	if err != nil {
		tb.Fatal(err)
	}

	slices.Sort(found)

	return slices.Compact(found)
}
