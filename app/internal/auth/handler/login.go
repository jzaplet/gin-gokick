package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/api"
	"gokick/app/core/cookie"
	"gokick/app/core/locale"
	"gokick/app/core/password"
	"gokick/app/core/ratelimit"
	"gokick/app/core/reporting"
	"gokick/app/internal/auth/session"
	"gokick/app/internal/auth/session/clientinfo"
	"gokick/app/internal/user/account"
)

const KeyLoginFailed api.Key = "auth.login_failed"

//tsgen:assets/app/Auth/types/LoginRequest.ts LoginRequest request
type loginRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`

	Password string `json:"password" binding:"required,max=128"`
}

//tsgen:assets/shared/Auth/Login/types/SignedInUser.ts SignedInUser
type signedInUser struct {
	ID uuid.UUID `json:"id"`

	Locale string `json:"locale"`
}

func Login(pool *pgxpool.Pool, locales *locale.Set) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req loginRequest
		if api.Bind(c, &req) == false {
			return
		}

		accounts := account.New(pool, password.OWASP)

		user, rehash, err := accounts.Authenticate(c.Request.Context(), req.Email, req.Password)
		if errors.Is(err, account.ErrLoginFailed) {
			if recordErr := ratelimit.Failed(c); recordErr != nil {
				reporting.Error(c, recordErr)

				return
			}

			api.Fail(c, http.StatusUnauthorized, KeyLoginFailed)

			return
		}

		if err != nil {
			reporting.Error(c, err)

			return
		}

		if rehash {
			if rehashErr := accounts.Rehash(c.Request.Context(), user.ID, req.Password); rehashErr != nil {
				reporting.Warning(c, rehashErr)
			}
		}

		previous, _ := c.Cookie(session.CookieName)
		client := clientinfo.Client{UserAgent: c.Request.UserAgent(), IP: c.ClientIP()}

		token, err := session.New(pool).Start(c.Request.Context(), user.ID, client, previous)
		if err != nil {
			reporting.Error(c, err)

			return
		}

		cookie.Set(c.Writer, session.CookieName, token, session.Lifetime)
		api.JSON(c, http.StatusOK, signedInUser{ID: user.ID, Locale: locales.Nearest(user.Locale).String()})
	}
}
