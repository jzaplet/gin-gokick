package middleware

import (
	"net/http"
	"strings"
	"testing"

	"gokick/app/core/testkit"

	"github.com/gin-gonic/gin"
)

func TestCSRF(t *testing.T) {
	t.Run("refuses a cross-site POST", func(t *testing.T) { expectCSRF(t, http.MethodPost, "cross-site", "", http.StatusForbidden) })
	t.Run("refuses a DELETE from a sibling subdomain", func(t *testing.T) { expectCSRF(t, http.MethodDelete, "same-site", "", http.StatusForbidden) })
	t.Run("refuses a foreign Origin without Sec-Fetch-Site", func(t *testing.T) { expectCSRF(t, http.MethodPut, "", "https://evil.example", http.StatusForbidden) })
	t.Run("lets a same-origin POST through", func(t *testing.T) { expectCSRF(t, http.MethodPost, "same-origin", "", http.StatusNoContent) })
	t.Run("lets a user-initiated POST through", func(t *testing.T) { expectCSRF(t, http.MethodPost, "none", "", http.StatusNoContent) })
	t.Run("lets a matching Origin without Sec-Fetch-Site through", func(t *testing.T) { expectCSRF(t, http.MethodPatch, "", "http://example.com", http.StatusNoContent) })
	t.Run("lets a request without browser headers through", func(t *testing.T) { expectCSRF(t, http.MethodPost, "", "", http.StatusNoContent) })
	t.Run("lets a cross-site GET through", func(t *testing.T) { expectCSRF(t, http.MethodGet, "cross-site", "", http.StatusNoContent) })
	t.Run("lets a cross-site OPTIONS through", func(t *testing.T) { expectCSRF(t, http.MethodOptions, "cross-site", "", http.StatusNoContent) })
}

func TestCSRFLogsWhyItRefused(t *testing.T) {
	logger, buf := testkit.CaptureLogs()
	router := newTestRouter(t, logger)
	router.Use(CSRF())
	router.POST("/form", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := testkit.Request(t, http.MethodGet, "/form")
	req.Method = http.MethodPost
	req.Header.Set("Sec-Fetch-Site", "cross-site")

	testkit.Serve(router, req)

	entry := testkit.LogEntries(t, buf)[0]
	if message, _ := entry["error"].(string); entry["status"] != 403.0 || entry["level"] != "INFO" || strings.Contains(message, "cross-origin request") == false {
		t.Errorf("got %v", entry)
	}
}

func expectCSRF(t *testing.T, method, secFetchSite, origin string, want int) {
	t.Helper()
	router := newTestRouter(t, testkit.DiscardLogs())
	router.Use(CSRF())
	router.Any("/form", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	req := testkit.Request(t, http.MethodGet, "/form")

	req.Method = method
	if secFetchSite != "" {
		req.Header.Set("Sec-Fetch-Site", secFetchSite)
	}

	if origin != "" {
		req.Header.Set("Origin", origin)
	}

	if res := testkit.Serve(router, req); res.Code != want {
		t.Errorf("status %d, want %d", res.Code, want)
	}
}
