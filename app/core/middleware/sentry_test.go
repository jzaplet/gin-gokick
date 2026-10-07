package middleware

import (
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"gokick/app/core/reporting"
	"gokick/app/core/testkit"
)

func TestAPanicReachesSentryUnderTheRequestTrace(t *testing.T) {
	reporter, ingest := newIngestReporter(t)
	router := newReportingRouter(t, testkit.DiscardLogs(), reporter)
	router.GET("/dekujeme/:token", func(*gin.Context) { panic("boom") })

	res := testkit.Serve(router, testkit.Request(t, http.MethodGet, "/dekujeme/secret-token"))

	reporter.Flush(2 * time.Second)

	sent := ingest.received()
	if res.Code != http.StatusInternalServerError || strings.Contains(sent, `"boom"`) == false {
		t.Fatalf("status %d, sent %s", res.Code, sent)
	}

	if strings.Contains(sent, `"trace_id":"`+res.Header().Get("X-Trace-Id")+`"`) == false {
		t.Errorf("the event lacks trace id %s: %s", res.Header().Get("X-Trace-Id"), sent)
	}

	if strings.Contains(sent, "secret-token") {
		t.Errorf("the token from the path reached Sentry: %s", sent)
	}
}

func TestAHandlerErrorReachesSentryWithTheHandlerAsCulprit(t *testing.T) {
	reporter, ingest := newIngestReporter(t)
	router := newReportingRouter(t, testkit.DiscardLogs(), reporter)
	router.GET("/fail", func(c *gin.Context) {
		reporting.Error(c, errors.New("database unreachable"))
		c.Status(http.StatusServiceUnavailable)
	})

	testkit.Serve(router, testkit.Request(t, http.MethodGet, "/fail"))
	reporter.Flush(2 * time.Second)

	frames := exceptionFrames(t, ingest.received())
	var culprit string

	for _, frame := range frames {
		if frame.InApp && strings.Contains(cmp.Or(frame.AbsPath, frame.Filename), "core/reporting/") {
			t.Errorf("reporter frame %s counts as our code", frame.Function)
		}

		if frame.InApp {
			culprit = frame.Function
		}
	}

	if strings.HasPrefix(culprit, "TestAHandlerErrorReachesSentryWithTheHandlerAsCulprit") == false {
		t.Errorf("culprit %q in %+v", culprit, frames)
	}
}

type frame struct {
	Function string `json:"function"`

	AbsPath string `json:"abs_path"`

	Filename string `json:"filename"`

	InApp bool `json:"in_app"`
}

func exceptionFrames(t *testing.T, envelope string) []frame {
	t.Helper()

	for line := range strings.Lines(envelope) {
		var event struct {
			Exception []struct {
				Stacktrace struct {
					Frames []frame `json:"frames"`
				} `json:"stacktrace"`
			} `json:"exception"`
		}
		if json.Unmarshal([]byte(line), &event) == nil && len(event.Exception) > 0 {
			return event.Exception[len(event.Exception)-1].Stacktrace.Frames
		}
	}

	t.Fatalf("no exception in %s", envelope)

	return nil
}

func newIngestReporter(t *testing.T) (*reporting.Reporter, *fakeIngest) {
	t.Helper()

	ingest := &fakeIngest{}
	server := httptest.NewServer(ingest)
	t.Cleanup(server.Close)

	reporter, err := reporting.New("http://public@"+strings.TrimPrefix(server.URL, "http://")+"/1", "test", "")
	if err != nil {
		t.Fatal(err)
	}

	return reporter, ingest
}

type fakeIngest struct {
	mu sync.Mutex

	bodies []string
}

func (f *fakeIngest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	f.mu.Lock()
	f.bodies = append(f.bodies, string(body))
	f.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (f *fakeIngest) received() string {
	f.mu.Lock()
	defer f.mu.Unlock()

	return strings.Join(f.bodies, "\n")
}
