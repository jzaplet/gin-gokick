package authn

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gokick/app/core/api"
	"gokick/app/core/logging"
	"gokick/app/core/reporting"
)

const KeySignInRequired api.Key = "auth.sign_in_required"

var errNoMiddleware = errors.New("the route has no authn.Middleware, so it has no user")

type Sessions interface {
	User(ctx context.Context, token, ip string) (uuid.UUID, bool, error)
}

type userKey struct{}

func Middleware(sessions Sessions, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookieName)
		if err != nil {
			api.Fail(c, http.StatusUnauthorized, KeySignInRequired)

			return
		}

		userID, ok, err := sessions.User(c.Request.Context(), token, c.ClientIP())
		if err != nil {
			reporting.Error(c, err)
			c.Abort()

			return
		}

		if ok == false {
			api.Fail(c, http.StatusUnauthorized, KeySignInRequired)

			return
		}

		c.Set(userKey{}, userID)
		c.Request = c.Request.WithContext(logging.WithUserID(c.Request.Context(), userID.String()))
		c.Next()
	}
}

func UserID(c *gin.Context) (uuid.UUID, error) {
	value, _ := c.Get(userKey{})

	id, ok := value.(uuid.UUID)
	if ok == false {
		return uuid.Nil, errNoMiddleware
	}

	return id, nil
}
