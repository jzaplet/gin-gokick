package hreflang

import (
	"net/http"
	"reflect"
	"testing"

	"gokick/app/core/testkit"
)

func TestLinksThePageInEveryLanguage(t *testing.T) {
	want := []Link{
		{Lang: "en", URL: "http://example.com/en/app/sign%20in"},
		{Lang: "cs", URL: "http://example.com/cs/app/sign%20in"},
		{Lang: "x-default", URL: "http://example.com/cs/app/sign%20in"},
	}
	if got := Links(request(t), []string{"en", "cs"}, "cs", "/app/sign%20in"); reflect.DeepEqual(got, want) == false {
		t.Errorf("got %+v", got)
	}
}

func TestLinksTakeTheSchemeOfTheProxy(t *testing.T) {
	req := request(t)
	req.Host = "example.com"
	req.Header.Set("X-Forwarded-Proto", "https")

	if got := Links(req, []string{"cs", "en"}, "cs", ""); len(got) != 3 || got[1].URL != "https://example.com/en" {
		t.Errorf("got %+v", got)
	}
}

func request(t *testing.T) *http.Request {
	t.Helper()

	return testkit.Request(t, http.MethodGet, "/")
}
