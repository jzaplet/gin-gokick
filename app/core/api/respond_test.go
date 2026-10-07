package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gokick/app/core/api"
	"gokick/app/core/testkit"
)

type tag struct {
	Name string `json:"name"`

	Aliases []string `json:"aliases"`
}

type tagList struct {
	Tags []tag `json:"tags"`

	Empty []tag `json:"empty"`

	Main tag `json:"main"`

	Pinned *tag `json:"pinned"`

	Missing *tag `json:"missing"`

	Meta map[string]any `json:"meta"`

	Extra map[string]any `json:"extra"`

	ID uuid.UUID `json:"id"`

	At time.Time `json:"at"`

	hidden []string
}

func TestJSONSendsNilCollectionsAsEmptyOnes(t *testing.T) {
	body := &tagList{
		Tags: []tag{{Name: "a"}},

		Pinned: &tag{Name: "b"},

		Meta: map[string]any{"list": []int(nil)},

		At: time.Unix(0, 0).UTC(),
	}

	res := serveJSON(t, body)

	want := `{"tags":[{"name":"a","aliases":[]}],"empty":[],"main":{"name":"","aliases":[]},"pinned":{"name":"b","aliases":[]},` +
		`"missing":null,"meta":{"list":[]},"extra":{},"id":"00000000-0000-0000-0000-000000000000","at":"1970-01-01T00:00:00Z"}`
	assertResponse(t, res, http.StatusOK, want)

	if body.Tags[0].Aliases != nil || body.Empty != nil || body.Pinned.Aliases != nil || body.Extra != nil || body.hidden != nil {
		t.Errorf("the handler's value changed: %+v", body)
	}

	assertResponse(t, serveJSON(t, []tag(nil)), http.StatusOK, `[]`)
	assertResponse(t, serveJSON(t, nil), http.StatusOK, `null`)
}

func serveJSON(t *testing.T, body any) *httptest.ResponseRecorder {
	t.Helper()
	engine := newEngine(t)
	engine.GET("/", func(c *gin.Context) {
		api.JSON(c, http.StatusOK, body)
	})

	return testkit.Serve(engine, testkit.JSONRequest(t, http.MethodGet, "/", ``))
}

type node struct {
	Next *node `json:"next"`
}

func TestJSONStopsAtACycleWithoutOverflowingTheStack(t *testing.T) {
	cycle := &node{}
	cycle.Next = cycle

	engine := newEngine(t)
	engine.GET("/", api.InternalErrors(), func(c *gin.Context) {
		api.JSON(c, http.StatusOK, cycle)
	})

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodGet, "/", ``))

	assertResponse(t, res, http.StatusInternalServerError, `{"general":{"key":"request.internal"}}`)
}
