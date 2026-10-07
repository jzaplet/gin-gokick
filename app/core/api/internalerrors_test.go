package api_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/api"
	"gokick/app/core/reporting"
	"gokick/app/core/testkit"
)

func TestInternalErrors(t *testing.T) {
	engine := newEngine(t)
	group := engine.Group("/api", api.InternalErrors())
	group.GET("/reported", func(c *gin.Context) {
		reporting.Error(c, errors.New("database down"))
	})
	group.GET("/answered", func(c *gin.Context) {
		reporting.Error(c, errors.New("cache down"))
		c.JSON(http.StatusOK, gin.H{"id": "u-1"})
	})
	group.GET("/fine", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	t.Run("a reported error without a response becomes 500", func(t *testing.T) {
		res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/api/reported"))
		assertResponse(t, res, http.StatusInternalServerError, `{"general":{"key":"request.internal"}}`)
	})
	t.Run("a written response stays", func(t *testing.T) {
		res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/api/answered"))
		assertResponse(t, res, http.StatusOK, `{"id":"u-1"}`)
	})
	t.Run("no error, no change", func(t *testing.T) {
		if res := testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/api/fine")); res.Code != http.StatusNoContent {
			t.Errorf("status %d", res.Code)
		}
	})
}
