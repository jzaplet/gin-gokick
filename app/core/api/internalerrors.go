package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const KeyInternal Key = "request.internal"

func InternalErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}

		Fail(c, http.StatusInternalServerError, KeyInternal)
	}
}
