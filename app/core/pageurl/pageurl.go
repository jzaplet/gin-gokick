package pageurl

import "net/http"

func Origin(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil || req.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	return scheme + "://" + req.Host
}

func Of(req *http.Request) string {
	return Origin(req) + req.URL.EscapedPath()
}
