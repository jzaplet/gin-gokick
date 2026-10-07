package locale

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"gokick/app/core/locale/dialect"
	"gokick/app/core/locale/hreflang"
	"gokick/app/core/locale/posix"
)

var errNoMiddleware = errors.New("the route has no locale middleware, so its page has no locale")

type Page struct {
	Locale posix.Locale

	Home string

	Homes []Home

	Alternates []hreflang.Link

	Languages []Language
}

type Language struct {
	Locale posix.Locale

	Path string
}

type Home struct {
	Language string

	Path string
}

type pageKey struct{}

func (s *Set) Home() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(pageKey{}, Page{
			Locale: s.fallback,

			Home: "/",

			Homes: s.homes,

			Alternates: hreflang.Links(c.Request, s.languages, s.DefaultLanguage(), ""),

			Languages: s.languagesOf(""),
		})
	}
}

func (s *Set) Prefix(notFound gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang, dialects, ok := s.prefix(c)
		if ok == false {
			notFound(c)
			c.Abort()

			return
		}

		c.Writer.Header().Add("Vary", "Accept-Language")
		page := strings.TrimPrefix(c.Request.URL.EscapedPath(), "/"+lang)
		c.Set(pageKey{}, Page{
			Locale: dialects.Match(c.GetHeader("Accept-Language")),

			Home: hreflang.Path(lang, s.DefaultLanguage(), ""),

			Homes: s.homes,

			Alternates: hreflang.Links(c.Request, s.languages, s.DefaultLanguage(), page),

			Languages: s.languagesOf(page),
		})
	}
}

func (s *Set) Guess(c *gin.Context) {
	_, dialects, ok := s.prefix(c)
	if ok {
		c.Writer.Header().Add("Vary", "Accept-Language")
	} else {
		dialects = s.dialects[s.DefaultLanguage()]
	}

	c.Set(pageKey{}, s.Page(dialects.Match(c.GetHeader("Accept-Language"))))
}

func (s *Set) Page(l posix.Locale) Page {
	return Page{
		Locale: l,

		Home: hreflang.Path(l.Language(), s.DefaultLanguage(), ""),

		Homes: s.homes,

		Languages: s.languagesOf(""),
	}
}

func (s *Set) languagesOf(page string) []Language {
	languages := make([]Language, len(s.switcher))
	for i, l := range s.switcher {
		languages[i] = Language{Locale: l, Path: hreflang.Path(l.Language(), s.DefaultLanguage(), page)}
	}

	return languages
}

func (s *Set) prefix(c *gin.Context) (string, *dialect.Matcher, bool) {
	lang, _, _ := strings.Cut(strings.TrimPrefix(c.Request.URL.EscapedPath(), "/"), "/")
	dialects, ok := s.dialects[lang]

	return lang, dialects, ok
}

func PageOf(c *gin.Context) (Page, error) {
	value, _ := c.Get(pageKey{})

	page, ok := value.(Page)
	if ok == false {
		return Page{}, errNoMiddleware
	}

	return page, nil
}

func RedirectHome(c *gin.Context) {
	home := url.URL{Path: "/", RawQuery: c.Request.URL.RawQuery}
	c.Redirect(http.StatusMovedPermanently, home.String())
}
