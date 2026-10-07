package handler

import (
	"net/http"
	"net/netip"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/api"
	"gokick/app/core/authn"
	"gokick/app/core/listing"
	"gokick/app/core/reporting"
	"gokick/app/internal/auth/session"
)

//tsgen:assets/app/Auth/types/ListedSession.ts ListedSession
type listedSession struct {
	ID uuid.UUID `json:"id"`

	UserAgent string `json:"userAgent"`

	CreatedAt time.Time `json:"createdAt"`

	LastSeenAt time.Time `json:"lastSeenAt"`

	LastSeenIP *netip.Addr `json:"lastSeenIp"`

	Current bool `json:"current"`
}

type sessionFilters struct {
	IP string `form:"ip" binding:"max=45"`
}

func Sessions(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := authn.UserID(c)
		if err != nil {
			reporting.Error(c, err)

			return
		}

		var filters sessionFilters

		query, ok := api.BindListing(c, session.Grid, &filters)
		if ok == false {
			return
		}

		token, _ := c.Cookie(session.CookieName)
		filter := session.Filter{IP: filters.IP}

		found, err := session.New(pool).List(c.Request.Context(), userID, token, query, filter)
		if err != nil {
			reporting.Error(c, err)

			return
		}

		items := make([]listedSession, len(found.Sessions))
		for i, listed := range found.Sessions {
			items[i] = listedSession{
				ID: listed.ID,

				UserAgent: listed.UserAgent,

				CreatedAt: listed.CreatedAt,

				LastSeenAt: listed.LastSeenAt,

				LastSeenIP: listed.LastSeenIp,

				Current: listed.IsCurrent,
			}
		}

		api.JSON(c, http.StatusOK, listing.SelectableResult[listedSession]{
			Items: items,

			Total: found.Total,

			Selectable: found.Others,
		})
	}
}
