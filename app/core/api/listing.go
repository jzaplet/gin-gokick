package api

import (
	"cmp"
	"errors"
	"maps"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"gokick/app/core/listing"
)

const KeyInvalidQuery Key = "request.invalid_query"

type listingQuery struct {
	Page int `form:"page,default=1" binding:"min=1,max=100000"`

	PerPage int `form:"perPage" binding:"omitempty,min=1,max=100"`

	SortBy string `form:"sortBy"`

	SortDir listing.Direction `form:"sortDir" binding:"omitempty,oneof=asc desc"`
}

func BindListing[S ~string](c *gin.Context, grid listing.Grid[S], filters any) (listing.Query[S], bool) {
	wireNames.Do(useWireNames)

	errs := Errors{}

	var raw listingQuery
	if bindQuery(c, &raw, errs) == false {
		return listing.Query[S]{}, false
	}

	sort := S(raw.SortBy)
	if sort == "" && len(grid.Sorts) > 0 {
		sort = grid.Sorts[0]
	}

	if slices.Contains(grid.Sorts, sort) == false {
		errs["sortBy"] = oneOf(grid.Sorts)
	}

	if filters != nil && bindQuery(c, filters, errs) == false {
		return listing.Query[S]{}, false
	}

	if len(errs) > 0 {
		Invalid(c, errs)

		return listing.Query[S]{}, false
	}

	return listing.Query[S]{
		Page: raw.Page,

		PerPage: cmp.Or(raw.PerPage, grid.PerPage, listing.DefaultPerPage),

		Sort: sort,

		Direction: cmp.Or(raw.SortDir, grid.Direction, listing.Ascending),
	}, true
}

func bindQuery(c *gin.Context, dst any, errs Errors) bool {
	err := c.ShouldBindQuery(dst)
	if fields, ok := errors.AsType[validator.ValidationErrors](err); ok {
		maps.Copy(errs, fieldErrors(fields))

		return true
	}

	if err != nil {
		Fail(c, http.StatusBadRequest, KeyInvalidQuery)

		return false
	}

	return true
}
