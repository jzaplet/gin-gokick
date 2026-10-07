package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/api"
	"gokick/app/core/authn"
	"gokick/app/core/health"
	"gokick/app/core/locale"
	"gokick/app/core/mail"
	"gokick/app/core/metrics"
	"gokick/app/core/ratelimit"
	"gokick/app/core/view"
	"gokick/app/internal/auth/failures"
	authhandler "gokick/app/internal/auth/handler"
	"gokick/app/internal/auth/session"
	homehandler "gokick/app/internal/home/handler"
	"gokick/app/internal/shared/config"
	sharedhandler "gokick/app/internal/shared/handler"
	userhandler "gokick/app/internal/user/handler"
)

const readinessTimeout = 2 * time.Second

func routes(
	router *gin.Engine,
	cfg *config.Config,
	locales *locale.Set,
	renderer *view.Renderer,
	requestMetrics *metrics.Metrics,
	pool *pgxpool.Pool,
	mailer *mail.Mailer,
	notFound gin.HandlerFunc,
) {
	router.GET("/", locales.Home(), homehandler.Home(renderer))
	router.GET("/"+locales.DefaultLanguage(), locale.RedirectHome)

	for _, lang := range locales.OtherLanguages() {
		router.GET("/"+lang, locales.Prefix(notFound), homehandler.Home(renderer))
	}

	router.GET("/:lang/app", locales.Prefix(notFound), sharedhandler.App(renderer))
	router.GET("/:lang/app/*path", locales.Prefix(notFound), sharedhandler.App(renderer))
	router.GET("/ping", health.Ping)
	router.GET("/readyz", health.Handler(readinessTimeout, health.Check{Name: "database", Probe: pool.Ping}))
	router.GET("/metrics", metrics.RequireBasicAuth(cfg.Prometheus.User, cfg.Prometheus.Password), requestMetrics.Handler())

	failedAttempts := ratelimit.Middleware(failures.New(pool, "auth"), 10, 15*time.Minute)
	apiRoutes := router.Group("", api.InternalErrors())
	apiRoutes.POST("/api/auth/register", failedAttempts, authhandler.Register(pool, locales, mailer))
	apiRoutes.POST("/api/auth/login", failedAttempts, authhandler.Login(pool, locales))
	apiRoutes.POST("/api/auth/logout", authhandler.Logout(pool))

	signedIn := apiRoutes.Group("", authn.Middleware(session.New(pool), session.CookieName))
	signedIn.GET("/api/auth/session", authhandler.Session)
	signedIn.GET("/api/auth/sessions", authhandler.Sessions(pool))
	signedIn.POST("/api/auth/sessions/end", authhandler.EndSessions(pool))
	signedIn.GET("/api/user/me", userhandler.Me(pool))
	signedIn.PUT("/api/user/locale", userhandler.Locale(pool, locales))
}
