package vite

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"

	"gokick/app/core/testkit"
	"gokick/app/core/vite/manifest"
)

const built = `{
	"assets/app.css": {"file": "assets/app-1.css", "isEntry": true},
	"assets/app.ts": {"file": "assets/app-2.js", "isEntry": true, "css": ["assets/app-3.css"], "imports": ["_vue-4.js"]},
	"_vue-4.js": {"file": "assets/vue-4.js", "css": ["assets/vue-7.css"]},
	"assets/img/logo.svg": {"file": "assets/logo-8.svg", "src": "assets/img/logo.svg"}
}`

var public = fstest.MapFS{"build/manifest.json": {Data: []byte(built)}}

func TestCheckNamesEveryEntryOutsideTheManifest(t *testing.T) {
	err := mustLoad(t, public, "").Check("assets/app.ts", "b.ts", "a.ts", "b.ts")
	if errors.Is(err, manifest.ErrUnknownEntry) == false || strings.HasSuffix(err.Error(), ": a.ts, b.ts") == false {
		t.Errorf("error %v", err)
	}

	if err := mustLoad(t, public, "").Check("assets/app.css", "assets/app.ts"); err != nil {
		t.Errorf("known entries: %v", err)
	}

	if err := mustLoad(t, fstest.MapFS{}, writeHotFile(t, "http://localhost:5173")).Check("a.ts"); err != nil {
		t.Errorf("dev server: %v", err)
	}
}

func TestCheckFilesNamesEveryFileOutsideTheManifest(t *testing.T) {
	err := mustLoad(t, public, "").CheckFiles("assets/img/logo.svg", "b.svg", "a.svg", "b.svg")
	if errors.Is(err, manifest.ErrUnknownFile) == false || strings.HasSuffix(err.Error(), ": a.svg, b.svg") == false {
		t.Errorf("error %v", err)
	}

	if err := mustLoad(t, public, "").CheckFiles("assets/img/logo.svg"); err != nil {
		t.Errorf("known file: %v", err)
	}

	if err := mustLoad(t, fstest.MapFS{}, writeHotFile(t, "http://localhost:5173")).CheckFiles("a.svg"); err != nil {
		t.Errorf("dev server: %v", err)
	}
}

func TestTheHotFileSwitchesToTheDevServer(t *testing.T) {
	assets := mustLoad(t, fstest.MapFS{}, writeHotFile(t, "http://localhost:5173\n"))

	got, err := assets.Tags("abc", "assets/app.css", "assets/app.ts")

	want := `<link rel="stylesheet" href="/build/assets/app.css">
<script type="module" nonce="abc" src="/build/@vite/client"></script>
<script type="module" nonce="abc" src="/build/assets/app.ts"></script>
`
	if err != nil || string(got) != want {
		t.Errorf("got %v\n%s\nwant\n%s", err, got, want)
	}

	if routes := routes(assets); len(routes) == 0 {
		t.Error("no dev routes with the dev server")
	}
}

func mustLoad(t *testing.T, build fstest.MapFS, hot string) *Assets {
	t.Helper()

	assets, err := load(build, hot, testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	return assets
}

func writeHotFile(t *testing.T, content string) string {
	t.Helper()

	name := filepath.Join(t.TempDir(), "hot")
	if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return name
}

func routes(assets *Assets) gin.RoutesInfo {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	assets.Register(engine)

	return engine.Routes()
}
