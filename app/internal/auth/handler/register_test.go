package handler_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"gokick/app/core/testkit"
	"gokick/app/internal/auth/session"
	"gokick/app/internal/shared/routertest"
	"gokick/migrations"
)

func TestRegisterStoresTheLocaleOfThePageOrTheNearestOne(t *testing.T) {
	engine, pool := routertest.WithLocales(t, "cs_CZ", "cs_CZ", "en_US")

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", strings.Replace(jan, "cs_CZ", "en_GB", 1)))

	if body := `{"id":"` + janID(t, pool) + `","locale":"en_US"}`; res.Code != http.StatusCreated || res.Body.String() != body {
		t.Errorf("got %d %s, want %s", res.Code, res.Body.String(), body)
	}

	if n := testkit.Count(t, pool, "SELECT count(*) FROM users WHERE locale = 'en_US'"); n != 1 {
		t.Errorf("%d users with the locale en_US", n)
	}
}

func TestRegisterSendsTheWelcomeMailInTheLanguageOfTheAccount(t *testing.T) {
	engine, _, mailbox := routertest.WithMailbox(t, "cs_CZ", "cs_CZ", "en_US")

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", strings.Replace(jan, "cs_CZ", "en_GB", 1)))

	sent := mailbox.Sent()
	if res.Code != http.StatusCreated || len(sent) != 1 || sent[0].To != "jan@example.com" || strings.Contains(sent[0].HTML, `href="https://gokick.dev/en/app/dashboard"`) == false {
		t.Errorf("got %d, sent %+v", res.Code, sent)
	}
}

func TestRegisterSucceedsWhenTheWelcomeMailFails(t *testing.T) {
	logger, logs := testkit.CaptureLogs()
	mailbox := testkit.Mailbox()
	mailbox.Err = errors.New("smtp down")
	pool := testkit.Pool(t, migrations.FS)
	engine := routertest.Build(t, &routertest.Parts{Logger: logger, Pool: pool, Sender: mailbox})

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan))

	assertSessionStarted(t, pool, res, http.StatusCreated)

	entry := accessLog(t, logs, "/api/auth/register")
	if warning, _ := entry["warning"].(string); entry["level"] != "WARN" || strings.HasSuffix(warning, ": smtp down") == false {
		t.Errorf("access log %v", entry)
	}
}

func TestRegisterEndsThePreviousSessionOfTheBrowser(t *testing.T) {
	engine, pool := routertest.New(t)
	eva := strings.Replace(jan, "jan", "eva", 1)
	previous := testkit.Cookie(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", eva)), session.CookieName)
	req := testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan)
	req.AddCookie(previous)

	res := testkit.Serve(engine, req)

	assertSessionStarted(t, pool, res, http.StatusCreated)

	if n := testkit.Count(t, pool, sessionsOf, previous.Value, "eva@example.com"); n != 0 {
		t.Error("the previous session of the browser is still valid")
	}
}

func TestRegisterKeepsNoAccountWhenTheSessionCannotStart(t *testing.T) {
	engine, pool := routertest.New(t)
	if _, err := pool.Exec(t.Context(), blockSessions); err != nil {
		t.Fatal(err)
	}

	assertInternal(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan)))

	if n := testkit.Count(t, pool, "SELECT count(*) FROM users"); n != 0 {
		t.Errorf("%d accounts without a session", n)
	}
}

func TestRegisterDeletesTheAccountAfterTheRequestTimedOut(t *testing.T) {
	engine, pool := routertest.New(t)

	slowSessions := `
		CREATE FUNCTION slow_sessions() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN PERFORM pg_sleep(3); RETURN NEW; END $$;
		CREATE TRIGGER slow_sessions BEFORE INSERT ON sessions FOR EACH ROW EXECUTE FUNCTION slow_sessions();`
	if _, err := pool.Exec(t.Context(), slowSessions); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()

	assertInternal(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan).WithContext(ctx)))

	if n := testkit.Count(t, pool, "SELECT count(*) FROM users"); n != 0 {
		t.Errorf("%d accounts without a session", n)
	}
}

func TestRegisterReportsBothErrorsWhenTheAccountCannotBeDeleted(t *testing.T) {
	engine, pool, logs := routertest.WithLogs(t)

	refuseDeletes := `
		CREATE FUNCTION refuse_deletes() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'deletes are refused'; END $$;
		CREATE TRIGGER refuse_deletes BEFORE DELETE ON users FOR EACH ROW EXECUTE FUNCTION refuse_deletes();`
	for _, setup := range []string{blockSessions, refuseDeletes} {
		if _, err := pool.Exec(t.Context(), setup); err != nil {
			t.Fatal(err)
		}
	}

	assertInternal(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan)))
	entry := accessLog(t, logs, "/api/auth/register")

	errs, _ := entry["error"].(string)
	if strings.HasPrefix(errs, "create a session: ") == false || strings.Contains(errs, "; delete the user: ") == false || testkit.Count(t, pool, "SELECT count(*) FROM users") != 1 {
		t.Errorf("access log %v", entry)
	}
}

func TestRegisterAnswersInTheEnvelopeWhenTheFailureCannotBeRecorded(t *testing.T) {
	engine, pool := routertest.New(t)
	testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan))

	if _, err := pool.Exec(t.Context(), "ALTER TABLE auth_failures ADD CONSTRAINT blocked CHECK (false) NOT VALID"); err != nil {
		t.Fatal(err)
	}

	assertInternal(t, testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan)))
}

func TestRegisterRefusesATakenEmail(t *testing.T) {
	engine, pool := routertest.New(t)
	testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan))

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", strings.Replace(jan, "jan", "JAN", 1)))

	if res.Code != http.StatusUnprocessableEntity || res.Body.String() != `{"email":{"key":"user.email_taken"}}` || len(res.Header().Values("Set-Cookie")) != 0 {
		t.Errorf("got %d %s %v", res.Code, res.Body.String(), res.Header())
	}

	if n := testkit.Count(t, pool, "SELECT count(*) FROM auth_failures WHERE scope = 'auth' AND network = '192.0.2.1/32'"); n != 1 {
		t.Errorf("%d failures", n)
	}
}

func TestRegisterValidatesTheRequest(t *testing.T) {
	engine, pool := routertest.New(t)

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", `{"email":"jan","password":"short"}`))

	if res.Code != http.StatusUnprocessableEntity || res.Body.String() != `{"email":{"key":"validation.email"},"locale":{"key":"validation.required"},"password":{"key":"validation.min_length","params":{"min":12}}}` {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}

	if n := testkit.Count(t, pool, "SELECT count(*) FROM auth_failures"); n != 0 {
		t.Errorf("%d failures", n)
	}
}

func TestRegisterIsLimited(t *testing.T) {
	engine, pool := routertest.New(t)
	if _, err := pool.Exec(t.Context(), tenFailures); err != nil {
		t.Fatal(err)
	}

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/api/auth/register", jan))

	if res.Code != http.StatusTooManyRequests || testkit.Count(t, pool, "SELECT count(*) FROM users") != 0 {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}
}
