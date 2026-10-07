package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/api"
	"gokick/app/core/authn"
	"gokick/app/core/reporting"
	"gokick/app/internal/user/profile"
)

//tsgen:assets/app/User/types/CurrentUser.ts CurrentUser
type currentUser struct {
	ID uuid.UUID `json:"id"`

	Email string `json:"email"`
}

func Me(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := authn.UserID(c)
		if err != nil {
			reporting.Error(c, err)

			return
		}

		email, err := profile.New(pool).Email(c.Request.Context(), userID)
		if err != nil {
			reporting.Error(c, err)

			return
		}

		api.JSON(c, http.StatusOK, currentUser{ID: userID, Email: email})
	}
}
