package locale_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/locale"
	"gokick/app/core/locale/hreflang"
	"gokick/app/core/testkit"
)

var homes = []locale.Home{{Language: "cs", Path: "/"}, {Language: "en", Path: "/en"}}

func TestHomeIsInTheDefaultLocale(t *testing.T) {
	set := newSet(t, "cs_CZ", "en_US", "en_GB")

	page, res := servePage(t, set.Home(), "/", "/", "en-GB")

	want := []hreflang.Link{
		{Lang: "cs", URL: "http://example.com/"},
		{Lang: "en", URL: "http://example.com/en"},
		{Lang: "x-default", URL: "http://example.com/"},
	}
	if page.Locale.String() != "cs_CZ" || reflect.DeepEqual(page.Homes, homes) == false || page.Home != "/" || reflect.DeepEqual(page.Alternates, want) == false || res.Header().Values("Vary") != nil {
		t.Errorf("page %+v, vary %v", page, res.Header().Values("Vary"))
	}
}

func TestPrefixPicksTheDialectAndVaries(t *testing.T) {
	set := newSet(t, "cs_CZ", "en_US", "en_GB")

	page, res := servePage(t, set.Prefix(refuse), "/:lang/app/*path", "/en/app/sign%20in?ref=mail", "en-GB")

	want := []hreflang.Link{
		{Lang: "cs", URL: "http://example.com/cs/app/sign%20in"},
		{Lang: "en", URL: "http://example.com/en/app/sign%20in"},
		{Lang: "x-default", URL: "http://example.com/cs/app/sign%20in"},
	}
	if page.Locale.String() != "en_GB" || reflect.DeepEqual(page.Homes, homes) == false || page.Home != "/en" || reflect.DeepEqual(page.Alternates, want) == false || res.Header().Get("Vary") != "Accept-Language" {
		t.Errorf("page %+v, vary %v", page, res.Header().Values("Vary"))
	}
}

func TestPrefixLinksTheHomeOfEveryLanguage(t *testing.T) {
	set := newSet(t, "cs_CZ", "en_US")

	page, _ := servePage(t, set.Prefix(refuse), "/en", "/en", "")

	want := []hreflang.Link{
		{Lang: "cs", URL: "http://example.com/"},
		{Lang: "en", URL: "http://example.com/en"},
		{Lang: "x-default", URL: "http://example.com/"},
	}
	if page.Locale.String() != "en_US" || reflect.DeepEqual(page.Alternates, want) == false {
		t.Errorf("page %+v", page)
	}
}

func TestPrefixRefusesALanguageOutsideTheLocales(t *testing.T) {
	set := newSet(t, "cs_CZ", "en_US")
	router := newEngine()
	reached := false

	router.GET("/:lang/app/*path", set.Prefix(refuse), func(*gin.Context) {
		reached = true
	})

	for _, path := range []string{"/de/app/login", "/en_US/app/login", "/EN/app/login", "/%65n/app/login"} {
		res := testkit.Serve(router, testkit.Request(t, http.MethodGet, path))
		if res.Code != http.StatusNotFound || res.Body.String() != "refused" || res.Header().Get("Vary") != "" || reached {
			t.Errorf("%s: %d %v, reached %v", path, res.Code, res.Header(), reached)
		}
	}
}

func refuse(c *gin.Context) {
	c.String(http.StatusNotFound, "refused")
}

func newSet(t *testing.T, locales ...string) *locale.Set {
	t.Helper()

	set, err := locale.New("cs_CZ", locales)
	if err != nil {
		t.Fatal(err)
	}

	return set
}

func newEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)

	return gin.New()
}

func servePage(t *testing.T, middleware gin.HandlerFunc, route, path, acceptLanguage string) (locale.Page, *httptest.ResponseRecorder) {
	t.Helper()

	router := newEngine()
	var page locale.Page
	var err error

	router.GET(route, middleware, func(c *gin.Context) {
		page, err = locale.PageOf(c)
	})

	req := testkit.Request(t, http.MethodGet, path)
	req.Header.Set("Accept-Language", acceptLanguage)

	res := testkit.Serve(router, req)

	if err != nil {
		t.Fatal(err)
	}

	return page, res
}
