package middleware

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/csp"
	"gokick/app/core/testkit"
)

func TestSecurityHeadersOnEveryResponse(t *testing.T) {
	router := newTestRouter(t, testkit.DiscardLogs())
	router.Use(SecurityHeaders(csp.Policy{}))
	router.GET("/ok", func(c *gin.Context) { c.Status(http.StatusNoContent) })
	router.GET("/panic", func(*gin.Context) { panic("boom") })

	for _, path := range []string{"/ok", "/unknown", "/panic"} {
		res := testkit.Serve(router, testkit.Request(t, http.MethodGet, path))
		for name, want := range map[string]string{
			"Strict-Transport-Security": "max-age=63072000; includeSubDomains; preload",

			"Cross-Origin-Opener-Policy": "same-origin",

			"Cross-Origin-Resource-Policy": "same-origin",

			"Permissions-Policy": permissionsPolicy,

			"Referrer-Policy": "strict-origin-when-cross-origin",

			"X-Content-Type-Options": "nosniff",

			"X-Frame-Options": "DENY",
		} {
			if got := res.Header().Get(name); got != want {
				t.Errorf("%s %d: %s = %q", path, res.Code, name, got)
			}
		}

		if strings.HasPrefix(res.Header().Get("Content-Security-Policy"), "default-src 'none'; script-src 'nonce-") == false {
			t.Errorf("%s %d: CSP %q", path, res.Code, res.Header().Get("Content-Security-Policy"))
		}
	}
}

func TestSecurityHeadersGiveEachRequestItsOwnNonce(t *testing.T) {
	router := newTestRouter(t, testkit.DiscardLogs())
	router.Use(SecurityHeaders(csp.Policy{}))
	router.GET("/nonce", func(c *gin.Context) { c.String(http.StatusOK, csp.Nonce(c)) })

	nonce := regexp.MustCompile(`'nonce-([A-Z2-7]{26})'`)

	seen := map[string]bool{}

	for range 3 {
		res := testkit.Serve(router, testkit.Request(t, http.MethodGet, "/nonce"))

		match := nonce.FindStringSubmatch(res.Header().Get("Content-Security-Policy"))
		if len(match) != 2 || match[1] != res.Body.String() || seen[match[1]] {
			t.Fatalf("header %q, context %q", res.Header().Get("Content-Security-Policy"), res.Body.String())
		}

		seen[match[1]] = true
	}
}
