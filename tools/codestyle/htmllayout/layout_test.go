package htmllayout

import (
	"strings"
	"testing"
)

func TestFormatPutsBlocksOnTheirOwnLines(t *testing.T) {
	t.Run("nested blocks", func(t *testing.T) {
		expectLayout(t, "<div><p>{{t \"a\"}}</p><p>b</p></div>\n", "<div>\n    <p>{{t \"a\"}}</p>\n    <p>b</p>\n</div>\n")
	})
	t.Run("a block of the template", func(t *testing.T) {
		expectLayout(t, "{{if .A}}<div>a</div>{{else}}<div>b</div>{{end}}", "{{if .A}}\n    <div>a</div>\n{{else}}\n    <div>b</div>\n{{end}}")
	})
	t.Run("everything inside head and lists", func(t *testing.T) {
		expectLayout(t, "<ul>{{range .}}{{if .}}<li>a</li>{{end}}{{end}}</ul>", "<ul>\n    {{range .}}\n        {{if .}}\n            <li>a</li>\n        {{end}}\n    {{end}}\n</ul>")
	})
	t.Run("a blank line between blocks", func(t *testing.T) {
		expectLayout(t, "<div>\n\n<p>a</p>\n\n\n<p>b</p>\n\n</div>", "<div>\n    <p>a</p>\n\n    <p>b</p>\n</div>")
	})
	t.Run("the definitions of a page", func(t *testing.T) {
		expectLayout(t, "{{template \"layout.html\" .}}\n\n{{define \"content\"}}<main>a</main>{{end}}\n", "{{template \"layout.html\" .}}\n\n{{define \"content\"}}\n    <main>a</main>\n{{end}}\n")
	})
}

func TestFormatKeepsInlineContentOnItsLine(t *testing.T) {
	t.Run("text and inline elements", func(t *testing.T) {
		expectLayout(t, "<p>\n{{t \"a\"}}\n<a href=\"/\">b</a>\n</p>", "<p>{{t \"a\"}} <a href=\"/\">b</a></p>")
	})
	t.Run("a block of the template with inline content", func(t *testing.T) {
		expectLayout(t, "<div><p>a</p>{{if .A}}<button>b</button>{{end}}</div>", "<div>\n    <p>a</p>\n    {{if .A}}<button>b</button>{{end}}\n</div>")
	})
	t.Run("a definition the layout may write into an attribute, even a wide one", func(t *testing.T) {
		title := `{{define "title"}}` + strings.Repeat(`{{t "a"}} | `, 12) + `{{t "b"}}{{end}}`
		expectLayout(t, title, title)
	})
	t.Run("an empty element", func(t *testing.T) {
		expectLayout(t, "<div><div data-constellation></div><span class=\"a\"></span></div>", "<div>\n    <div data-constellation></div>\n    <span class=\"a\"></span>\n</div>")
	})
	t.Run("a range inside an attribute", func(t *testing.T) {
		expectLayout(t, `<meta data-homes="{{range $i, $h := .Homes}}{{if $i}} {{end}}{{$h.Path}}{{end}}">`, `<meta data-homes="{{range $i, $h := .Homes}}{{if $i}} {{end}}{{$h.Path}}{{end}}">`)
	})
	t.Run("the content of title and script", func(t *testing.T) {
		expectLayout(t, "<head><title> {{template \"title\" .}} </title><script>\n  if (a < b) {}\n</script></head>", "<head>\n    <title> {{template \"title\" .}} </title>\n    <script>\n  if (a < b) {}\n</script>\n</head>")
	})
}

func TestFormatBreaksOnlyWhereTheWhitespaceDoesNotShow(t *testing.T) {
	t.Run("glued inline content", func(t *testing.T) {
		expectLayout(t, "<span><b>a</b>{{if .A}}<i>b</i>{{end}}</span>", "<span><b>a</b>{{if .A}}<i>b</i>{{end}}</span>")
	})
	t.Run("a script, which shows no box", func(t *testing.T) {
		expectLayout(t, "<p><span>a</span><script>b</script><span>c</span></p>", "<p><span>a</span><script>b</script><span>c</span></p>")
	})
	t.Run("an action that trims", func(t *testing.T) {
		long := strings.Repeat("x", 95)
		expectLayout(t, `<span class="`+long+`">{{- t "a" -}}<b>b</b></span>`, "<span class=\""+long+"\">\n    {{- t \"a\" -}}\n    <b>b</b></span>")
	})
	t.Run("an action with quotes inside an attribute", func(t *testing.T) {
		expectLayout(t, `<div><meta content="{{template "title" .}}"><p title='{{t "a"}}'>a</p></div>`, "<div>\n    <meta content=\"{{template \"title\" .}}\">\n    <p title='{{t \"a\"}}'>a</p>\n</div>")
	})
}

func TestFormatNarrowsLinesWiderThanTheLimit(t *testing.T) {
	long := strings.Repeat("x", 100)

	t.Run("inline content breaks at its spaces", func(t *testing.T) {
		expectLayout(t, `<h1 class="`+long+`">{{t "a"}} <span>b</span> {{t "c"}}</h1>`, "<h1 class=\""+long+"\">\n    {{t \"a\"}}\n    <span>b</span>\n    {{t \"c\"}}\n</h1>")
	})
	t.Run("a block tag puts each attribute on its own line", func(t *testing.T) {
		expectLayout(t, `<div id="a" class="`+long+`"><p>a</p></div>`, "<div\n    id=\"a\"\n    class=\""+long+"\"\n>\n    <p>a</p>\n</div>")
	})
	t.Run("an inline tag keeps its content after the bracket", func(t *testing.T) {
		expectLayout(t, `<p><span><b>a</b><a href="/" class="`+long+`">b</a></span></p>`, "<p>\n    <span><b>a</b><a\n        href=\"/\"\n        class=\""+long+"\"\n    >b</a></span>\n</p>")
	})
	t.Run("a tag that fits joins its attributes again", func(t *testing.T) {
		expectLayout(t, "<div\n    id=\"a\"\n    class=\"b\"\n>\n<p>a</p>\n</div>", "<div id=\"a\" class=\"b\">\n    <p>a</p>\n</div>")
	})
	t.Run("a void tag", func(t *testing.T) {
		expectLayout(t, `<head><meta name="a" content="`+long+`"></head>`, "<head>\n    <meta\n        name=\"a\"\n        content=\""+long+"\"\n    >\n</head>")
	})
}

func expectLayout(t *testing.T, src, want string) {
	t.Helper()

	got, err := Format(src)
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}

	if again, err := Format(got); err != nil || again != got {
		t.Errorf("not stable: %v\n%s", err, again)
	}
}
