package router_test

import (
	"context"
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/database"
	"gokick/app/core/reporting"
	"gokick/app/core/testkit"
	"gokick/app/internal/shared/config"
	"gokick/app/internal/shared/config/prometheus"
	"gokick/app/internal/shared/csp"
	"gokick/app/internal/shared/router"
	"gokick/app/internal/shared/routertest"
	dictionaries "gokick/locale"
	"gokick/public"
)

var titleOf = regexp.MustCompile(`<title>([^<]*)</title>`)

func TestNewRouterRunsTheCoreMiddleware(t *testing.T) {
	logger, buf := testkit.CaptureLogs()
	engine := routertest.Build(t, &routertest.Parts{Logger: logger})
	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/ping"))

	traceID := res.Header().Get("X-Trace-Id")

	entries := testkit.LogEntries(t, buf)
	if traceID == "" || len(entries) != 1 || entries[0]["msg"] != "http request" || entries[0]["trace_id"] != traceID {
		t.Errorf("trace id %q, log %v", traceID, entries)
	}
}

func TestNewRouterSendsThePolicyOfTheEnabledTools(t *testing.T) {
	cfg := routertest.Config()
	cfg.Tracking.GA4MeasurementID = "G-AB12CD34EF"
	engine := routertest.Build(t, &routertest.Parts{Config: cfg})
	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/ping"))

	policy := res.Header().Get("Content-Security-Policy")
	if want := csp.Policy(&cfg.Tracking).Header().Value(testkit.Nonce(t, res)); policy != want || strings.Contains(policy, "https://www.googletagmanager.com") == false {
		t.Errorf("got  %s\nwant %s", policy, want)
	}
}

func TestNewRouterServesThePublicDirectory(t *testing.T) {
	webRoot := routertest.Public(t, routertest.Locales(t, "cs_CZ", "cs_CZ"))
	for _, name := range []string{"public.go", ".hot", "robots.txt", "ping"} {
		webRoot[name] = &fstest.MapFile{Data: []byte(name)}
	}

	engine := routertest.Build(t, &routertest.Parts{Public: webRoot})

	for path, want := range map[string]string{
		"/build/assets/app-test.js": "public, max-age=31536000, immutable",

		"/robots.txt": "no-cache",
	} {
		if res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, path)); res.Code != http.StatusOK || res.Header().Get("Cache-Control") != want {
			t.Errorf("%s: %d %v", path, res.Code, res.Header())
		}
	}

	for _, path := range []string{"/public.go", "/.hot", "/build/assets/missing.js", "/missing.png"} {
		if res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, path)); res.Code != http.StatusNotFound || res.Header().Get("Cache-Control") != "no-store" {
			t.Errorf("%s: %d %v", path, res.Code, res.Header())
		}
	}

	if res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/ping")); strings.Contains(res.Body.String(), "pong") == false {
		t.Errorf("a public file took the route: %s", res.Body.String())
	}
}

func TestNewRouterGivesTheBrowserItsSentryOptions(t *testing.T) {
	browser := reporting.NewBrowser("https://public@o1.ingest.sentry.io/3", "staging", "v1")

	want := `<meta name="sentry" data-dsn="https://public@o1.ingest.sentry.io/3" data-environment="staging" data-release="v1">`
	if body := homepage(t, routertest.Config(), browser); strings.Contains(body, want) == false {
		t.Errorf("body lacks %s: %s", want, body)
	}

	if body := homepage(t, routertest.Config(), nil); strings.Contains(body, `name="sentry"`) {
		t.Errorf("sentry without a DSN: %s", body)
	}
}

func TestNewRouterLoadsUmamiWithTheNonceOfThePage(t *testing.T) {
	cfg := routertest.Config()
	cfg.Tracking.UmamiWebsiteID = "0192f3a4-5b6c-7d8e-9f01-23456789abcd"
	engine := routertest.Build(t, &routertest.Parts{Config: cfg})
	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/"))

	want := `<script nonce="` + testkit.Nonce(t, res) + `" defer src="https://analytics.strategio.dev/script.js" data-website-id="0192f3a4-5b6c-7d8e-9f01-23456789abcd" data-exclude-hash="true"></script>`
	if body := testkit.Markup(res); strings.Contains(body, want) == false {
		t.Errorf("body lacks %s: %s", want, body)
	}

	if body := homepage(t, routertest.Config(), nil); strings.Contains(body, "analytics.strategio.dev") {
		t.Errorf("umami without a website: %s", body)
	}
}

