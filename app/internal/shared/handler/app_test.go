package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gokick/app/core/testkit"
	"gokick/app/internal/shared/routertest"
)

func TestAppServesTheShellOnEveryPathOfTheSPA(t *testing.T) {
	engine := routertest.Build(t, &routertest.Parts{})

	for _, path := range []string{"/cs/app", "/cs/app/", "/cs/app/login", "/cs/app/a/b?c=d"} {
		t.Run(path, func(t *testing.T) {
			assertShell(t, testkit.Serve(engine, testkit.Request(t, http.MethodGet, path)))
		})
	}
}

func TestAppPreloadsTheDictionaryOnceWithTheConsentBanner(t *testing.T) {
	cfg := routertest.Config()
	cfg.Tracking.GA4MeasurementID = "G-AB12CD34EF"
	engine := routertest.Build(t, &routertest.Parts{Config: cfg})
	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/cs/app/login"))

	if body := testkit.Markup(res); res.Code != http.StatusOK || strings.Contains(body, "consent-test.js") == false || strings.Count(body, "cs_CZ-test.js") != 1 {
		t.Errorf("want the consent entry and one dictionary preload: %d %s", res.Code, body)
	}
}

func assertShell(t *testing.T, res *httptest.ResponseRecorder) {
	t.Helper()

	if res.Code != http.StatusOK {
		t.Fatalf("status %d", res.Code)
	}

	nonce := testkit.Nonce(t, res)

	body := testkit.Markup(res)
	for _, want := range []string{
		`<meta property="csp-nonce" nonce="` + nonce + `">`,
		`<script type="module" nonce="` + nonce + `" src="/build/assets/app-test.js"></script>`,
		`<script type="module" nonce="` + nonce + `" src="/build/assets/constellation-test.js"></script>`,
		`<link rel="modulepreload" nonce="` + nonce + `" href="/build/assets/cs_CZ-test.js">`,
		`<div id="app">`,
	} {
		if strings.Contains(body, want) == false {
			t.Errorf("body lacks %s: %s", want, body)
		}
	}
}
