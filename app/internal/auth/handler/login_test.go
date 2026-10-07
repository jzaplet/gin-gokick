package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/password"
	"gokick/app/core/testkit"
	"gokick/app/internal/auth/session"
	"gokick/app/internal/shared/routertest"
)

const jan = `{"email":"jan@example.com","password":"correct horse battery","locale":"cs_CZ"}`

const sessionsOf = "SELECT count(*) FROM sessions s JOIN users u ON u.id = s.user_id WHERE s.token_hash = sha256(convert_to($1, 'UTF8')) AND u.email = $2"

const tenFailures = "INSERT INTO auth_failures (scope, network) SELECT 'auth', '192.0.2.1/32' FROM generate_series(1, 10)"

const owaspHashes = "SELECT count(*) FROM users WHERE email = 'jan@example.com' AND password_hash LIKE '$argon2id$v=19$m=19456,t=2,p=1$%'"

const blockSessions = "ALTER TABLE sessions ADD CONSTRAINT blocked CHECK (false) NOT VALID"

func TestLoginGivesTheLocaleOfTheAccountOrTheNearestOne(t *testing.T) {
	engine, pool := routertest.WithLocales(t, "cs_CZ", "cs_CZ", "en_US")
	testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan))

	if _, err := pool.Exec(t.Context(), "UPDATE users SET locale = 'en_GB'"); err != nil {
		t.Fatal(err)
	}

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/login", jan))

	if want := `{"id":"` + janID(t, pool) + `","locale":"en_US"}`; res.Code != http.StatusOK || res.Body.String() != want {
		t.Errorf("got %d %s, want %s", res.Code, res.Body.String(), want)
	}
}

func TestLoginEndsThePreviousSessionOfTheBrowser(t *testing.T) {
	engine, pool := routertest.New(t)
	previous := testkit.Cookie(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan)), session.CookieName)
	req := testkit.JSONRequest(t, http.MethodPost, "/api/auth/login", jan)
	req.AddCookie(previous)

	res := testkit.Serve(engine, req)

	assertSessionStarted(t, pool, res, http.StatusOK)

	if n := testkit.Count(t, pool, sessionsOf, previous.Value, "jan@example.com"); n != 0 {
		t.Error("the previous session of the browser is still valid")
	}
}

func TestLoginRehashesAPasswordWithOtherParams(t *testing.T) {
	engine, pool := routertest.New(t)
	addCheapUser(t, pool)

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/login", jan))

	assertSessionStarted(t, pool, res, http.StatusOK)
}

func TestLoginSucceedsWhenTheRehashFailsAndLogsAWarning(t *testing.T) {
	engine, pool, logs := routertest.WithLogs(t)
	addCheapUser(t, pool)

	if _, err := pool.Exec(t.Context(), "ALTER TABLE users ADD CONSTRAINT frozen CHECK (password_hash NOT LIKE '$argon2id$v=19$m=19456,%') NOT VALID"); err != nil {
		t.Fatal(err)
	}

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/login", jan))

	issued := testkit.Cookie(t, res, session.CookieName)
	if res.Code != http.StatusOK || testkit.Count(t, pool, sessionsOf, issued.Value, "jan@example.com") != 1 || testkit.Count(t, pool, owaspHashes) != 0 {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}

	entry := accessLog(t, logs, "/api/auth/login")
	if warning, _ := entry["warning"].(string); entry["level"] != "WARN" || strings.HasPrefix(warning, "store the new hash: ") == false {
		t.Errorf("access log %v", entry)
	}
}

