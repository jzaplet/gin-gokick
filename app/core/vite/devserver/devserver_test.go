package devserver_test

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/csp"
	"gokick/app/core/middleware"
	"gokick/app/core/testkit"
	"gokick/app/core/vite/devserver"
)

func TestTheDevServerGetsTheRequestPastTheRequestTimeout(t *testing.T) {
	devServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.Header().Set("X-Dev-Server", r.Host+r.URL.RequestURI())
	}))
	t.Cleanup(devServer.Close)
	app := httptest.NewServer(serve(newServer(t, devServer.URL)))
	t.Cleanup(app.Close)

	res := get(t, app.URL+"/build/@vite/client?t=1")
	if want := strings.TrimPrefix(app.URL, "http://") + "/build/@vite/client?t=1"; res.code != http.StatusOK || res.header.Get("X-Dev-Server") != want {
		t.Errorf("%d %v, want %s", res.code, res.header, want)
	}
}

func TestTheDevProxyServesOnlyLocalClients(t *testing.T) {
	engine := serve(newServer(t, "http://127.0.0.1:1"))
	req := testkit.Request(t, http.MethodGet, "/build/@vite/client")
	req.Header.Set("CF-Connecting-IP", "127.0.0.1")

	if res := testkit.Serve(engine, req); res.Code != http.StatusForbidden {
		t.Errorf("remote client %s: %d", req.RemoteAddr, res.Code)
	}
}

func TestOnlyImagesOfTheDevServerLoadFromOtherOrigins(t *testing.T) {
	devServer := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	t.Cleanup(devServer.Close)
	app := httptest.NewServer(serve(newServer(t, devServer.URL)))
	t.Cleanup(app.Close)

	for path, want := range map[string]string{
		"/build/assets/img/mail/mark.png": "cross-origin",

		"/build/assets/app.ts": "same-origin",
	} {
		res := get(t, app.URL+path)
		if got := res.header.Values("Cross-Origin-Resource-Policy"); res.code != http.StatusOK || slices.Equal(got, []string{want}) == false {
			t.Errorf("%s: %d, policy %q", path, res.code, got)
		}
	}
}

func newServer(t *testing.T, raw string) *devserver.Server {
	t.Helper()

	return devserver.New(target(t, raw), "/build/", testkit.DiscardLogs())
}

func target(t *testing.T, raw string) *url.URL {
	t.Helper()

	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}

	return parsed
}

type response struct {
	code int

	header http.Header

	body string
}

func get(t *testing.T, address string) response {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, address, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	if err = errors.Join(err, res.Body.Close()); err != nil {
		t.Fatal(err)
	}

	return response{code: res.StatusCode, header: res.Header, body: string(body)}
}

func serve(server *devserver.Server) *gin.Engine {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(middleware.SecurityHeaders(csp.Policy{}), middleware.RequestTimeout(10*time.Millisecond))
	server.Register(engine)

	return engine
}