func TestNewRouterEscapesTheUmamiWebsite(t *testing.T) {
	cfg := routertest.Config()
	cfg.Tracking.UmamiWebsiteID = `"><b>`

	want := `data-website-id="&#34;&gt;&lt;b&gt;"`
	if body := homepage(t, cfg, nil); strings.Contains(body, want) == false {
		t.Errorf("body lacks %s: %s", want, body)
	}
}

func TestNewRouterStartsTheConsentBannerWithItsTools(t *testing.T) {
	cfg := routertest.Config()
	cfg.Tracking.GA4MeasurementID = "G-AB12CD34EF"
	cfg.Tracking.MetaPixelID = "1234567890123456"
	engine := routertest.Build(t, &routertest.Parts{Config: cfg})
	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/"))

	nonce := testkit.Nonce(t, res)
	meta := `<meta name="tracking" data-ga4="G-AB12CD34EF" data-google-ads="" data-meta-pixel="1234567890123456" data-google-ads-conversions="">`

	body := testkit.Markup(res)
	for _, want := range []string{meta, `<script type="module" nonce="` + nonce + `" src="/build/assets/consent-test.js"></script>`} {
		if strings.Contains(body, want) == false {
			t.Errorf("body lacks %s: %s", want, body)
		}
	}

	preload := `<link rel="modulepreload" nonce="` + nonce + `" href="/build/assets/cs_CZ-test.js">`
	if strings.Count(body, preload) != 1 {
		t.Errorf("body preloads the dictionary %d times: %s", strings.Count(body, preload), body)
	}

	umami := routertest.Config()

	umami.Tracking.UmamiWebsiteID = "0192f3a4-5b6c-7d8e-9f01-23456789abcd"
	if body := homepage(t, umami, nil); strings.Contains(body, `name="tracking"`) || strings.Contains(body, "consent-test.js") || strings.Contains(body, "cs_CZ-test.js") {
		t.Errorf("consent banner without a tool that needs it: %s", body)
	}
}

func TestNewRouterEscapesTheTrackingIDs(t *testing.T) {
	cfg := routertest.Config()
	cfg.Tracking.MetaPixelID = `"><b>`

	want := `data-meta-pixel="&#34;&gt;&lt;b&gt;"`
	if body := homepage(t, cfg, nil); strings.Contains(body, want) == false {
		t.Errorf("body lacks %s: %s", want, body)
	}
}

func TestNewRouterLoadsTheBuiltAssets(t *testing.T) {
	if _, err := fs.Stat(public.FS, "build/manifest.json"); err != nil {
		if os.Getenv("CI") != "" {
			t.Fatalf("no Vite build: %v", err)
		}

		t.Skipf("no Vite build, run yarn build: %v", err)
	}

	engine := routertest.Build(t, &routertest.Parts{Public: public.FS})
	for _, path := range []string{"/", "/cs/app", "/favicon.ico"} {
		if res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, path)); res.Code != http.StatusOK {
			t.Errorf("%s: %d", path, res.Code)
		}
	}
}

func TestNewRouterServesPagesUnderTheirLanguage(t *testing.T) {
	engine := bilingual(t)

	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/cs/app"))
	if body := testkit.Markup(res); res.Code != http.StatusOK || strings.Contains(body, `<html lang="cs-CZ">`) == false {
		t.Errorf("/cs/app: %d %s", res.Code, body)
	}

	for _, path := range []string{"/app", "/app/login", "/de", "/de/app/login", "/en_US/app/login", "/cs-CZ/app", "/ping/app"} {
		if res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, path)); res.Code != http.StatusNotFound {
			t.Errorf("%s: %d", path, res.Code)
		}
	}
}

func TestEveryPageRendersInEveryLanguage(t *testing.T) {
	cfg := routertest.Config()
	cfg.Tracking.GA4MeasurementID = "G-AB12CD34EF"
	locales := routertest.Locales(t, "cs_CZ", "cs_CZ", "en_US")
	engine := routertest.Build(t, &routertest.Parts{Config: cfg, Locales: locales})

	for _, page := range []struct {
		path, locale, title string

		status int

		card, controls bool
	}{
		{"/?utm_source=mail", "cs_CZ", "home.title", http.StatusOK, true, true},
		{"/en?utm_source=mail", "en_US", "home.title", http.StatusOK, true, true},
		{"/cs/app/login?redirect=%2Fdashboard", "cs_CZ", "app.title", http.StatusOK, true, false},
		{"/en/app/login", "en_US", "app.title", http.StatusOK, true, false},
		{"/cs/nothing", "cs_CZ", "not_found.title", http.StatusNotFound, false, true},
		{"/en/nothing", "en_US", "not_found.title", http.StatusNotFound, false, true},
	} {
		t.Run(page.path, func(t *testing.T) {
			res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, page.path))
			body := testkit.Markup(res)

			title := assertPage(t, res, body, page.locale, page.title, page.status)
			if page.card {
				assertCard(t, body, page.path, page.locale, title)
			}

			if page.controls {
				assertControls(t, body, page.locale)
			}
		})
	}
}

