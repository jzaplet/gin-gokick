package reporting

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
	"gokick/app/core/reporting/inapp"
	"gokick/app/core/reporting/request"
	"gokick/app/core/reporting/trace"
)

const recoveredKey = "reporting.recovered"

const unversioned = "dev"

type Reporter struct {
	hub *sentry.Hub
}

func New(dsn, environment, release string) (*Reporter, error) {
	if dsn == "" {
		return &Reporter{}, nil
	}

	if release == "" {
		release = buildRevision()
	}

	return newReporter(&sentry.ClientOptions{Dsn: dsn, Environment: environment, Release: release})
}

func newReporter(options *sentry.ClientOptions) (*Reporter, error) {
	options.AttachStacktrace = true
	options.DisableTelemetryBuffer = true
	options.BeforeSend = beforeSend

	client, err := sentry.NewClient(*options)
	if err != nil {
		return nil, fmt.Errorf("sentry: %w", err)
	}

	return &Reporter{hub: sentry.NewHub(client, sentry.NewScope())}, nil
}

func (r *Reporter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if r.hub == nil {
			c.Next()

			return
		}

		hub := r.hub.Clone()
		hub.Scope().SetPropagationContext(trace.Propagation(c))
		hub.Scope().AddEventProcessor(request.Describe(c))
		c.Request = c.Request.WithContext(sentry.SetHubOnContext(c.Request.Context(), hub))
		c.Next()

		if c.Writer.Status() < http.StatusInternalServerError || len(c.Errors) == 0 || c.GetBool(recoveredKey) {
			return
		}

		capture(c.Request.Context(), hub, sentry.LevelError, handlerError(c.Errors), reportedStack(c))
	}
}

func (r *Reporter) Recover(c *gin.Context, recovered any) {
	c.Set(recoveredKey, true)

	ctx := c.Request.Context()
	if hub := sentry.GetHubFromContext(ctx); hub != nil {
		identify(ctx, hub)
		hub.RecoverWithContext(ctx, recovered)
	}
}

func (r *Reporter) Flush(timeout time.Duration) {
	if r.hub != nil {
		r.hub.Flush(timeout)
	}
}

func capture(ctx context.Context, hub *sentry.Hub, level sentry.Level, err error, stack *sentry.Stacktrace) {
	identify(ctx, hub)

	event := hub.Client().EventFromException(err, level)
	if sentry.ExtractStacktrace(err) == nil {
		for i := range event.Exception {
			event.Exception[i].Stacktrace = nil
		}

		if stack != nil && len(event.Exception) > 0 {
			event.Exception[len(event.Exception)-1].Stacktrace = stack
		}
	}

	hub.CaptureEvent(event)
}

func identify(ctx context.Context, hub *sentry.Hub) {
	if id := logging.UserID(ctx); id != "" {
		hub.Scope().SetUser(sentry.User{ID: id})
	}
}

func beforeSend(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	inapp.Mark(event)

	return request.Scrub(event)
}

func buildRevision() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" && setting.Value != "" {
				return setting.Value
			}
		}
	}

	return unversioned
}
