package view

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path"
	"slices"

	"gokick/app/core/view/literals"
)

var errNoPages = errors.New("templates: no pages")

var errNoTexts = errors.New("templates: no texts of any locale")

var errUntranslated = errors.New("templates: t outside a locale")

type Translate func(key string, pairs ...any) (string, error)

type Renderer struct {
	pages map[string]map[string]*template.Template
}

func Parse(fsys fs.FS, funcs template.FuncMap, texts map[string]Translate) (*Renderer, error) {
	if len(texts) == 0 {
		return nil, errNoTexts
	}

	all, err := read(fsys)
	if err != nil {
		return nil, fmt.Errorf("templates: %w", err)
	}

	sorted, err := classify(all)
	if err != nil {
		return nil, err
	}

	shared := template.New("").Funcs(funcs).Funcs(template.FuncMap{"t": untranslated})
	if err := parseInto(shared, sorted.shared); err != nil {
		return nil, err
	}

	pages := map[string]map[string]*template.Template{}
	for _, s := range sorted.pages {
		page, err := shared.Clone()
		if err != nil {
			return nil, fmt.Errorf("templates: %w", err)
		}

		err = parseInto(page, slices.Concat(sorted.components[path.Dir(s.name)], []source{s}))
		if err != nil {
			return nil, err
		}

		pages[s.name], err = localize(page, s.name, texts)
		if err != nil {
			return nil, fmt.Errorf("templates: %w", err)
		}
	}

	if len(pages) == 0 {
		return nil, errNoPages
	}

	return &Renderer{pages: pages}, nil
}

func (r *Renderer) Literals(fn string) []string {
	var found []string

	for _, localized := range r.pages {
		for _, page := range localized {
			for _, t := range page.Templates() {
				if t.Tree != nil {
					found = append(found, literals.Of(t.Tree.Root, fn)...)
				}
			}
		}
	}

	slices.Sort(found)

	return slices.Compact(found)
}

func localize(page *template.Template, name string, texts map[string]Translate) (map[string]*template.Template, error) {
	localized := make(map[string]*template.Template, len(texts))
	for l, translate := range texts {
		clone, err := page.Clone()
		if err != nil {
			return nil, err
		}

		clone.Funcs(template.FuncMap{"t": translate})

		if err := escape(clone, name); err != nil {
			return nil, err
		}

		localized[l] = clone
	}

	return localized, nil
}

func untranslated(key string, _ ...any) (string, error) {
	return "", fmt.Errorf("%w: %s", errUntranslated, key)
}

func escape(page *template.Template, name string) error {
	var escapeErr *template.Error
	if err := page.ExecuteTemplate(io.Discard, name, nil); errors.As(err, &escapeErr) {
		return escapeErr
	}

	return nil
}