func TestLoginRefusesWrongCredentials(t *testing.T) {
	engine, pool := routertest.New(t)
	testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan))

	for name, body := range map[string]string{
		"a typo": `{"email":"jan@example.com","password":"correct horse battery!"}`,

		"an unknown email": `{"email":"eva@example.com","password":"correct horse battery"}`,
	} {
		req := testkit.JSONRequest(t, http.MethodPost, "/api/auth/login", body)
		req.Header.Set("CF-Connecting-IP", "2001:db8:1:2::abcd")
		res := testkit.Serve(engine, req)

		if res.Code != http.StatusUnauthorized || res.Body.String() != `{"general":{"key":"auth.login_failed"}}` || len(res.Header().Values("Set-Cookie")) != 0 {
			t.Errorf("%s: got %d %s %v", name, res.Code, res.Body.String(), res.Header())
		}
	}

	if n := testkit.Count(t, pool, "SELECT count(*) FROM auth_failures WHERE scope = 'auth' AND network = '2001:db8:1:2::/64'"); n != 2 {
		t.Errorf("%d failures", n)
	}
}

func TestLoginIsLimited(t *testing.T) {
	engine, pool := routertest.New(t)
	if _, err := pool.Exec(t.Context(), tenFailures); err != nil {
		t.Fatal(err)
	}

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/login", jan))

	if res.Code != http.StatusTooManyRequests || res.Body.String() != `{"general":{"key":"request.too_many_attempts"}}` {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}
}

func TestLoginAnswersInTheEnvelopeWhenTheDatabaseFails(t *testing.T) {
	engine, pool := routertest.New(t)
	pool.Close()

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/login", jan))

	assertInternal(t, res)
}

func TestLoginAnswersInTheEnvelopeWhenAStepFails(t *testing.T) {
	for name, tc := range map[string]struct {
		setup string

		body string
	}{
		"a broken hash": {setup: "UPDATE users SET password_hash = 'broken'", body: jan},

		"a session that cannot start": {setup: blockSessions, body: jan},

		"a failure that cannot be recorded": {
			setup: "ALTER TABLE auth_failures ADD CONSTRAINT blocked CHECK (false) NOT VALID",

			body: `{"email":"jan@example.com","password":"correct horse battery!"}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			engine, pool := routertest.New(t)
			testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan))

			if _, err := pool.Exec(t.Context(), tc.setup); err != nil {
				t.Fatal(err)
			}

			assertInternal(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/login", tc.body)))
		})
	}
}

func accessLog(t *testing.T, logs *bytes.Buffer, path string) map[string]any {
	t.Helper()

	for _, entry := range testkit.LogEntries(t, logs) {
		if entry["msg"] == "http request" && entry["path"] == path {
			return entry
		}
	}

	t.Fatalf("no access log of %s", path)

	return nil
}

func addCheapUser(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	hash, err := password.Params{Memory: 64, Time: 1, Threads: 1}.Hash("correct horse battery")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := pool.Exec(t.Context(), "INSERT INTO users (email, password_hash, locale) VALUES ('jan@example.com', $1, 'cs_CZ')", hash); err != nil {
		t.Fatal(err)
	}
}

func assertSessionStarted(t *testing.T, pool *pgxpool.Pool, res *httptest.ResponseRecorder, status int) {
	t.Helper()

	issued := testkit.Cookie(t, res, session.CookieName)
	if res.Code != status || res.Body.String() != `{"id":"`+janID(t, pool)+`","locale":"cs_CZ"}` || issued.MaxAge != int(session.Lifetime.Seconds()) {
		t.Errorf("got %d %s, Max-Age %d", res.Code, res.Body.String(), issued.MaxAge)
	}

	if n := testkit.Count(t, pool, sessionsOf, issued.Value, "jan@example.com"); n != 1 {
		t.Errorf("%d sessions for the cookie", n)
	}

	if n := testkit.Count(t, pool, owaspHashes); n != 1 {
		t.Error("the stored hash has other parameters than OWASP")
	}
}

func janID(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()

	var id string
	if err := pool.QueryRow(t.Context(), "SELECT id FROM users WHERE email = 'jan@example.com'").Scan(&id); err != nil {
		t.Fatal(err)
	}

	return id
}

func assertInternal(t *testing.T, res *httptest.ResponseRecorder) {
	t.Helper()

	if res.Code != http.StatusInternalServerError || res.Body.String() != `{"general":{"key":"request.internal"}}` || len(res.Header().Values("Set-Cookie")) != 0 {
		t.Errorf("got %d %s %v", res.Code, res.Body.String(), res.Header())
	}
}