func assertPage(t *testing.T, res *httptest.ResponseRecorder, body, locale, key string, status int) string {
	t.Helper()

	if res.Code != status || res.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("status %d, content type %q", res.Code, res.Header().Get("Content-Type"))
	}

	if strings.Contains(body, `<html lang="`+strings.ReplaceAll(locale, "_", "-")+`">`) == false {
		t.Errorf("body lacks the lang of %s: %s", locale, body)
	}

	title := titleOf.FindStringSubmatch(body)
	if len(title) != 2 || strings.Contains(title[1], template.HTMLEscapeString(dictionaries.All()[locale][key])) == false {
		t.Fatalf("title %q lacks %s of %s", title, key, locale)
	}

	return title[1]
}

func assertCard(t *testing.T, body, path, locale, title string) {
	t.Helper()

	address, _, _ := strings.Cut(path, "?")
	for _, want := range []string{
		`<link rel="alternate" hreflang="cs" href="`,
		`<link rel="alternate" hreflang="en" href="`,
		`<link rel="alternate" hreflang="x-default" href="`,
		`<meta property="og:type" content="website">`,
		`<meta property="og:locale" content="` + locale + `">`,
		`<meta property="og:url" content="http://example.com` + address + `">`,
		`<meta property="og:title" content="` + title + `">`,
		`<meta property="og:image" content="http://example.com/build/assets/` + locale + `-test.png">`,
		`<meta property="og:image:type" content="image/png">`,
		`<meta property="og:image:width" content="1200">`,
		`<meta property="og:image:height" content="630">`,
		`<meta name="twitter:card" content="summary_large_image">`,
	} {
		if strings.Contains(body, want) == false {
			t.Errorf("body lacks %s: %s", want, body)
		}
	}

	for _, property := range []string{"og:site_name", "og:description", "og:image:alt"} {
		if regexp.MustCompile(`<meta property="`+regexp.QuoteMeta(property)+`" content="[^"]+">`).MatchString(body) == false {
			t.Errorf("body lacks a value of %s: %s", property, body)
		}
	}
}

func assertControls(t *testing.T, body, locale string) {
	t.Helper()

	for each, home := range map[string]string{"cs_CZ": "/", "en_US": "/en"} {
		lang, name := each[:2], dictionaries.All()[each][dictionaries.NameKey]

		want := []string{`<li lang="` + lang + `"> <a href="` + home + `"`, `> ` + name + ` </a>`}
		if each == locale {
			want = []string{`<li lang="` + lang + `" aria-current="true"`, `> ` + name + ` </li>`}
		}

		for _, part := range want {
			if strings.Contains(body, part) == false {
				t.Errorf("the switcher lacks %s: %s", part, body)
			}
		}
	}

	if strings.Contains(body, "data-consent-settings") == false {
		t.Errorf("body lacks the cookie settings: %s", body)
	}
}

func TestNewRouterSendsTheDefaultLanguageHome(t *testing.T) {
	res := testkit.Serve(bilingual(t), testkit.Request(t, http.MethodGet, "/cs?ref=mail"))

	if res.Code != http.StatusMovedPermanently || res.Header().Get("Location") != "/?ref=mail" {
		t.Errorf("%d %s", res.Code, res.Header().Get("Location"))
	}
}

func TestNewRouterLinksThePageInEveryLanguage(t *testing.T) {
	req := testkit.Request(t, http.MethodGet, "/en/app/login?redirect=%2Fdashboard")
	req.Header.Set("Accept-Language", "en-GB")
	res := testkit.Serve(bilingual(t), req)

	body := testkit.Markup(res)
	for _, want := range []string{
		`<meta name="locale" content="en_US" data-homes="cs=/ en=/en" data-languages="cs_CZ en_US">`,
		`<link rel="alternate" hreflang="cs" href="http://example.com/cs/app/login">`,
		`<link rel="alternate" hreflang="en" href="http://example.com/en/app/login">`,
		`<link rel="alternate" hreflang="x-default" href="http://example.com/cs/app/login">`,
	} {
		if strings.Contains(body, want) == false {
			t.Errorf("body lacks %s: %s", want, body)
		}
	}

	if res.Header().Get("Vary") != "Accept-Language" {
		t.Errorf("vary %v", res.Header().Values("Vary"))
	}
}

