package trace

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
	"gokick/app/core/testkit"
)

const traceID = "771a43a4192642f0b136d5159a501700"

const spanID = "b7ad6b7169203331"

func propagationFor(t *testing.T, sentryTrace, baggage string) sentry.PropagationContext {
	t.Helper()

	c, engine := gin.CreateTestContext(httptest.NewRecorder())
	engine.ContextWithFallback = true
	req := testkit.Request(t, http.MethodGet, "/")
	req.Header.Set("sentry-trace", sentryTrace)
	req.Header.Set("baggage", baggage)
	c.Request = req.WithContext(logging.WithTraceID(req.Context(), traceID))

	return Propagation(c)
}

func TestFollowsTheFrontendSpanOfTheSameTrace(t *testing.T) {
	pc := propagationFor(t, traceID+"-"+spanID+"-1", "sentry-trace_id="+traceID+",sentry-public_key=public")

	if pc.TraceID.String() != traceID || pc.ParentSpanID.String() != spanID {
		t.Errorf("trace %s, parent span %s", pc.TraceID, pc.ParentSpanID)
	}
}

func TestDropsAParentSpanFromAnotherTrace(t *testing.T) {
	pc := propagationFor(t, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-"+spanID+"-1", "")

	if pc.TraceID.String() != traceID || pc.ParentSpanID.String() == spanID {
		t.Errorf("trace %s, parent span %s", pc.TraceID, pc.ParentSpanID)
	}
}
