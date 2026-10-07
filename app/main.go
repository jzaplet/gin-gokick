package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gokick/app/core/database"
	"gokick/app/core/httpserver"
	"gokick/app/core/locale"
	"gokick/app/core/logging"
	"gokick/app/core/mail"
	"gokick/app/core/reporting"
	"gokick/app/internal/shared/config"
	"gokick/app/internal/shared/router"
	dictionaries "gokick/locale"
	"gokick/migrations"
	"gokick/public"
)

const reportFlushTimeout = 2 * time.Second

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		logging.New(os.Stdout, logging.FormatJSON, slog.LevelInfo).ErrorContext(ctx, "invalid configuration", logging.Err(err))

		return 1
	}

	logger := logging.New(os.Stdout, cfg.Log.Format, cfg.Log.Level)

	locales, err := locale.New(cfg.Locale.Default, cfg.Locale.Locales)
	if err != nil {
		logger.ErrorContext(ctx, "invalid DEFAULT_LOCALE or LOCALES", logging.Err(err))

		return 1
	}

	if err = dictionaries.Require(cfg.Locale.Locales); err != nil {
		logger.ErrorContext(ctx, "a locale has no dictionary", logging.Err(err))

		return 1
	}

	for _, conversion := range cfg.Tracking.UnlabeledConversions() {
		logger.WarnContext(ctx, "google ads conversion has no label", slog.String(logging.KeyDetail, string(conversion)))
	}

	sender, err := mail.New(cfg.SMTP.Server(), cfg.SMTP.From())
	if err != nil {
		logger.ErrorContext(ctx, "invalid SMTP configuration", logging.Err(err))

		return 1
	}

	if cfg.SMTP.Enabled == false {
		logger.WarnContext(ctx, "mail is off, SMTP_ENABLED is not true")
	}

	reporter, err := reporting.New(cfg.Sentry.DSN, cfg.Sentry.Environment, cfg.Sentry.Release)
	if err != nil {
		logger.ErrorContext(ctx, "cannot start error reporting", logging.Err(err))

		return 1
	}
	defer reporter.Flush(reportFlushTimeout)

	browser := reporting.NewBrowser(cfg.Sentry.FrontendDSN, cfg.Sentry.Environment, cfg.Sentry.Release)

	pool, err := database.Open(ctx, cfg.Database.URL())
	if err != nil {
		logger.ErrorContext(ctx, "cannot connect to the database", logging.Err(err))

		return 1
	}
	defer pool.Close()

	if err = database.Migrate(ctx, pool, migrations.FS, logger); err != nil {
		logger.ErrorContext(ctx, "cannot migrate the database", logging.Err(err))

		return 1
	}

	engine, err := router.NewRouter(cfg, logger, reporter, browser, pool, locales, public.FS, sender)
	if err != nil {
		logger.ErrorContext(ctx, "cannot build the router", logging.Err(err))

		return 1
	}

	if err := httpserver.Serve(ctx, cfg.Server.Addr(), engine, logger); err != nil {
		logger.ErrorContext(ctx, "server stopped", logging.Err(err))

		return 1
	}

	return 0
}
