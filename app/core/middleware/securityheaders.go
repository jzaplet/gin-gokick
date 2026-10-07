package middleware

import (
	"crypto/rand"

	"github.com/gin-gonic/gin"

	"gokick/app/core/csp"
)

const permissionsPolicy = "accelerometer=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), magnetometer=(), microphone=(), payment=(), usb=()"

func SecurityHeaders(policy csp.Policy) gin.HandlerFunc {
	contentSecurityPolicy := policy.Header()

	return func(c *gin.Context) {
		nonce := rand.Text()
		c.Request = c.Request.WithContext(csp.WithNonce(c.Request.Context(), nonce))

		header := c.Writer.Header()
		header.Set("Content-Security-Policy", contentSecurityPolicy.Value(nonce))
		header.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		header.Set("Cross-Origin-Opener-Policy", "same-origin")
		header.Set("Cross-Origin-Resource-Policy", "same-origin")
		header.Set("Permissions-Policy", permissionsPolicy)
		header.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		header.Set("X-Content-Type-Options", "nosniff")
		header.Set("X-Frame-Options", "DENY")
		c.Next()
	}
}
