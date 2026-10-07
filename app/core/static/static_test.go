package static

import (
	"net/http"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"

	"gokick/app/core/testkit"
)

var public = fstest.MapFS{
	"robots.txt": {Data: []byte("User-agent: *")},

	"favicon.ico": {Data: []byte("icon")},

	"build/assets/app-1.js": {Data: []byte("app")},

	"build/assets/mark-1.png": {Data: []byte("png")},

	"build/manifest.json": {Data: []byte("{}")},

	"images/logo.svg": {Data: []byte("<svg/>")},

	"docs/index.html": {Data: []byte("<p>docs</p>")},

	"public.go": {Data: []byte("package public")},

	"images/generate.go": {Data: []byte("package images")},

	".gitignore": {Data: []byte("build")},

	".well-known/secret": {Data: []byte("secret")},
}

func TestFilesRevalidateWithTheirETag(t *testing.T) {
	engine := serve(newRoot(t))

	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/images/logo.svg"))

	etag := res.Header().Get("ETag")
	if res.Code != http.StatusOK || res.Body.String() != "<svg/>" || res.Header().Get("Cache-Control") != "no-cache" || len(etag) != 34 {
		t.Fatalf("%d %s %v", res.Code, res.Body.String(), res.Header())
	}

	req := testkit.Request(t, http.MethodGet, "/images/logo.svg")
	req.Header.Set("If-None-Match", etag)

	if res := testkit.Serve(engine, req); res.Code != http.StatusNotModified || res.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("revalidation %d %v", res.Code, res.Header())
	}
}

func TestHashedFilesAreImmutable(t *testing.T) {
	engine := serve(newRoot(t))

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		res := testkit.Serve(engine, testkit.Request(t, method, "/build/assets/app-1.js"))
		if res.Code != http.StatusOK || res.Header().Get("Cache-Control") != immutable || res.Header().Get("Content-Type") != "text/javascript; charset=utf-8" {
			t.Errorf("%s: %d %v", method, res.Code, res.Header())
		}
	}

	req := testkit.Request(t, http.MethodGet, "/build/assets/app-1.js")
	req.Header.Set("Range", "bytes=0-1")

	if res := testkit.Serve(engine, req); res.Code != http.StatusPartialContent || res.Body.String() != "ap" || res.Header().Get("Cache-Control") != immutable {
		t.Errorf("range %d %s %v", res.Code, res.Body.String(), res.Header())
	}
}

func TestOnlyHashedImagesLoadFromOtherOrigins(t *testing.T) {
	engine := serve(newRoot(t))

	for path, want := range map[string]string{
		"/build/assets/mark-1.png": "cross-origin",

		"/build/assets/app-1.js": "",

		"/images/logo.svg": "",
	} {
		res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, path))
		if got := res.Header().Get("Cross-Origin-Resource-Policy"); res.Code != http.StatusOK || got != want {
			t.Errorf("%s: %d, policy %q", path, res.Code, got)
		}
	}
}

func TestAFailedPreconditionIsNotCached(t *testing.T) {
	req := testkit.Request(t, http.MethodGet, "/build/assets/app-1.js")
	req.Header.Set("If-Match", `"other"`)

	if res := testkit.Serve(serve(newRoot(t)), req); res.Code != http.StatusPreconditionFailed || res.Header().Get("Cache-Control") != "" {
		t.Errorf("%d %v", res.Code, res.Header())
	}
}

func TestAppRoutesWinOverFiles(t *testing.T) {
	engine := serve(newRoot(t))
	engine.GET("/robots.txt", func(c *gin.Context) {
		c.String(http.StatusOK, "app")
	})

	if res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/robots.txt")); res.Body.String() != "app" {
		t.Errorf("%d %s", res.Code, res.Body.String())
	}
}

func TestEveryOtherPathGoesToTheNotFoundHandler(t *testing.T) {
	engine := serve(newRoot(t))

	for _, path := range []string{"/", "/build/", "/build/assets", "/docs/", "/build/missing.js", "/build/assets/../manifest.json", "/images/../public.go", "/images/generate.go", "/public.go", "/.gitignore", "/.well-known/secret"} {
		req := testkit.Request(t, http.MethodGet, "/")

		req.URL.Path = path
		if res := testkit.Serve(engine, req); res.Code != http.StatusNotFound || res.Body.String() != "not found" || res.Header().Get("Cache-Control") != "" {
			t.Errorf("%s: %d %v %q", path, res.Code, res.Header(), res.Body.String())
		}
	}

	if res := testkit.Serve(engine, testkit.Request(t, http.MethodPost, "/robots.txt")); res.Code != http.StatusNotFound || res.Body.String() != "not found" {
		t.Errorf("POST: %d %v", res.Code, res.Header())
	}
}

func newRoot(t *testing.T) *Root {
	t.Helper()

	root, err := New(public, "build/assets/")
	if err != nil {
		t.Fatal(err)
	}

	return root
}

func serve(root *Root) *gin.Engine {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	root.Register(engine, func(c *gin.Context) {
		c.String(http.StatusNotFound, "not found")
	})

	return engine
}
