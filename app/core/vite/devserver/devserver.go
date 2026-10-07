package devserver

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/netip"
	"net/url"

	"github.com/gin-gonic/gin"

	"gokick/app/core/logging"
	"gokick/app/core/vite/tags"
)

var methods = []string{http.MethodGet, http.MethodHead}

type Server struct {
	base string

	proxy *httputil.ReverseProxy
}

func New(target *url.URL, base string, logger *slog.Logger) *Server {
	return &Server{base: base, proxy: newProxy(target, logger)}
}

func (s *Server) Client() string {
	return s.base + "@vite/client"
}

func (s *Server) Entry(name string) (tags.Set, error) {
	var set tags.Set
	set.Add(s.base + name)

	return set, nil
}

func (s *Server) Module(name string) (tags.Set, error) {
	var set tags.Set
	set.Preload(s.base + name)

	return set, nil
}

func (s *Server) URL(name string) (string, error) {
	return s.base + name, nil
}

func (s *Server) Register(router gin.IRoutes) {
	router.Match(methods, s.base+"*filepath", s.serve)
}

func (s *Server) serve(c *gin.Context) {
	if local(c.Request.RemoteAddr) == false {
		c.AbortWithStatus(http.StatusForbidden)

		return
	}

	s.proxy.ServeHTTP(c.Writer, c.Request.WithContext(context.WithoutCancel(c.Request.Context())))
}

func newProxy(target *url.URL, logger *slog.Logger) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(target)
			r.Out.Host = r.In.Host
		},
		ErrorHandler: func(w http.ResponseWriter, _ *http.Request, err error) {
			http.Error(w, "vite dev server: "+err.Error(), http.StatusBadGateway)
		},
		ErrorLog: logging.ErrorLog(logger),
	}
}

func local(remoteAddr string) bool {
	addr, err := netip.ParseAddrPort(remoteAddr)

	return err == nil && addr.Addr().Unmap().IsLoopback()
}
