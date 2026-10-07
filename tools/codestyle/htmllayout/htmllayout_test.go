package htmllayout

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRewriteLaysOutOnlyTheTemplatesThatNeedIt(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "views", "home", "home.html"), "<div><p>a</p></div>\n")
	write(t, filepath.Join(root, "views", "shared", "footer.html"), "<footer>\n    <p>a</p>\n</footer>\n")
	write(t, filepath.Join(root, "views", "notes.txt"), "<div><p>a</p></div>\n")

	changed, err := Rewrite(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(changed) != 1 || changed[0] != filepath.Join(root, "views", "home", "home.html") {
		t.Errorf("changed %v", changed)
	}

	if issues, err := Issues(root); err != nil || len(issues) != 0 {
		t.Errorf("issues %v, err %v after the rewrite", issues, err)
	}
}

func TestIssuesPointAtTheFirstLineToChange(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "views", "page.html"), "{{define \"content\"}}\n<main><p>a</p></main>\n{{end}}\n")

	issues, err := Issues(root)
	if err != nil {
		t.Fatal(err)
	}

	if len(issues) != 1 || issues[0].Pos.Line != 2 || strings.HasSuffix(issues[0].Pos.Filename, "page.html") == false {
		t.Errorf("issues %v", issues)
	}
}

func TestFormatRefusesWhatItCannotNest(t *testing.T) {
	for name, src := range map[string]string{
		"an element closed outside its block": "{{if .A}}<div>{{end}}</div>",

		"an end tag without its element": "<div></span></div>",

		"a block of the template left open": "<div>{{range .}}</div>",

		"an else outside a block": "<div>{{else}}</div>",

		"an element left open": "<main><div></main>",
	} {
		if _, err := Format(src); errors.Is(err, errNesting) == false {
			t.Errorf("%s: error %v", name, err)
		}
	}
}

func TestFormatRefusesWhatItCannotRead(t *testing.T) {
	for name, src := range map[string]string{
		"an action": `<p>{{t "a"</p>`,

		"a tag": `<p class="a"`,

		"a string inside an action": `{{t "a}}`,

		"a raw text element": `<script>a`,
	} {
		if _, err := Format(src); errors.Is(err, errUnclosed) == false {
			t.Errorf("%s: error %v", name, err)
		}
	}
}

func TestSameContentCatchesALostOrMovedText(t *testing.T) {
	if err := sameContent("<p>{{t \"a\"}} b</p>", "<p>\n    {{t \"a\"}}\n    b\n</p>"); err != nil {
		t.Errorf("whitespace only: %v", err)
	}

	for _, changed := range []string{"<p>{{t \"a\"}}</p>", "<p>b {{t \"a\"}}</p>", "<p>{{t \"b\"}} b</p>"} {
		if err := sameContent("<p>{{t \"a\"}} b</p>", changed); errors.Is(err, errChanged) == false {
			t.Errorf("%s: error %v", changed, err)
		}
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
