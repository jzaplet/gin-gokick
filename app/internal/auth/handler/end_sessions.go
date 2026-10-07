package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/api"
	"gokick/app/core/authn"
	"gokick/app/core/reporting"
	"gokick/app/internal/auth/session"
)

//tsgen:assets/app/Auth/types/EndSessionsRequest.ts EndSessionsRequest request
type endSessionsRequest struct {
	IDs []uuid.UUID `json:"ids" binding:"max=100"`

	All bool `json:"all"`

	IP string `json:"ip" binding:"max=45"`
}

//tsgen:assets/app/Auth/types/EndedSessions.ts EndedSessions
type endedSessions struct {
	Ended int64 `json:"ended"`
}

func EndSessions(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := authn.UserID(c)
		if err != nil {
			reporting.Error(c, err)

			return
		}

		var req endSessionsRequest
		if api.Bind(c, &req) == false {
			return
		}

		token, _ := c.Cookie(session.CookieName)
		sessions := session.New(pool)

		var ended int64
		if req.All {
			ended, err = sessions.EndAllOthers(c.Request.Context(), userID, token, session.Filter{IP: req.IP})
		} else {
			ended, err = sessions.EndOthers(c.Request.Context(), userID, token, req.IDs)
		}

		if err != nil {
			reporting.Error(c, err)

			return
		}

		api.JSON(c, http.StatusOK, endedSessions{Ended: ended})
	}
}
