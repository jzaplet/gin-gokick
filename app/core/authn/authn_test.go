package authn_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gokick/app/core/api"
	"gokick/app/core/authn"
	"gokick/app/core/httpserver"
	"gokick/app/core/testkit"
)

const cookieName = "__Host-session"

var errSessions = errors.New("sessions down")

var jan = uuid.MustParse("0192f3a4-5b6c-7d8e-9f01-23456789abcd")

type sessions struct {
	users map[string]uuid.UUID

	err error
}

func (s sessions) User(_ context.Context, token, _ string) (uuid.UUID, bool, error) {
	id, ok := s.users[token]

	return id, ok, s.err
}

func TestMiddlewareRefusesARequestWithoutASession(t *testing.T) {
	live := sessions{users: map[string]uuid.UUID{"live": jan}}

	for name, cookie := range map[string]string{
		"no cookie": "",

		"an unknown token": cookieName + "=unknown",

		"the token in another cookie": "session=live",
	} {
		handled := false

		res := serve(t, live, cookie, func(*gin.Context) { handled = true })

		if res.Code != http.StatusUnauthorized || res.Body.String() != `{"general":{"key":"auth.sign_in_required"}}` || handled {
			t.Errorf("%s: got %d %s, handled %t", name, res.Code, res.Body.String(), handled)
		}
	}
}

func TestMiddlewareAnswers500WhenTheSessionsFail(t *testing.T) {
	handled := false

	res := serve(t, sessions{err: errSessions}, cookieName+"=live", func(*gin.Context) { handled = true })

	if res.Code != http.StatusInternalServerError || res.Body.String() != `{"general":{"key":"request.internal"}}` || handled {
		t.Errorf("got %d %s, handled %t", res.Code, res.Body.String(), handled)
	}
}

func TestUserIDNeedsTheMiddleware(t *testing.T) {
	engine := newEngine(t)
	var err error

	engine.GET("/me", func(c *gin.Context) { _, err = authn.UserID(c) })

	testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/me"))

	if err == nil {
		t.Error("a route without the middleware has a user")
	}
}

func serve(t *testing.T, users sessions, cookie string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	engine := newEngine(t)
	engine.Group("", api.InternalErrors()).GET("/me", authn.Middleware(users, cookieName), handler)

	req := testkit.Request(t, http.MethodGet, "/me")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	return testkit.Serve(engine, req)
}

func newEngine(t *testing.T) *gin.Engine {
	t.Helper()

	engine, err := httpserver.NewEngine(gin.TestMode, nil, testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	return engine
}
