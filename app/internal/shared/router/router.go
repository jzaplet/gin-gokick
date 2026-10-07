package router

import (
	"io/fs"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/httpserver"
	"gokick/app/core/locale"
	"gokick/app/core/mail"
	"gokick/app/core/metrics"
	"gokick/app/core/middleware"
	"gokick/app/core/reporting"
	"gokick/app/core/static"
	"gokick/app/core/vite"
	"gokick/app/internal/shared/config"
	"gokick/app/internal/shared/csp"
	sharedhandler "gokick/app/internal/shared/handler"
	"gokick/app/internal/shared/templates"
)

const requestTimeout = 5 * time.Second

func NewRouter(
	cfg *config.Config,
	logger *slog.Logger,
	reporter *reporting.Reporter,
	browser *reporting.Browser,
	pool *pgxpool.Pool,
	locales *locale.Set,
	public fs.FS,
	sender mail.Sender,
) (*gin.Engine, error) {
	router, err := httpserver.NewEngine(cfg.Server.GinMode, cfg.Server.TrustedProxies, logger)
	if err != nil {
		return nil, err
	}

	requestMetrics := metrics.New()
	router.Use(
		middleware.Trace(),
		reporter.Middleware(),
		middleware.AccessLog(logger),
		requestMetrics.Middleware(),
		middleware.Recovery(logger, reporter),
		middleware.RequestTimeout(requestTimeout),
		middleware.SecurityHeaders(csp.Policy(&cfg.Tracking)),
		middleware.CSRF(),
	)

	assets, err := vite.Load(public, logger)
	if err != nil {
		return nil, err
	}

	files, err := static.New(public, vite.HashedAssets)
	if err != nil {
		return nil, err
	}

	renderer, err := templates.Parse(cfg, assets, browser, locales)
	if err != nil {
		return nil, err
	}

	mailer := mail.NewMailer(renderer, sender, cfg.Server.URL)
	notFound := sharedhandler.NotFound(renderer, locales)
	routes(router, cfg, locales, renderer, requestMetrics, pool, mailer, notFound)

	assets.Register(router)
	files.Register(router, notFound)

	return router, nil
}
