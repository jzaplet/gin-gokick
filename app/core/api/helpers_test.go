package api_test

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/httpserver"
	"gokick/app/core/testkit"
)

func newEngine(t *testing.T) *gin.Engine {
	t.Helper()

	engine, err := httpserver.NewEngine(gin.TestMode, nil, testkit.DiscardLogs())
	if err != nil {
		t.Fatal(err)
	}

	return engine
}

func assertResponse(t *testing.T, res *httptest.ResponseRecorder, status int, body string) {
	t.Helper()

	if res.Code != status || res.Body.String() != body || res.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("got %d %q %s, want %d %s", res.Code, res.Header().Get("Content-Type"), res.Body.String(), status, body)
	}
}
