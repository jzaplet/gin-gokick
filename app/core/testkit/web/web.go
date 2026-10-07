package web

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var spaces = regexp.MustCompile(`[ \t\n\r\f]+`)

var tagEnds = regexp.MustCompile(` (/?>)`)

var nonce = regexp.MustCompile(`'nonce-([^']+)'`)

func Request(tb testing.TB, method, path string) *http.Request {
	tb.Helper()

	return httptest.NewRequestWithContext(tb.Context(), method, path, http.NoBody)
}

func JSONRequest(tb testing.TB, method, path, body string) *http.Request {
	tb.Helper()
	req := httptest.NewRequestWithContext(tb.Context(), method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	return req
}

func Cookie(tb testing.TB, res *httptest.ResponseRecorder, name string) *http.Cookie {
	tb.Helper()

	for _, header := range res.Header().Values("Set-Cookie") {
		if cookie, err := http.ParseSetCookie(header); err == nil && cookie.Name == name {
			return cookie
		}
	}

	tb.Fatalf("no cookie %s in Set-Cookie %q", name, res.Header().Values("Set-Cookie"))

	return nil
}

func Nonce(tb testing.TB, res *httptest.ResponseRecorder) string {
	tb.Helper()

	match := nonce.FindStringSubmatch(res.Header().Get("Content-Security-Policy"))
	if len(match) != 2 {
		tb.Fatalf("no nonce in Content-Security-Policy %q", res.Header().Get("Content-Security-Policy"))
	}

	return match[1]
}

func Serve(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	return res
}

func Markup(res *httptest.ResponseRecorder) string {
	return MarkupOf(res.Body.String())
}

func MarkupOf(html string) string {
	return tagEnds.ReplaceAllString(spaces.ReplaceAllString(html, " "), "$1")
}
