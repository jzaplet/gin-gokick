package health

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/reporting"
)

const healthy = "healthy"

const unhealthy = "unhealthy"

type Check struct {
	Name string

	Probe func(context.Context) error
}

type report struct {
	Status string `json:"status"`

	Checks map[string]string `json:"checks"`
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

func Handler(timeout time.Duration, checks ...Check) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		failures := make([]error, len(checks))

		var wg sync.WaitGroup
		for i, check := range checks {
			wg.Go(func() {
				if err := check.Probe(ctx); err != nil {
					failures[i] = fmt.Errorf("%s: %w", check.Name, err)
				}
			})
		}

		wg.Wait()

		result := report{Status: healthy, Checks: make(map[string]string, len(checks))}
		for i, check := range checks {
			result.Checks[check.Name] = healthy
			if failures[i] != nil {
				result.Checks[check.Name] = unhealthy
				result.Status = unhealthy
			}
		}

		if err := errors.Join(failures...); err != nil {
			reporting.Error(c, err)
			c.JSON(http.StatusServiceUnavailable, result)

			return
		}

		c.JSON(http.StatusOK, result)
	}
}
