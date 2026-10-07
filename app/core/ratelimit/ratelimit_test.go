package ratelimit_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/api"
	"gokick/app/core/httpserver"
	"gokick/app/core/ratelimit"
	"gokick/app/core/testkit"
)

const window = 15 * time.Minute

var errStore = errors.New("store down")

type store struct {
	failures map[netip.Prefix]int64

	recorded []netip.Prefix

	windows []time.Duration

	countErr error

	recordErr error
}

func (s *store) Count(_ context.Context, network netip.Prefix, window time.Duration) (int64, error) {
	s.windows = append(s.windows, window)

	return s.failures[network], s.countErr
}

func (s *store) Record(_ context.Context, network netip.Prefix, window time.Duration) error {
	s.windows = append(s.windows, window)
	if s.recordErr != nil {
		return s.recordErr
	}

	s.recorded = append(s.recorded, network)

	return nil
}

func TestNetworkKeysIPv4ByAddressAndIPv6By64(t *testing.T) {
	for ip, want := range map[string]string{
		"203.0.113.7": "203.0.113.7/32",

		"::ffff:203.0.113.7": "203.0.113.7/32",

		"2001:db8:1:2::abcd": "2001:db8:1:2::/64",

		"fe80::1%eth0": "fe80::/64",

		"not an address": "0.0.0.0/0",
	} {
		if got := ratelimit.Network(ip).String(); got != want {
			t.Errorf("%s: got %s, want %s", ip, got, want)
		}
	}
}

func TestMiddlewareLetsTheRequestThroughBelowTheLimit(t *testing.T) {
	failures := &store{failures: map[netip.Prefix]int64{netip.MustParsePrefix("203.0.113.7/32"): 9}}

	res := serve(t, failures, "203.0.113.7", func(c *gin.Context) { c.String(http.StatusOK, "handled") })

	if res.Code != http.StatusOK || res.Body.String() != "handled" || len(failures.windows) != 1 || failures.windows[0] != window {
		t.Errorf("got %d %s, windows %v", res.Code, res.Body.String(), failures.windows)
	}
}

func TestMiddlewareRefusesTheRequestAtTheLimit(t *testing.T) {
	failures := &store{failures: map[netip.Prefix]int64{netip.MustParsePrefix("2001:db8:1:2::/64"): 10}}

	res := serve(t, failures, "2001:db8:1:2::abcd", func(c *gin.Context) { c.String(http.StatusOK, "handled") })

	if res.Code != http.StatusTooManyRequests || res.Body.String() != `{"general":{"key":"request.too_many_attempts"}}` {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}
}

func TestMiddlewareAnswers500WhenTheStoreFails(t *testing.T) {
	res := serve(t, &store{countErr: errStore}, "203.0.113.7", func(c *gin.Context) { c.String(http.StatusOK, "handled") })

	if res.Code != http.StatusInternalServerError || res.Body.String() != `{"general":{"key":"request.internal"}}` {
		t.Errorf("got %d %s", res.Code, res.Body.String())
	}
}

func TestFailedRecordsTheNetworkOfTheRequest(t *testing.T) {
	failures := &store{}
	var err error

	serve(t, failures, "2001:db8:1:2::abcd", func(c *gin.Context) { err = ratelimit.Failed(c) })

	if err != nil || len(failures.recorded) != 1 || failures.recorded[0].String() != "2001:db8:1:2::/64" || failures.windows[1] != window {
		t.Errorf("error %v, recorded %v, windows %v", err, failures.recorded, failures.windows)
	}
}

func TestFailedNeedsTheMiddleware(t *testing.T) {
	engine := newEngine(t)
	var err error

	engine.GET("/attempt", func(c *gin.Context) { err = ratelimit.Failed(c) })

	testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/attempt"))

	if err == nil {
		t.Error("a failure without the middleware was accepted")
	}
}

func serve(t *testing.T, failures *store, ip string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	engine := newEngine(t)
	engine.Group("", api.InternalErrors()).GET("/attempt", ratelimit.Middleware(failures, 10, window), handler)

	req := testkit.Request(t, http.MethodGet, "/attempt")
	req.Header.Set("CF-Connecting-IP", ip)

	return testkit.Serve(engine, req)
}

func newEngine(t *testing.T) *gin.Engine {
	t.Helper()

	engine, err := httpserver.NewEngine(gin.TestMode, nil, testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	return engine
}
