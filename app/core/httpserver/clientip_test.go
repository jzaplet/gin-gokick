package httpserver

import (
	"net/http"
	"testing"

	"gokick/app/core/testkit"

	"github.com/gin-gonic/gin"
)

func TestClientIP(t *testing.T) {
	t.Run("Cloudflare header wins", func(t *testing.T) {
		req := requestFrom(t, "198.51.100.1:4000")
		req.Header.Set("CF-Connecting-IP", "203.0.113.7")
		expectClientIP(t, req, nil, "203.0.113.7")
	})
	t.Run("X-Forwarded-For from an untrusted peer is ignored", func(t *testing.T) {
		req := requestFrom(t, "198.51.100.1:4000")
		req.Header.Set("X-Forwarded-For", "203.0.113.9")
		expectClientIP(t, req, nil, "198.51.100.1")
	})
	t.Run("X-Forwarded-For from a trusted proxy", func(t *testing.T) {
		req := requestFrom(t, "10.0.0.2:4000")
		req.Header.Set("X-Forwarded-For", "203.0.113.9")
		expectClientIP(t, req, []string{"10.0.0.0/8"}, "203.0.113.9")
	})
}

func requestFrom(t *testing.T, remote string) *http.Request {
	t.Helper()
	req := testkit.Request(t, http.MethodGet, "/ip")
	req.RemoteAddr = remote

	return req
}

func expectClientIP(t *testing.T, req *http.Request, trusted []string, want string) {
	t.Helper()

	router, err := NewEngine(gin.TestMode, trusted, testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	router.GET("/ip", func(c *gin.Context) { c.String(http.StatusOK, c.ClientIP()) })

	if got := testkit.Serve(router, req).Body.String(); got != want {
		t.Errorf("ClientIP() = %q, want %q", got, want)
	}
}
