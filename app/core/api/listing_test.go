package api_test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/api"
	"gokick/app/core/listing"
	"gokick/app/core/testkit"
)

type sort string

type filters struct {
	Search string `form:"search" binding:"max=5"`
}

var grid = listing.Grid[sort]{Sorts: []sort{"name", "email"}, Direction: listing.Descending, PerPage: 10}

func TestBindListing(t *testing.T) {
	cases := []struct {
		query string

		status int

		want string
	}{
		{"", http.StatusOK, `{"Page":1,"PerPage":10,"Sort":"name","Direction":"desc","Search":""}`},
		{"?page=3&perPage=50&sortBy=email&sortDir=asc&search=jan&other=x", http.StatusOK, `{"Page":3,"PerPage":50,"Sort":"email","Direction":"asc","Search":"jan"}`},
		{"?page=0&sortBy=age&search=jaroslav", http.StatusUnprocessableEntity, `{"page":{"key":"validation.min","params":{"min":"1"}},"search":{"key":"validation.max_length","params":{"max":5}},"sortBy":{"key":"validation.one_of","params":{"values":"name, email"}}}`},
		{"?perPage=101&sortDir=up", http.StatusUnprocessableEntity, `{"perPage":{"key":"validation.max","params":{"max":"100"}},"sortDir":{"key":"validation.one_of","params":{"values":"asc, desc"}}}`},
		{"?page=abc", http.StatusBadRequest, `{"general":{"key":"request.invalid_query"}}`},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			engine := newEngine(t)
			engine.GET("/", func(c *gin.Context) {
				var found filters
				if query, ok := api.BindListing(c, grid, &found); ok {
					c.JSON(http.StatusOK, struct {
						listing.Query[sort]
						filters
					}{query, found})
				}
			})

			assertResponse(t, testkit.Serve(engine, testkit.Request(t, http.MethodGet, "/"+tc.query)), tc.status, tc.want)
		})
	}
}
