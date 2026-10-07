package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/api"
	"gokick/app/core/authn"
	"gokick/app/core/locale"
	"gokick/app/core/reporting"
	"gokick/app/internal/user/profile"
)

//tsgen:assets/app/User/types/LocaleRequest.ts LocaleRequest request
type localeRequest struct {
	Locale string `json:"locale" binding:"required,max=5"`
}

func Locale(pool *pgxpool.Pool, locales *locale.Set) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := authn.UserID(c)
		if err != nil {
			reporting.Error(c, err)

			return
		}

		var req localeRequest
		if api.Bind(c, &req) == false {
			return
		}

		if err := profile.New(pool).SetLocale(c.Request.Context(), userID, locales.Nearest(req.Locale).String()); err != nil {
			reporting.Error(c, err)

			return
		}

		c.Status(http.StatusNoContent)
	}
}
