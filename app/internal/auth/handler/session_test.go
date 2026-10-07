package handler_test

import (
	"net/http"
	"testing"

	"gokick/app/core/testkit"
	"gokick/app/internal/auth/session"
	"gokick/app/internal/shared/routertest"
)

func TestSessionAnswersTheUserOfTheSession(t *testing.T) {
	engine, pool := routertest.New(t)
	registered := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan))
	req := testkit.Request(t, http.MethodGet, "/api/auth/session")
	req.AddCookie(testkit.Cookie(t, registered, session.CookieName))

	res := testkit.Serve(engine, req)

	if res.Code != http.StatusOK || res.Body.String() != `{"id":"`+janID(t, pool)+`"}` {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}
}
