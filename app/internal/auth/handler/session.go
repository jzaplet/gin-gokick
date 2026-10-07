package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"gokick/app/core/api"
	"gokick/app/core/authn"
	"gokick/app/core/reporting"
)

//tsgen:assets/shared/Auth/Session/types/SessionUser.ts SessionUser
type sessionUser struct {
	ID uuid.UUID `json:"id"`
}

func Session(c *gin.Context) {
	userID, err := authn.UserID(c)
	if err != nil {
		reporting.Error(c, err)

		return
	}

	api.JSON(c, http.StatusOK, sessionUser{ID: userID})
}
