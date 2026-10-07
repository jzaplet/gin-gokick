package request

import (
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

const filtered = "[Filtered]"

const unmatchedRoute = "unmatched"

func Describe(c *gin.Context) sentry.EventProcessor {
	route := c.FullPath()
	if route == "" {
		route = unmatchedRoute
	}

	request := &sentry.Request{
		Method: c.Request.Method,

		URL: route,

		QueryString: c.Request.URL.RawQuery,

		Headers: safeHeaders(c.Request.Header),
	}
	transaction := c.Request.Method + " " + route

	return func(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
		event.Request = request
		event.Transaction = transaction

		return event
	}
}

func Scrub(event *sentry.Event) *sentry.Event {
	if event.Request == nil {
		return event
	}

	event.Request.Cookies = ""
	event.Request.Data = ""

	event.Request.Env = nil
	for name := range event.Request.Headers {
		if http.CanonicalHeaderKey(name) != "User-Agent" {
			event.Request.Headers[name] = filtered
		}
	}

	return event
}

func safeHeaders(header http.Header) map[string]string {
	headers := map[string]string{}

	for _, name := range []string{"User-Agent", "Authorization", "Cookie"} {
		if value := header.Get(name); value != "" {
			headers[name] = value
		}
	}

	return headers
}
