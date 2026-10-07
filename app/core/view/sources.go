package view

import (
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"maps"
	"path"
	"slices"
	"strings"
	"text/template/parse"
)

var errLonelyComponents = errors.New("templates: components without a page beside their folder")

var errNestedComponent = errors.New("templates: a folder inside components")

var errComponentDefines = errors.New("templates: a component defines a template")

type source struct {
	name, content string
}

type sources struct {
	shared []source

	components map[string][]source

	pages []source
}

func read(fsys fs.FS) ([]source, error) {
	var found []source
	err := fs.WalkDir(fsys, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || path.Ext(name) != ".html" {
			return err
		}

		content, err := fs.ReadFile(fsys, name)
		found = append(found, source{name: name, content: string(content)})

		return err
	})

	return found, err
}

func classify(all []source) (*sources, error) {
	sorted := &sources{components: map[string][]source{}}

	for _, s := range all {
		dir := path.Dir(s.name)
		switch {
		case strings.HasPrefix(s.name, "shared/"):
			sorted.shared = append(sorted.shared, s)
		case path.Base(dir) == "components":
			if err := definesNothing(s); err != nil {
				return nil, err
			}

			sorted.components[path.Dir(dir)] = append(sorted.components[path.Dir(dir)], s)
		case slices.Contains(strings.Split(dir, "/"), "components"):
			return nil, fmt.Errorf("%w: %s", errNestedComponent, s.name)
		default:
			sorted.pages = append(sorted.pages, s)
		}
	}

	for _, folder := range slices.Sorted(maps.Keys(sorted.components)) {
		if slices.ContainsFunc(sorted.pages, func(page source) bool { return path.Dir(page.name) == folder }) == false {
			return nil, fmt.Errorf("%w: %s", errLonelyComponents, path.Join(folder, "components"))
		}
	}

	return sorted, nil
}

func definesNothing(component source) error {
	tree := parse.New(component.name)
	tree.Mode = parse.SkipFuncCheck

	trees := map[string]*parse.Tree{}
	if _, err := tree.Parse(component.content, "", "", trees); err != nil {
		return fmt.Errorf("templates: %w", err)
	}

	for _, name := range slices.Sorted(maps.Keys(trees)) {
		if name != component.name {
			return fmt.Errorf("%w: %s in %s", errComponentDefines, name, component.name)
		}
	}

	return nil
}

func parseInto(set *template.Template, all []source) error {
	for _, s := range all {
		if _, err := set.New(s.name).Parse(s.content); err != nil {
			return fmt.Errorf("templates: %w", err)
		}
	}

	return nil
}