func TestMetricsCountTheSPAOfEveryLanguageAsOneRoute(t *testing.T) {
	engine := bilingual(t)
	testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/cs/app/login"))
	testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/en/app/dashboard"))
	testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/en"))
	req := testkit.Request(t, http.MethodGet, "/metrics")
	req.SetBasicAuth("prometheus", "secret")

	body := testkit.Serve(engine, req).Body.String()
	for _, want := range []string{`route="/:lang/app/*path",status="200"} 2`, `route="/en",status="200"} 1`} {
		if strings.Contains(body, want) == false {
			t.Errorf("metrics lack %s: %s", want, body)
		}
	}
}

func TestNewRouterLimitsTheRequestTime(t *testing.T) {
	engine := routertest.Build(t, &routertest.Parts{})
	var left time.Duration

	engine.GET("/deadline", func(c *gin.Context) {
		if deadline, ok := c.Request.Context().Deadline(); ok {
			left = time.Until(deadline)
		}
	})

	testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/deadline"))

	if left <= 0 || left > router.RequestTimeout {
		t.Errorf("time left %s", left)
	}
}

func TestNewRouterRefusesCrossSiteForms(t *testing.T) {
	req := testkit.Request(t, http.MethodGet, "/ping")
	req.Method = http.MethodPost
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	if res := testkit.Serve(routertest.Build(t, &routertest.Parts{}), req); res.Code != http.StatusForbidden {
		t.Errorf("status %d", res.Code)
	}
}

func TestMetricsNeedTheAccount(t *testing.T) {
	cfg := routertest.Config()
	cfg.Prometheus = prometheus.Config{User: "prometheus", Password: "secret"}
	engine := routertest.Build(t, &routertest.Parts{Config: cfg})
	testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/ping"))

	if res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/metrics")); res.Code != http.StatusUnauthorized {
		t.Errorf("without the account: %d", res.Code)
	}

	wrong := testkit.Request(t, http.MethodGet, "/metrics")
	wrong.SetBasicAuth("prometheus", "guess")

	if res := testkit.Serve(engine, wrong); res.Code != http.StatusUnauthorized {
		t.Errorf("with a wrong password: %d", res.Code)
	}

	right := testkit.Request(t, http.MethodGet, "/metrics")
	right.SetBasicAuth("prometheus", "secret")

	res := testkit.Serve(engine, right)
	if res.Code != http.StatusOK || strings.Contains(res.Body.String(), `http_requests_total{method="GET",route="/ping",status="200"} 1`) == false {
		t.Errorf("with the account: %d %s", res.Code, res.Body.String())
	}
}

func TestMetricsAreOffWithoutAnAccount(t *testing.T) {
	req := testkit.Request(t, http.MethodGet, "/metrics")
	req.SetBasicAuth("", "")

	if res := testkit.Serve(routertest.Build(t, &routertest.Parts{}), req); res.Code != http.StatusNotFound {
		t.Errorf("status %d", res.Code)
	}
}

func TestReadinessChecksTheDatabase(t *testing.T) {
	pool, err := database.Open(context.Background(), testkit.Database(t))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)
	engine := routertest.Build(t, &routertest.Parts{Pool: pool})

	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/readyz"))
	if res.Code != http.StatusOK || res.Body.String() != `{"status":"healthy","checks":{"database":"healthy"}}` {
		t.Errorf("%d %s", res.Code, res.Body.String())
	}
}

func TestReadinessFailsWithoutTheDatabase(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), "postgres://gokick@127.0.0.1:1/gokick")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)

	logger, buf := testkit.CaptureLogs()
	engine := routertest.Build(t, &routertest.Parts{Logger: logger, Pool: pool})

	res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/readyz"))

	entries := testkit.LogEntries(t, buf)
	if res.Code != http.StatusServiceUnavailable || res.Body.String() != `{"status":"unhealthy","checks":{"database":"unhealthy"}}` {
		t.Errorf("%d %s", res.Code, res.Body.String())
	}

	if len(entries) != 1 || strings.HasPrefix(entries[0]["error"].(string), "database: ") == false {
		t.Errorf("log %v", entries)
	}
}

func bilingual(t *testing.T) *gin.Engine {
	t.Helper()

	cfg := routertest.Config()
	cfg.Prometheus = prometheus.Config{User: "prometheus", Password: "secret"}
	locales := routertest.Locales(t, "cs_CZ", "cs_CZ", "en_US")

	return routertest.Build(t, &routertest.Parts{Config: cfg, Locales: locales})
}

func homepage(t *testing.T, cfg *config.Config, browser *reporting.Browser) string {
	t.Helper()
	engine := routertest.Build(t, &routertest.Parts{Config: cfg, Browser: browser})

	return testkit.Markup(testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/")))
}
