package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gokick/app/core/testkit"
	"gokick/app/internal/auth/session"
	"gokick/app/internal/shared/routertest"
)

func TestLogoutEndsTheSession(t *testing.T) {
	engine, pool := routertest.New(t)
	thisDevice := testkit.Cookie(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan)), session.CookieName)
	anotherDevice := testkit.Cookie(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/login", jan)), session.CookieName)
	req := testkit.JSONRequest(t, http.MethodPost, "/api/auth/logout", "")
	req.AddCookie(thisDevice)

	assertCleared(t, testkit.Serve(engine, req))

	if testkit.Count(t, pool, sessionsOf, thisDevice.Value, "jan@example.com") != 0 || testkit.Count(t, pool, sessionsOf, anotherDevice.Value, "jan@example.com") != 1 {
		t.Error("the logout did not end exactly its own session")
	}
}

func TestLogoutWithoutASession(t *testing.T) {
	engine, _ := routertest.New(t)

	assertCleared(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/logout", "")))
}

func assertCleared(t *testing.T, res *httptest.ResponseRecorder) {
	t.Helper()

	cleared := testkit.Cookie(t, res, session.CookieName)
	if res.Code != http.StatusNoContent || cleared.Value != "" || cleared.MaxAge >= 0 {
		t.Errorf("got %d, Set-Cookie %v", res.Code, res.Header().Values("Set-Cookie"))
	}
}
