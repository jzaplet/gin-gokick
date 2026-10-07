package view

import (
	"errors"
	"html/template"
	"strings"
	"testing"
	"testing/fstest"
)

const layout = `{{template "shared/layout.html" .}}`

func TestParseRefusesComponentsOutsideTheirFolder(t *testing.T) {
	t.Run("the page beside them", func(t *testing.T) {
		res := render(t, "home/page.html", struct{ Name string }{Name: "Jan"})

		if strings.Contains(res.Body.String(), `<h1>cs home.title [] Jan</h1>`) == false {
			t.Errorf("page %s", res.Body.String())
		}
	})
	t.Run("a page of another folder", func(t *testing.T) {
		var escapeErr *template.Error

		err := parseComponents(map[string]string{
			"home/page.html": layout,

			"admin/page.html": layout + `{{define "content"}}{{template "home/components/hero.html" .}}{{end}}`,
		})
		if errors.As(err, &escapeErr) == false || escapeErr.ErrorCode != template.ErrNoSuchTemplate {
			t.Errorf("error %v", err)
		}
	})
	t.Run("a page in a subfolder", func(t *testing.T) {
		var escapeErr *template.Error

		err := parseComponents(map[string]string{
			"home/page.html": layout,

			"home/admin/page.html": layout + `{{define "content"}}{{template "home/components/hero.html" .}}{{end}}`,
		})
		if errors.As(err, &escapeErr) == false || escapeErr.ErrorCode != template.ErrNoSuchTemplate {
			t.Errorf("error %v", err)
		}
	})
}

func TestParseRefusesComponentsWithoutAPageBesideTheirFolder(t *testing.T) {
	err := parseComponents(map[string]string{"admin/page.html": layout})

	if errors.Is(err, errLonelyComponents) == false || strings.HasSuffix(err.Error(), ": home/components") == false {
		t.Errorf("error %v", err)
	}
}

func TestParseRefusesAFolderInsideComponents(t *testing.T) {
	err := parseComponents(map[string]string{
		"home/page.html": layout,

		"home/components/cards/card.html": `<li>card</li>`,
	})
	if errors.Is(err, errNestedComponent) == false || strings.HasSuffix(err.Error(), ": home/components/cards/card.html") == false {
		t.Errorf("error %v", err)
	}
}

func TestParseRefusesAComponentThatDefinesATemplate(t *testing.T) {
	for kind, content := range map[string]string{
		"define": `<section>hero</section>{{define "assets"}}<script></script>{{end}}`,

		"block": `<section>{{block "assets" .}}<script></script>{{end}}</section>`,
	} {
		err := parseComponents(map[string]string{"home/page.html": layout, "home/components/hero.html": content})
		if errors.Is(err, errComponentDefines) == false || strings.HasSuffix(err.Error(), ": assets in home/components/hero.html") == false {
			t.Errorf("%s: error %v", kind, err)
		}
	}
}

func parseComponents(files map[string]string) error {
	fsys := fstest.MapFS{
		"shared/layout.html": {Data: []byte(`{{block "content" .}}{{end}}`)},

		"home/components/hero.html": {Data: []byte(`<section>hero</section>`)},
	}
	for name, content := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(content)}
	}

	_, err := Parse(fsys, nil, texts)

	return err
}
