package reporting

import (
	"errors"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

const stackKey = "reporting.stack"

const warningsKey = "reporting.warnings"

func Error(c *gin.Context, err error) {
	_ = c.Error(err)
	if _, exists := c.Get(stackKey); exists == false {
		c.Set(stackKey, sentry.NewStacktrace())
	}
}

func Warning(c *gin.Context, err error) {
	c.Set(warningsKey, append(Warnings(c), err))

	if hub := sentry.GetHubFromContext(c.Request.Context()); hub != nil {
		capture(c.Request.Context(), hub, sentry.LevelWarning, err, sentry.NewStacktrace())
	}
}

func Warnings(c *gin.Context) []error {
	value, _ := c.Get(warningsKey)
	warnings, _ := value.([]error)

	return warnings
}

func handlerError(errs []*gin.Error) error {
	if len(errs) == 1 {
		return errs[0].Err
	}

	joined := make([]error, 0, len(errs))
	for _, err := range errs {
		joined = append(joined, err.Err)
	}

	return errors.Join(joined...)
}

func reportedStack(c *gin.Context) *sentry.Stacktrace {
	stack, _ := c.Get(stackKey)
	stacktrace, _ := stack.(*sentry.Stacktrace)

	return stacktrace
}
