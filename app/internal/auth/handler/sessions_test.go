package handler_test

import (
	"encoding/json"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/testkit"
	"gokick/app/internal/auth/session"
	"gokick/app/internal/shared/routertest"
)

const eva = `{"email":"eva@example.com","password":"correct horse battery","locale":"cs_CZ"}`

func TestSessionsListsTheLiveSessionsOfTheUserWithTheAddressOfTheirLastRequest(t *testing.T) {
	engine, _ := routertest.New(t)
	laptop := signIn(t, engine, "/api/auth/register", jan, "Firefox", "198.51.100.4")
	signIn(t, engine, "/api/auth/login", jan, "Safari", "203.0.113.7")
	signIn(t, engine, "/api/auth/register", eva, "Chrome", "")

	res := listSessions(t, engine, laptop, "?sortBy=createdAt&sortDir=asc")

	var list struct {
		Items []map[string]any `json:"items"`

		Total int `json:"total"`

		Selectable int `json:"selectable"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &list); err != nil || res.Code != http.StatusOK || list.Total != 2 || list.Selectable != 1 || len(list.Items) != 2 {
		t.Fatalf("got %d %s", res.Code, res.Body.String())
	}

	for i, want := range []struct {
		userAgent, ip string

		current bool
	}{{"Firefox", "192.0.2.1", true}, {"Safari", "203.0.113.7", false}} {
		item := list.Items[i]

		keys := slices.Sorted(maps.Keys(item))
		if item["userAgent"] != want.userAgent || item["lastSeenIp"] != want.ip || item["current"] != want.current || strings.Join(keys, " ") != "createdAt current id lastSeenAt lastSeenIp userAgent" {
			t.Errorf("item %d: %v", i, item)
		}
	}
}

func TestSessionsFiltersByAPartOfTheAddressOfTheLastRequest(t *testing.T) {
	engine, _ := routertest.New(t)
	laptop := signIn(t, engine, "/api/auth/register", jan, "Firefox", "198.51.100.4")
	signIn(t, engine, "/api/auth/login", jan, "Safari", "203.0.113.7")
	signIn(t, engine, "/api/auth/login", jan, "Chrome", "2001:db8::1")

	for query, want := range map[string]string{
		"?ip=203.0": `"total":1,"selectable":1}`,

		"?ip=192.0.2": `"total":1,"selectable":0}`,
	} {
		res := listSessions(t, engine, laptop, query)

		if res.Code != http.StatusOK || strings.HasSuffix(res.Body.String(), want) == false {
			t.Errorf("%s: got %d %s", query, res.Code, res.Body.String())
		}
	}

	if res := listSessions(t, engine, laptop, "?ip=203.0"); strings.Contains(res.Body.String(), `"userAgent":"Safari"`) == false {
		t.Errorf("got %s", res.Body.String())
	}
}

func TestSessionsValidatesTheQuery(t *testing.T) {
	engine, _ := routertest.New(t)
	laptop := signIn(t, engine, "/api/auth/register", jan, "Firefox", "")

	for _, tc := range []struct {
		query string

		status int

		want string
	}{
		{"?sortBy=email", http.StatusUnprocessableEntity, `{"sortBy":{"key":"validation.one_of","params":{"values":"lastSeenAt, createdAt"}}}`},
		{"?ip=" + strings.Repeat("1", 46), http.StatusUnprocessableEntity, `{"ip":{"key":"validation.max_length","params":{"max":45}}}`},
	} {
		if res := listSessions(t, engine, laptop, tc.query); res.Code != tc.status || res.Body.String() != tc.want {
			t.Errorf("%s: got %d %s", tc.query, res.Code, res.Body.String())
		}
	}
}

func signIn(t *testing.T, engine *gin.Engine, path, body, userAgent, ip string) *http.Cookie {
	t.Helper()
	req := testkit.JSONRequest(t, http.MethodPost, path, body)
	req.Header.Set("User-Agent", userAgent)

	if ip != "" {
		req.Header.Set("CF-Connecting-IP", ip)
	}

	return testkit.Cookie(t, testkit.Serve(engine, req), session.CookieName)
}

func listSessions(t *testing.T, engine *gin.Engine, cookie *http.Cookie, query string) *httptest.ResponseRecorder {
	t.Helper()
	req := testkit.Request(t, http.MethodGet, "/api/auth/sessions"+query)
	req.AddCookie(cookie)

	return testkit.Serve(engine, req)
}
