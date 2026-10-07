package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gokick/app/core/api"
	"gokick/app/core/locale"
	"gokick/app/core/view"
)

const KeyNotFound api.Key = "request.not_found"

func NotFound(renderer *view.Renderer, locales *locale.Set) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")

		if c.NegotiateFormat(gin.MIMEHTML, gin.MIMEJSON) == gin.MIMEJSON {
			api.Fail(c, http.StatusNotFound, KeyNotFound)

			return
		}

		locales.Guess(c)
		renderer.Page(c, http.StatusNotFound, "errors/not_found.html", nil)
	}
}
