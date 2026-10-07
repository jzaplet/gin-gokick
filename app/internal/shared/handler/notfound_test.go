package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/testkit"
	"gokick/app/internal/shared/routertest"
)

const browserAccept = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"

func TestNotFoundShowsAPageInTheLanguageOfThePath(t *testing.T) {
	engine := bilingualRouter(t)

	for _, tc := range []struct{ path, accept, lang, home string }{
		{"/nothing", browserAccept, "cs-CZ", "/"},
		{"/en/nothing", browserAccept, "en-US", "/en"},
		{"/cs/nothing", browserAccept, "cs-CZ", "/"},
		{"/xx/app/", browserAccept, "cs-CZ", "/"},
		{"/xx/app/login", browserAccept, "cs-CZ", "/"},
		{"/missing.png", browserAccept, "cs-CZ", "/"},
		{"/nothing", "", "cs-CZ", "/"},
		{"/nothing", "*/*", "cs-CZ", "/"},
		{"/nothing", "image/avif,image/webp,*/*", "cs-CZ", "/"},
		{"/nothing", "text/plain", "cs-CZ", "/"},
	} {
		req := testkit.Request(t, http.MethodGet, tc.path)
		req.Header.Set("Accept", tc.accept)
		assertNotFoundPage(t, tc.path+" "+tc.accept, testkit.Serve(engine, req), tc.lang, tc.home)
	}

	for _, path := range []string{"/xx/app/", "/xx/app/login"} {
		if vary := testkit.Serve(engine, testkit.Request(t, http.MethodGet, path)).Header().Get("Vary"); vary != "" {
			t.Errorf("%s: vary %q", path, vary)
		}
	}
}

func TestNotFoundAnswersTheEnvelopeToWhoAcceptsJSON(t *testing.T) {
	engine := bilingualRouter(t)

	for _, accept := range []string{"application/json", "application/json, text/plain, */*"} {
		for _, req := range []*http.Request{
			testkit.Request(t, http.MethodGet, "/api/nothing"),
			testkit.Request(t, http.MethodPost, "/cs/app/login"),
			testkit.Request(t, http.MethodGet, "/xx/app/login"),
			testkit.Request(t, http.MethodGet, "/missing.png"),
		} {
			req.Header.Set("Accept", accept)

			res := testkit.Serve(engine, req)
			if res.Code != http.StatusNotFound || res.Body.String() != `{"general":{"key":"request.not_found"}}` || res.Header().Get("Cache-Control") != "no-store" || res.Header().Get("Content-Type") != "application/json; charset=utf-8" {
				t.Errorf("%s %s, %s: %d %v %s", req.Method, req.URL.Path, accept, res.Code, res.Header(), res.Body.String())
			}
		}
	}
}

func bilingualRouter(t *testing.T) *gin.Engine {
	t.Helper()

	return routertest.Build(t, &routertest.Parts{Locales: routertest.Locales(t, "cs_CZ", "cs_CZ", "en_US")})
}

func assertNotFoundPage(t *testing.T, name string, res *httptest.ResponseRecorder, lang, home string) {
	t.Helper()

	if res.Code != http.StatusNotFound || res.Header().Get("Cache-Control") != "no-store" || res.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("%s: %d %v", name, res.Code, res.Header())
	}

	body := testkit.Markup(res)
	for _, want := range []string{
		`<html lang="` + lang + `">`,
		`<meta name="robots" content="noindex">`,
		`<script type="module" nonce="` + testkit.Nonce(t, res) + `"`,
		`<a href="` + home + `"`,
	} {
		if strings.Contains(body, want) == false {
			t.Errorf("%s: body lacks %s: %s", name, want, body)
		}
	}

	for _, unwanted := range []string{"hreflang", `property="og:`, "twitter:card"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("%s: body has %s: %s", name, unwanted, body)
		}
	}
}
