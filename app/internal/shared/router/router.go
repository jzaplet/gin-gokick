package router

import (
	"io/fs"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/business"
	"gokick/app/core/httpserver"
	"gokick/app/core/locale"
	"gokick/app/core/mail"
	"gokick/app/core/middleware"
	"gokick/app/core/reporting"
	"gokick/app/core/static"
	"gokick/app/core/telemetry"
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
	tel *telemetry.Telemetry,
	pool *pgxpool.Pool,
	locales *locale.Set,
	public fs.FS,
	sender mail.Sender,
) (*gin.Engine, error) {
	router, err := httpserver.NewEngine(cfg.Server.GinMode, cfg.Server.TrustedProxies, logger)
	if err != nil {
		return nil, err
	}

	router.Use(tel.Middleware("/ping", "/readyz")...)
	router.Use(
		reporter.Middleware(),
		middleware.AccessLog(logger),
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

	events := business.New(logger, tel)
	mailer := mail.NewMailer(renderer, sender, cfg.App.URL, events)
	notFound := sharedhandler.NotFound(renderer, locales)
	routes(router, locales, renderer, pool, mailer, events, notFound)

	assets.Register(router)
	files.Register(router, notFound)

	return router, nil
}
