package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"gokick/app/core/api"
	"gokick/app/core/testkit"
)

const validSignUp = `{"email":"jan@example.com","password":"correct horse"}`

type signUp struct {
	Email string `json:"email" binding:"required,email,max=254"`

	Password string `json:"password" binding:"required,min=12"`

	Age int `json:"age" binding:"omitempty,min=18"`

	Plan string `json:"plan" binding:"omitempty,oneof=start business"`

	ID string `json:"id" binding:"omitempty,uuid"`

	Code string `json:"code" binding:"omitempty,alpha"`
}

func TestBind(t *testing.T) {
	cases := []struct {
		name string

		body string

		status int

		want string
	}{
		{
			name: "valid",

			body: validSignUp,

			status: http.StatusOK,

			want: `{"email":"jan@example.com","password":"correct horse","age":0,"plan":"","id":"","code":""}`,
		},

		{
			name: "malformed",

			body: `{"email":`,

			status: http.StatusBadRequest,

			want: `{"general":{"key":"request.invalid_body"}}`,
		},

		{name: "empty", body: ``, status: http.StatusBadRequest, want: `{"general":{"key":"request.invalid_body"}}`},

		{
			name: "wrong type",

			body: `{"email":1}`,

			status: http.StatusBadRequest,

			want: `{"general":{"key":"request.invalid_body"}}`,
		},

		{
			name: "missing fields",

			body: `{}`,

			status: http.StatusUnprocessableEntity,

			want: `{"email":{"key":"validation.required"},"password":{"key":"validation.required"}}`,
		},

		{
			name: "every field wrong",

			body: `{"email":"jan","password":"short","age":3,"plan":"gold","id":"x","code":"42"}`,

			status: http.StatusUnprocessableEntity,

			want: `{"age":{"key":"validation.min","params":{"min":"18"}},"code":{"key":"validation.invalid"},"email":{"key":"validation.email"},"id":{"key":"validation.uuid"},"password":{"key":"validation.min_length","params":{"min":12}},"plan":{"key":"validation.one_of","params":{"values":"start, business"}}}`,
		},

		{
			name: "too long",

			body: `{"email":"` + strings.Repeat("a", 250) + `@example.com","password":"correct horse"}`,

			status: http.StatusUnprocessableEntity,

			want: `{"email":{"key":"validation.max_length","params":{"max":254}}}`,
		},

		{
			name: "a body at the limit",

			body: strings.Repeat(" ", api.MaxBodyBytes-len(validSignUp)) + validSignUp,

			status: http.StatusOK,

			want: `{"email":"jan@example.com","password":"correct horse","age":0,"plan":"","id":"","code":""}`,
		},

		{
			name: "a body over the limit",

			body: strings.Repeat(" ", api.MaxBodyBytes) + validSignUp,

			status: http.StatusRequestEntityTooLarge,

			want: `{"general":{"key":"request.body_too_large"}}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertResponse(t, bindSignUp(t, tc.body), tc.status, tc.want)
		})
	}
}

type address struct {
	Street string `json:"street" binding:"required"`
}

type item struct {
	Name string `json:"name" binding:"required"`
}

type order struct {
	Billing address `json:"billing"`

	Shipping *address `json:"shipping"`

	Items []item `json:"items" binding:"required,min=1,dive"`
}

func TestBindNamesNestedErrorsByTheirWholePath(t *testing.T) {
	engine := newEngine(t)
	engine.POST("/", func(c *gin.Context) {
		var req order
		if api.Bind(c, &req) {
			c.Status(http.StatusOK)
		}
	})

	body := `{"billing":{"street":""},"shipping":{"street":""},"items":[{"name":"a"},{"name":""}]}`

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/", body))

	want := `{"billing.street":{"key":"validation.required"},"items[1].name":{"key":"validation.required"},"shipping.street":{"key":"validation.required"}}`
	assertResponse(t, res, http.StatusUnprocessableEntity, want)
}

func TestBindCountsTheItemsOfAList(t *testing.T) {
	engine := newEngine(t)
	engine.POST("/", func(c *gin.Context) {
		var req order
		if api.Bind(c, &req) {
			c.Status(http.StatusOK)
		}
	})

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/", `{"billing":{"street":"a"},"items":[]}`))

	assertResponse(t, res, http.StatusUnprocessableEntity, `{"items":{"key":"validation.min_items","params":{"min":1}}}`)
}

func TestBindReportsADestinationThatIsNoPointer(t *testing.T) {
	engine := newEngine(t)
	var reported []*gin.Error

	engine.POST("/", func(c *gin.Context) {
		var req signUp
		api.Bind(c, req)
		reported = c.Errors
	})

	res := testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/", `{}`))

	assertResponse(t, res, http.StatusInternalServerError, `{"general":{"key":"request.internal"}}`)

	if len(reported) != 1 {
		t.Errorf("reported %v", reported)
	}
}

func bindSignUp(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()
	engine := newEngine(t)
	engine.POST("/", func(c *gin.Context) {
		var req signUp
		if api.Bind(c, &req) {
			c.JSON(http.StatusOK, req)
		}
	})

	return testkit.Serve(engine, testkit.JSONRequest(t, http.MethodPost, "/", body))
}
