package ratelimit

import (
	"context"
	"errors"
	"net/http"
	"net/netip"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/api"
	"gokick/app/core/reporting"
)

const KeyTooManyAttempts api.Key = "request.too_many_attempts"

const ipv6NetworkBits = 64

var errNoMiddleware = errors.New("the route has no ratelimit.Middleware, so the failure cannot be recorded")

type Store interface {
	Count(ctx context.Context, network netip.Prefix, window time.Duration) (int64, error)

	Record(ctx context.Context, network netip.Prefix, window time.Duration) error
}

type attemptKey struct{}

type attempt struct {
	store Store

	network netip.Prefix

	window time.Duration
}

func Middleware(store Store, limit int64, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		network := Network(c.ClientIP())

		failures, err := store.Count(c.Request.Context(), network, window)
		if err != nil {
			reporting.Error(c, err)
			c.Abort()

			return
		}

		if failures >= limit {
			api.Fail(c, http.StatusTooManyRequests, KeyTooManyAttempts)

			return
		}

		c.Set(attemptKey{}, attempt{store: store, network: network, window: window})
		c.Next()
	}
}

func Failed(c *gin.Context) error {
	value, _ := c.Get(attemptKey{})

	current, ok := value.(attempt)
	if ok == false {
		return errNoMiddleware
	}

	return current.store.Record(c.Request.Context(), current.network, current.window)
}

func Network(ip string) netip.Prefix {
	addr, err := netip.ParseAddr(ip)
	if err != nil {
		return netip.PrefixFrom(netip.IPv4Unspecified(), 0)
	}

	addr = addr.Unmap()

	bits := addr.BitLen()
	if addr.Is6() {
		bits = ipv6NetworkBits
	}

	network, _ := addr.Prefix(bits)

	return network
}
