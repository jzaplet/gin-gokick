package manifest_test

import (
	"errors"
	"testing"
	"testing/fstest"

	"gokick/app/core/vite/manifest"
)

const built = `{
	"assets/admin.ts": {"file": "assets/admin-5.js", "isEntry": true, "css": ["assets/admin-3.css"], "imports": ["_vue-4.js", "_chart-6.js"]},
	"_vue-4.js": {"file": "assets/vue-4.js", "css": ["assets/vue-7.css"]},
	"_chart-6.js": {"file": "assets/chart-6.js", "imports": ["_vue-4.js"]},
	"assets/img/logo.svg": {"file": "assets/logo-8.svg", "src": "assets/img/logo.svg"}
}`

func TestAnEntryLoadsItsImportsAndTheirStylesOnce(t *testing.T) {
	set, err := read(t, built).Entry("assets/admin.ts")

	want := `<link rel="stylesheet" href="/build/assets/vue-7.css">
<link rel="stylesheet" href="/build/assets/admin-3.css">
<link rel="modulepreload" nonce="n" href="/build/assets/vue-4.js">
<link rel="modulepreload" nonce="n" href="/build/assets/chart-6.js">
<script type="module" nonce="n" src="/build/assets/admin-5.js"></script>
`
	if got := set.HTML("n"); err != nil || string(got) != want {
		t.Errorf("got %v\n%s\nwant\n%s", err, got, want)
	}
}

func TestImportsThatImportEachOtherEndTheWalk(t *testing.T) {
	set, err := read(t, `{
		"a.ts": {"file": "assets/a.js", "isEntry": true, "imports": ["_b.js"]},
		"_b.js": {"file": "assets/b.js", "imports": ["_c.js"]},
		"_c.js": {"file": "assets/c.js", "imports": ["_b.js"]}
	}`).Entry("a.ts")

	want := `<link rel="modulepreload" nonce="n" href="/build/assets/b.js">
<link rel="modulepreload" nonce="n" href="/build/assets/c.js">
<script type="module" nonce="n" src="/build/assets/a.js"></script>
`
	if got := set.HTML("n"); err != nil || string(got) != want {
		t.Errorf("got %v\n%s\nwant\n%s", err, got, want)
	}
}

func TestOnlyAnEntryHasTags(t *testing.T) {
	for _, name := range []string{"assets/missing.ts", "_vue-4.js", "assets/img/logo.svg"} {
		if _, err := read(t, built).Entry(name); errors.Is(err, manifest.ErrUnknownEntry) == false {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestURLNamesTheBuiltFile(t *testing.T) {
	m := read(t, built)

	if url, err := m.URL("assets/img/logo.svg"); err != nil || url != "/build/assets/logo-8.svg" {
		t.Errorf("built file %q, error %v", url, err)
	}

	if url, err := m.URL("assets/admin.ts"); err != nil || url != "/build/assets/admin-5.js" {
		t.Errorf("entry %q, error %v", url, err)
	}

	if _, err := m.URL("assets/img/missing.svg"); errors.Is(err, manifest.ErrUnknownFile) == false {
		t.Errorf("missing file: %v", err)
	}
}

func TestReadNeedsAValidManifest(t *testing.T) {
	if _, err := manifest.Read(fstest.MapFS{}, "build/manifest.json", "/build/"); errors.Is(err, manifest.ErrNoManifest) == false {
		t.Errorf("without a manifest: %v", err)
	}

	broken := fstest.MapFS{"build/manifest.json": {Data: []byte("{")}}
	if _, err := manifest.Read(broken, "build/manifest.json", "/build/"); err == nil || errors.Is(err, manifest.ErrNoManifest) {
		t.Errorf("broken manifest: %v", err)
	}
}

func read(t *testing.T, data string) *manifest.Manifest {
	t.Helper()

	m, err := manifest.Read(fstest.MapFS{"build/manifest.json": {Data: []byte(data)}}, "build/manifest.json", "/build/")
	if err != nil {
		t.Fatal(err)
	}

	return m
}
