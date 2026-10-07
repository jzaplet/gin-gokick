package request

import (
	"fmt"
	"testing"

	"github.com/getsentry/sentry-go"
)

func TestScrubMasksWhatElseReachesTheEvent(t *testing.T) {
	event := Scrub(&sentry.Event{Request: &sentry.Request{
		QueryString: "ref=x",

		Cookies: "a=b",

		Data: "password=x",

		Env: map[string]string{"REMOTE_ADDR": "192.0.2.1"},

		Headers: map[string]string{"X-Api-Key": "k", "User-Agent": "ua"},
	}})

	want := map[string]string{"X-Api-Key": "[Filtered]", "User-Agent": "ua"}
	if event.Request.QueryString != "ref=x" || event.Request.Cookies != "" || event.Request.Data != "" || event.Request.Env != nil || fmt.Sprint(event.Request.Headers) != fmt.Sprint(want) {
		t.Errorf("request %+v", event.Request)
	}
}
