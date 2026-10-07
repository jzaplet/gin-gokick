package literals

import (
	"slices"
	"strings"
	"testing"
	"text/template"
)

func TestOfFindsTheStringsPassedToAFunc(t *testing.T) {
	tmpl, err := template.New("page").Funcs(template.FuncMap{"shout": strings.ToUpper}).Parse(`{{shout "top" "two"}}{{if .}}{{shout "if"}}{{else}}{{shout "else"}}{{end}}{{range .}}{{with .}}{{shout "with"}}{{end}}{{end}}{{printf "%s" (shout "nested")}}{{template "part" (shout "argument")}}{{shout .Name}}{{printf "%s" "printf"}}`)
	if err != nil {
		t.Fatal(err)
	}

	got := Of(tmpl.Root, "shout")
	slices.Sort(got)

	if want := []string{"argument", "else", "if", "nested", "top", "two", "with"}; slices.Equal(got, want) == false {
		t.Errorf("got %v, want %v", got, want)
	}
}
