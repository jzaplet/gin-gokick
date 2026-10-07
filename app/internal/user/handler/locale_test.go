package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/testkit"
	"gokick/app/internal/shared/routertest"
)

func TestLocaleStoresTheLocaleOrTheNearestOne(t *testing.T) {
	engine, pool := routertest.WithLocales(t, "cs_CZ", "cs_CZ", "en_US")

	res := setLocale(t, engine, register(t, engine), `{"locale":"en_GB"}`)

	if n := testkit.Count(t, pool, "SELECT count(*) FROM users WHERE locale = 'en_US'"); res.Code != http.StatusNoContent || res.Body.Len() != 0 || n != 1 {
		t.Errorf("got %d %s, %d users with en_US", res.Code, res.Body.String(), n)
	}
}

func TestLocaleWantsALocale(t *testing.T) {
	engine, _ := routertest.New(t)

	res := setLocale(t, engine, register(t, engine), `{}`)

	if res.Code != http.StatusUnprocessableEntity || res.Body.String() != `{"locale":{"key":"validation.required"}}` {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}
}

func setLocale(t *testing.T, engine *gin.Engine, cookie *http.Cookie, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := testkit.JSONRequest(t, http.MethodPut, "/api/user/locale", body)
	req.AddCookie(cookie)

	return testkit.Serve(engine, req)
}
