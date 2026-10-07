package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/cookie"
	"gokick/app/core/reporting"
	"gokick/app/internal/auth/session"
)

func Logout(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token, err := c.Cookie(session.CookieName); err == nil {
			if err := session.New(pool).End(c.Request.Context(), token); err != nil {
				reporting.Error(c, err)

				return
			}
		}

		cookie.Clear(c.Writer, session.CookieName)
		c.Status(http.StatusNoContent)
	}
}
