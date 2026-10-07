package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gokick/app/core/reporting"
)

func CSRF() gin.HandlerFunc {
	protection := http.NewCrossOriginProtection()

	return func(c *gin.Context) {
		if err := protection.Check(c.Request); err != nil {
			reporting.Error(c, err)
			c.AbortWithStatus(http.StatusForbidden)

			return
		}

		c.Next()
	}
}
