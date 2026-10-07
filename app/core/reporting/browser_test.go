package reporting

import (
	"html/template"
	"testing"
)

func TestBrowserMetaEscapesItsValues(t *testing.T) {
	meta := NewBrowser("https://public@example.com/2", `"><script>`, "v1").Meta()

	want := template.HTML(`<meta name="sentry" data-dsn="https://public@example.com/2" data-environment="&#34;&gt;&lt;script&gt;" data-release="v1">`)
	if meta != want {
		t.Errorf("got  %s\nwant %s", meta, want)
	}
}

func TestBrowserSharesTheServerRelease(t *testing.T) {
	want := `<meta name="sentry" data-dsn="https://public@example.com/2" data-environment="staging" data-release="` + template.HTMLEscapeString(buildRevision()) + `">`
	if meta := NewBrowser("https://public@example.com/2", "staging", "").Meta(); string(meta) != want {
		t.Errorf("got  %s\nwant %s", meta, want)
	}
}
