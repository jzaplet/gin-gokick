package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/api"
	"gokick/app/core/cookie"
	"gokick/app/core/locale"
	"gokick/app/core/mail"
	"gokick/app/core/password"
	"gokick/app/core/ratelimit"
	"gokick/app/core/reporting"
	"gokick/app/internal/auth/session"
	"gokick/app/internal/auth/session/clientinfo"
	"gokick/app/internal/user/account"
	usermail "gokick/app/internal/user/mail"
)

const KeyEmailTaken api.Key = "user.email_taken"

const cleanupTimeout = 5 * time.Second

//tsgen:assets/app/Auth/types/RegisterRequest.ts RegisterRequest request
type registerRequest struct {
	Email string `json:"email" binding:"required,email,max=254"`

	Password string `json:"password" binding:"required,min=12,max=128"`

	Locale string `json:"locale" binding:"required,max=5"`
}

func Register(pool *pgxpool.Pool, locales *locale.Set, mailer *mail.Mailer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req registerRequest
		if api.Bind(c, &req) == false {
			return
		}

		accounts := account.New(pool, password.OWASP)
		nearest := locales.Nearest(req.Locale)
		userLocale := nearest.String()

		userID, err := accounts.Register(c.Request.Context(), req.Email, req.Password, userLocale)
		if errors.Is(err, account.ErrEmailTaken) {
			if recordErr := ratelimit.Failed(c); recordErr != nil {
				reporting.Error(c, recordErr)

				return
			}

			api.Invalid(c, api.Errors{"email": {Key: KeyEmailTaken}})

			return
		}

		if err != nil {
			reporting.Error(c, err)

			return
		}

		previous, _ := c.Cookie(session.CookieName)
		client := clientinfo.Client{UserAgent: c.Request.UserAgent(), IP: c.ClientIP()}

		token, err := session.New(pool).Start(c.Request.Context(), userID, client, previous)
		if err != nil {
			reporting.Error(c, err)

			cleanup, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), cleanupTimeout)
			defer cancel()

			if deleteErr := accounts.Delete(cleanup, userID); deleteErr != nil {
				reporting.Error(c, deleteErr)
			}

			return
		}

		cookie.Set(c.Writer, session.CookieName, token, session.Lifetime)

		if err := usermail.Welcome(c.Request.Context(), mailer, nearest, req.Email); err != nil {
			reporting.Warning(c, err)
		}

		api.JSON(c, http.StatusCreated, signedInUser{ID: userID, Locale: userLocale})
	}
}
