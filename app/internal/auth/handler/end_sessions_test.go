package handler_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/testkit"
	"gokick/app/internal/shared/routertest"
)

const sessionOfCookie = "SELECT count(*) FROM sessions WHERE token_hash = sha256(convert_to($1, 'UTF8'))"

func TestEndSessionsEndsTheChosenSessionsOfTheUserButNeverTheCurrentOne(t *testing.T) {
	engine, pool := routertest.New(t)
	laptop := signIn(t, engine, "/api/auth/register", jan, "Firefox", "")
	phone := signIn(t, engine, "/api/auth/login", jan, "Safari", "")
	tablet := signIn(t, engine, "/api/auth/login", jan, "Chrome", "")
	evas := signIn(t, engine, "/api/auth/register", eva, "Edge", "")
	body := `{"ids":["` + sessionID(t, pool, laptop) + `","` + sessionID(t, pool, phone) + `","` + sessionID(t, pool, evas) + `"],"all":false}`

	res := endSessions(t, engine, laptop, body)

	if res.Code != http.StatusOK || res.Body.String() != `{"ended":1}` || live(t, pool, laptop, phone, tablet, evas) != "1011" {
		t.Errorf("got %d %s, live %s", res.Code, res.Body.String(), live(t, pool, laptop, phone, tablet, evas))
	}
}

func TestEndSessionsEndsAllOtherSessionsTheFilterLists(t *testing.T) {
	for name, tc := range map[string]struct {
		ip, ended, live string
	}{
		"without a filter": {"", `{"ended":2}`, "1001"},

		"with a filter": {"203.0", `{"ended":1}`, "1011"},
	} {
		t.Run(name, func(t *testing.T) {
			engine, pool := routertest.New(t)
			laptop := signIn(t, engine, "/api/auth/register", jan, "Firefox", "203.0.113.7")
			phone := signIn(t, engine, "/api/auth/login", jan, "Safari", "203.0.113.9")
			tablet := signIn(t, engine, "/api/auth/login", jan, "Chrome", "198.51.100.4")
			evas := signIn(t, engine, "/api/auth/register", eva, "Edge", "203.0.113.5")

			res := endSessions(t, engine, laptop, `{"ids":[],"all":true,"ip":"`+tc.ip+`"}`)

			if res.Code != http.StatusOK || res.Body.String() != tc.ended || live(t, pool, laptop, phone, tablet, evas) != tc.live {
				t.Errorf("got %d %s, live %s", res.Code, res.Body.String(), live(t, pool, laptop, phone, tablet, evas))
			}
		})
	}
}

func TestEndSessionsRefusesAFilterLongerThanAnAddress(t *testing.T) {
	engine, pool := routertest.New(t)
	laptop := signIn(t, engine, "/api/auth/register", jan, "Firefox", "")
	phone := signIn(t, engine, "/api/auth/login", jan, "Safari", "")

	res := endSessions(t, engine, laptop, `{"ids":[],"all":true,"ip":"`+strings.Repeat("1", 46)+`"}`)

	if res.Code != http.StatusUnprocessableEntity || res.Body.String() != `{"ip":{"key":"validation.max_length","params":{"max":45}}}` || live(t, pool, laptop, phone) != "11" {
		t.Errorf("got %d %s, live %s", res.Code, res.Body.String(), live(t, pool, laptop, phone))
	}
}

func endSessions(t *testing.T, engine http.Handler, cookie *http.Cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := testkit.JSONRequest(t, http.MethodPost, "/api/auth/sessions/end", body)
	req.AddCookie(cookie)

	return testkit.Serve(engine, req)
}

func sessionID(t *testing.T, pool *pgxpool.Pool, cookie *http.Cookie) string {
	t.Helper()

	var id string
	if err := pool.QueryRow(t.Context(), "SELECT id FROM sessions WHERE token_hash = sha256(convert_to($1, 'UTF8'))", cookie.Value).Scan(&id); err != nil {
		t.Fatal(err)
	}

	return id
}

func live(t *testing.T, pool *pgxpool.Pool, cookies ...*http.Cookie) string {
	t.Helper()

	counts := make([]string, len(cookies))
	for i, cookie := range cookies {
		counts[i] = strconv.Itoa(testkit.Count(t, pool, sessionOfCookie, cookie.Value))
	}

	return strings.Join(counts, "")
}
