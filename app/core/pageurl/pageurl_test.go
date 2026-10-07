package pageurl

import (
	"crypto/tls"
	"net/http"
	"testing"

	"gokick/app/core/testkit"
)

func TestOriginTakesTheSchemeOfTheProxyOrTLS(t *testing.T) {
	proxied := testkit.Request(t, http.MethodGet, "/")
	proxied.Host = "gokick.dev"
	proxied.Header.Set("X-Forwarded-Proto", "https")

	direct := testkit.Request(t, http.MethodGet, "/")
	direct.TLS = &tls.ConnectionState{}

	if got := Origin(proxied); got != "https://gokick.dev" {
		t.Errorf("behind a proxy %s", got)
	}

	if got := Origin(direct); got != "https://example.com" {
		t.Errorf("over TLS %s", got)
	}
}

func TestOfLeavesOutTheQuery(t *testing.T) {
	if got := Of(testkit.Request(t, http.MethodGet, "/cs/app/sign%20in?utm_source=mail")); got != "http://example.com/cs/app/sign%20in" {
		t.Errorf("url %s", got)
	}
}
