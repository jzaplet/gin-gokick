package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

//tsgen:assets/shared/Fetch/Envelope/ApiMessageKey.ts ApiMessageKey union
type Key string

type Message struct {
	Key Key `json:"key"`

	Params map[string]any `json:"params,omitempty"`
}

type Errors map[string]Message

const General = "general"

func Invalid(c *gin.Context, errs Errors) {
	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, errs)
}

func Fail(c *gin.Context, status int, key Key) {
	c.AbortWithStatusJSON(status, Errors{General: {Key: key}})
}
