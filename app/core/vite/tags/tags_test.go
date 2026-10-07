package tags_test

import (
	"testing"

	"gokick/app/core/vite/tags"
)

func TestHTMLEscapesTheNonceAndTheHrefsOnce(t *testing.T) {
	var set tags.Set
	set.Add("/build/assets/a&b.js")

	if got, want := set.HTML(`"x`), `<script type="module" nonce="&#34;x" src="/build/assets/a&amp;b.js"></script>`+"\n"; string(got) != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestASetHoldsEveryHrefOnceWithTheStylesFirst(t *testing.T) {
	var set, other tags.Set
	set.Add("/app.js")
	set.Add("/app.css")

	if set.Preload("/vue.js") == false || set.Preload("/vue.js") {
		t.Error("Preload reports a known href as new")
	}

	other.Style("/app.css")
	other.Style("/vue.css")
	other.Preload("/vue.js")
	other.Preload("/chart.js")
	other.Add("/app.js")
	other.Add("/admin.js")
	set.Merge(other)

	want := `<link rel="stylesheet" href="/app.css">
<link rel="stylesheet" href="/vue.css">
<link rel="modulepreload" nonce="n" href="/vue.js">
<link rel="modulepreload" nonce="n" href="/chart.js">
<script type="module" nonce="n" src="/app.js"></script>
<script type="module" nonce="n" src="/admin.js"></script>
`
	if got := set.HTML("n"); string(got) != want {
		t.Errorf("got\n%s\nwant\n%s", got, want)
	}
}
