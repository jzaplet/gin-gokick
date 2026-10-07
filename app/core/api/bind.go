package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"gokick/app/core/reporting"
)

const KeyInvalidBody Key = "request.invalid_body"

const KeyBodyTooLarge Key = "request.body_too_large"

const MaxBodyBytes = 1 << 20

func Bind(c *gin.Context, dst any) bool {
	wireNames.Do(useWireNames)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	err := c.ShouldBindJSON(dst)
	if err == nil {
		return true
	}

	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		Fail(c, http.StatusRequestEntityTooLarge, KeyBodyTooLarge)

		return false
	}

	if fields, ok := errors.AsType[validator.ValidationErrors](err); ok {
		Invalid(c, fieldErrors(fields))

		return false
	}

	if _, ok := errors.AsType[*json.InvalidUnmarshalError](err); ok {
		reporting.Error(c, err)
		Fail(c, http.StatusInternalServerError, KeyInternal)

		return false
	}

	Fail(c, http.StatusBadRequest, KeyInvalidBody)

	return false
}
