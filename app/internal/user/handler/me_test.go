package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/testkit"
	"gokick/app/internal/auth/session"
	"gokick/app/internal/shared/routertest"
)

const jan = `{"email":"jan@example.com","password":"correct horse battery","locale":"cs_CZ"}`

func TestMeAnswersTheSignedInUser(t *testing.T) {
	engine, pool := routertest.New(t)

	res := me(t, engine, register(t, engine))

	if id := janID(t, pool); res.Code != http.StatusOK || res.Body.String() != `{"id":"`+id+`","email":"jan@example.com"}` {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}
}

func TestMeLogsTheUserOfTheSessionOnly(t *testing.T) {
	engine, pool, logs := routertest.WithLogs(t)
	signedIn := register(t, engine)
	me(t, engine, signedIn)
	me(t, engine, nil)

	var users []any

	for _, entry := range testkit.LogEntries(t, logs) {
		if entry["msg"] == "http request" && entry["path"] == "/api/user/me" {
			users = append(users, entry["user_id"])
		}
	}

	if id := janID(t, pool); len(users) != 2 || users[0] != id || users[1] != nil {
		t.Errorf("logged users %v, want %s and none", users, id)
	}
}

func TestMeRefusesARequestWithoutASession(t *testing.T) {
	engine, _ := routertest.New(t)
	register(t, engine)

	assertSignedOut(t, me(t, engine, nil))
}

func TestMeAnswersInTheEnvelopeWhenALookupFails(t *testing.T) {
	for name, breakLookup := range map[string]string{
		"the session": "ALTER TABLE sessions RENAME COLUMN token_hash TO hash",

		"the user": "ALTER TABLE users RENAME COLUMN email TO address",
	} {
		t.Run(name, func(t *testing.T) {
			engine, pool := routertest.New(t)
			signedIn := register(t, engine)
			exec(t, pool, breakLookup)

			res := me(t, engine, signedIn)

			if res.Code != http.StatusInternalServerError || res.Body.String() != `{"general":{"key":"request.internal"}}` {
				t.Errorf("got %d %s", res.Code, res.Body.String())
			}
		})
	}
}

func register(t *testing.T, engine *gin.Engine) *http.Cookie {
	t.Helper()

	return testkit.Cookie(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan)), session.CookieName)
}

func me(t *testing.T, engine *gin.Engine, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()

	req := testkit.Request(t, http.MethodGet, "/api/user/me")
	if cookie != nil {
		req.AddCookie(cookie)
	}

	return testkit.Serve(engine, req)
}

func janID(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()

	var id string
	if err := pool.QueryRow(t.Context(), "SELECT id FROM users WHERE email = 'jan@example.com'").Scan(&id); err != nil {
		t.Fatal(err)
	}

	return id
}

func exec(t *testing.T, pool *pgxpool.Pool, query string) {
	t.Helper()

	if _, err := pool.Exec(t.Context(), query); err != nil {
		t.Fatal(err)
	}
}

func assertSignedOut(t *testing.T, res *httptest.ResponseRecorder) {
	t.Helper()

	if res.Code != http.StatusUnauthorized || res.Body.String() != `{"general":{"key":"auth.sign_in_required"}}` {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}
}
