package view

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"gokick/app/core/csp"
	"gokick/app/core/locale"
	"gokick/app/core/locale/posix"
	"gokick/app/core/pageurl"
	"gokick/app/core/reporting"
)

var errUnknownPage = errors.New("templates: unknown page")

var errUnknownLocale = errors.New("templates: no texts of the locale")

type Page struct {
	Nonce string

	Origin string

	URL string

	locale.Page

	Params any
}

func (r *Renderer) Page(c *gin.Context, status int, name string, params any) {
	body, err := r.body(c, name, params)
	if err != nil {
		reporting.Error(c, err)
		c.AbortWithStatus(http.StatusInternalServerError)

		return
	}

	if c.Writer.Header().Get("Cache-Control") == "" {
		c.Header("Cache-Control", "private, no-cache")
	}

	c.Data(status, "text/html; charset=utf-8", body)
}

func (r *Renderer) body(c *gin.Context, name string, params any) ([]byte, error) {
	page, err := locale.PageOf(c)
	if err != nil {
		return nil, err
	}

	return r.execute(name, name, page.Locale, &Page{
		Nonce: csp.Nonce(c),

		Origin: pageurl.Origin(c.Request),

		URL: pageurl.Of(c.Request),

		Page: page,

		Params: params,
	})
}

func (r *Renderer) execute(name, entry string, l posix.Locale, data any) ([]byte, error) {
	localized, ok := r.pages[name]
	if ok == false {
		return nil, fmt.Errorf("%w: %s", errUnknownPage, name)
	}

	page, ok := localized[l.String()]
	if ok == false {
		return nil, fmt.Errorf("%w: %s", errUnknownLocale, l)
	}

	var body bytes.Buffer
	if err := page.ExecuteTemplate(&body, entry, data); err != nil {
		return nil, err
	}

	return body.Bytes(), nil
}
