package cookie

import (
	"net/http"
	"time"
)

func Set(w http.ResponseWriter, name, value string, maxAge time.Duration) {
	write(w, name, value, int(maxAge.Seconds()))
}

func Clear(w http.ResponseWriter, name string) {
	write(w, name, "", -1)
}

func write(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name: name,

		Value: value,

		Path: "/",

		MaxAge: maxAge,

		Secure: true,

		HttpOnly: true,

		SameSite: http.SameSiteLaxMode,
	})
}
