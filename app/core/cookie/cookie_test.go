package cookie_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"gokick/app/core/cookie"
)

func TestSetWritesAHostOnlyCookieTheScriptCannotRead(t *testing.T) {
	res := httptest.NewRecorder()

	cookie.Set(res, "__Host-session", "token", time.Hour)

	want := "__Host-session=token; Path=/; Max-Age=3600; HttpOnly; Secure; SameSite=Lax"
	if got := res.Header().Values("Set-Cookie"); len(got) != 1 || got[0] != want {
		t.Errorf("got %q, want %s", got, want)
	}
}

func TestClearDeletesTheCookie(t *testing.T) {
	res := httptest.NewRecorder()

	cookie.Clear(res, "__Host-session")

	want := "__Host-session=; Path=/; Max-Age=0; HttpOnly; Secure; SameSite=Lax"
	if got := res.Header().Values("Set-Cookie"); len(got) != 1 || got[0] != want {
		t.Errorf("got %q, want %s", got, want)
	}
}
