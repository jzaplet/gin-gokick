package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"gokick/app/core/view"
)

func Home(renderer *view.Renderer) gin.HandlerFunc {
	return func(c *gin.Context) {
		renderer.Page(c, http.StatusOK, "home/home.html", nil)
	}
}
