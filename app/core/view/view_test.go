package view

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"

	"gokick/app/core/csp"
	"gokick/app/core/locale"
	"gokick/app/core/testkit"
)

var templates = fstest.MapFS{
	"shared/layout.html": {Data: []byte(`<title>{{block "title" .}}Site{{end}}</title><script nonce="{{.Nonce}}"></script>{{block "content" .}}{{end}}`)},

	"shared/name.html": {Data: []byte(`<b>{{.}}</b>`)},

	"page.html": {Data: []byte(`{{template "shared/layout.html" .}}{{define "content"}}<p>{{template "shared/name.html" .Params.Name}}</p>{{end}}`)},

	"admin/page.html": {Data: []byte(`{{template "shared/layout.html" .}}{{define "title"}}Admin{{end}}{{define "content"}}<p>admin</p>{{end}}`)},

	"broken.html": {Data: []byte(`{{template "shared/layout.html" .}}{{define "content"}}{{.Params.Missing}}{{end}}`)},

	"texts.html": {Data: []byte(`{{template "shared/layout.html" .}}{{define "title"}}{{t "home.title"}}{{end}}{{define "content"}}<img alt="{{t "files.count" "n" 3}}">{{end}}`)},

	"home/components/greeting.html": {Data: []byte(`<h1>{{t "home.title"}} {{.Params.Name}}</h1>`)},

	"home/page.html": {Data: []byte(`{{template "shared/layout.html" .}}{{define "content"}}{{template "home/components/greeting.html" .}}{{end}}`)},
}

var funcs = template.FuncMap{"shout": strings.ToUpper}

var texts = map[string]Translate{"cs_CZ": textsOf("cs"), "en_US": textsOf("en")}

func TestPageGoesIntoItsLayout(t *testing.T) {
	res := render(t, "page.html", struct{ Name string }{Name: `<i>"Jan"</i>`})

	if want := `<title>Site</title><script nonce="abc"></script><p><b>&lt;i&gt;&#34;Jan&#34;&lt;/i&gt;</b></p>`; res.Body.String() != want {
		t.Errorf("got  %s\nwant %s", res.Body.String(), want)
	}

	if res.Code != http.StatusCreated || res.Header().Get("Content-Type") != "text/html; charset=utf-8" || res.Header().Get("Cache-Control") != "private, no-cache" {
		t.Errorf("status %d, headers %v", res.Code, res.Header())
	}
}

func TestEachPageOverridesTheLayoutBlocksOnItsOwn(t *testing.T) {
	if res := render(t, "admin/page.html", nil); res.Body.String() != `<title>Admin</title><script nonce="abc"></script><p>admin</p>` {
		t.Errorf("admin page %s", res.Body.String())
	}

	if res := render(t, "page.html", struct{ Name string }{Name: "Jan"}); strings.HasPrefix(res.Body.String(), "<title>Site</title>") == false || strings.HasSuffix(res.Body.String(), "<p><b>Jan</b></p>") == false {
		t.Errorf("page %s", res.Body.String())
	}
}

func TestLiteralsAreTheStringsEveryPagePassesToAFunc(t *testing.T) {
	renderer, err := Parse(fstest.MapFS{
		"shared/layout.html": {Data: []byte(`{{block "head" .}}{{shout "layout"}}{{end}}{{block "content" .}}{{end}}`)},

		"page.html": {Data: []byte(`{{template "shared/layout.html" .}}{{define "content"}}{{if .}}{{shout "page"}}{{end}}{{template "part"}}{{end}}{{define "part"}}{{shout "part"}}{{end}}`)},

		"other.html": {Data: []byte(`{{template "shared/layout.html" .}}{{define "head"}}{{shout "override"}}{{end}}`)},

		"components/unused.html": {Data: []byte(`{{shout "component"}}`)},
	}, funcs, texts)
	if err != nil {
		t.Fatal(err)
	}

	got := renderer.Literals("shout")
	slices.Sort(got)

	if want := []string{"component", "layout", "override", "page", "part"}; slices.Equal(got, want) == false {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestPageWritesTheTextsOfItsLanguage(t *testing.T) {
	for path, lang := range map[string]string{"/": "cs", "/en/page": "en"} {
		want := `<title>` + lang + ` home.title []</title><script nonce="abc"></script><img alt="` + lang + ` files.count [n 3]">`
		if res := serve(t, path, "texts.html", nil); res.Code != http.StatusCreated || res.Body.String() != want {
			t.Errorf("%s: status %d, got  %s\nwant %s", path, res.Code, res.Body.String(), want)
		}
	}
}

func TestPageAnswers500WithoutAHalfWrittenPage(t *testing.T) {
	for _, name := range []string{"broken.html", "missing.html", "shared/layout.html", "home/components/greeting.html"} {
		if res := render(t, name, struct{ Name string }{}); res.Code != http.StatusInternalServerError || res.Body.Len() != 0 {
			t.Errorf("%s: status %d, body %q", name, res.Code, res.Body.String())
		}
	}
}

func TestParseRejectsAPageWithoutABlockItsLayoutCalls(t *testing.T) {
	_, err := Parse(fstest.MapFS{
		"shared/layout.html": {Data: []byte(`<title>{{template "title" .}}</title><meta content="{{template "description" .}}">`)},

		"page.html": {Data: []byte(`{{template "shared/layout.html" .}}{{define "title"}}Page{{end}}`)},
	}, nil, texts)

	var escapeErr *template.Error
	if errors.As(err, &escapeErr) == false || escapeErr.ErrorCode != template.ErrNoSuchTemplate {
		t.Errorf("error %v", err)
	}
}

func render(t *testing.T, name string, params any) *httptest.ResponseRecorder {
	t.Helper()

	return serve(t, "/", name, params)
}

func serve(t *testing.T, path, name string, params any) *httptest.ResponseRecorder {
	t.Helper()

	renderer, err := Parse(templates, funcs, texts)
	if err != nil {
		t.Fatal(err)
	}

	locales, err := locale.New("cs_CZ", []string{"cs_CZ", "en_US"})
	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.ContextWithFallback = true
	page := func(c *gin.Context) {
		c.Request = c.Request.WithContext(csp.WithNonce(c.Request.Context(), "abc"))
		renderer.Page(c, http.StatusCreated, name, params)
	}
	router.GET("/", locales.Home(), page)
	router.GET("/:lang/page", locales.Prefix(func(c *gin.Context) { c.Status(http.StatusNotFound) }), page)

	return testkit.Serve(router, testkit.Request(t, http.MethodGet, path))
}

func textsOf(lang string) Translate {
	return func(key string, pairs ...any) (string, error) {
		return fmt.Sprintf("%s %s %v", lang, key, pairs), nil
	}
}
