package routertest

import (
	"bytes"
	"io/fs"
	"log/slog"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/locale"
	"gokick/app/core/mail"
	"gokick/app/core/reporting"
	"gokick/app/core/testkit"
	"gokick/app/core/testkit/mailbox"
	"gokick/app/core/view"
	"gokick/app/core/vite"
	"gokick/app/internal/shared/config"
	"gokick/app/internal/shared/config/server"
	"gokick/app/internal/shared/localeimages"
	"gokick/app/internal/shared/router"
	"gokick/app/internal/shared/templates"
	dictionaries "gokick/locale"
	"gokick/migrations"
	"gokick/views"
)

type Parts struct {
	Config *config.Config

	Logger *slog.Logger

	Reporter *reporting.Reporter

	Browser *reporting.Browser

	Pool *pgxpool.Pool

	Locales *locale.Set

	Public fs.FS

	Sender mail.Sender
}

func Build(tb testing.TB, parts *Parts) *gin.Engine {
	tb.Helper()

	filled := *parts
	if filled.Config == nil {
		filled.Config = Config()
	}

	if filled.Logger == nil {
		filled.Logger = testkit.DiscardLogs()
	}

	if filled.Reporter == nil {
		filled.Reporter = &reporting.Reporter{}
	}

	if filled.Locales == nil {
		filled.Locales = Locales(tb, "cs_CZ", "cs_CZ")
	}

	if filled.Public == nil {
		filled.Public = Public(tb, filled.Locales)
	}

	if filled.Sender == nil {
		filled.Sender = mail.Discard
	}

	engine, err := router.NewRouter(
		filled.Config,
		filled.Logger,
		filled.Reporter,
		filled.Browser,
		filled.Pool,
		filled.Locales,
		filled.Public,
		filled.Sender,
	)
	if err != nil {
		tb.Fatal(err)
	}

	return engine
}

func Config() *config.Config {
	return &config.Config{Server: server.Config{GinMode: gin.TestMode, URL: "https://gokick.dev"}}
}

func Locales(tb testing.TB, defaultLocale string, names ...string) *locale.Set {
	tb.Helper()

	set, err := locale.New(defaultLocale, names)
	if err != nil {
		tb.Fatal(err)
	}

	return set
}

func Files(tb testing.TB, locales *locale.Set) []string {
	tb.Helper()
	named := testkit.TemplateLiterals(tb, views.FS, "vite", "asset")

	names := slices.Concat(named, localeimages.Flags(locales.Languages()), localeimages.OGImages(locales.Locales()))
	for _, l := range locales.Locales() {
		names = append(names, dictionaries.Module(l.String()))
	}

	return names
}

func Public(tb testing.TB, locales *locale.Set) fstest.MapFS {
	tb.Helper()

	return testkit.Public(tb, Files(tb, locales)...)
}

func New(tb testing.TB) (*gin.Engine, *pgxpool.Pool) {
	tb.Helper()
	pool := testkit.Pool(tb, migrations.FS)

	return Build(tb, &Parts{Pool: pool}), pool
}

func WithLogs(tb testing.TB) (*gin.Engine, *pgxpool.Pool, *bytes.Buffer) {
	tb.Helper()

	logger, logs := testkit.CaptureLogs()
	pool := testkit.Pool(tb, migrations.FS)

	return Build(tb, &Parts{Logger: logger, Pool: pool}), pool, logs
}

func WithLocales(tb testing.TB, defaultLocale string, names ...string) (*gin.Engine, *pgxpool.Pool) {
	tb.Helper()
	pool := testkit.Pool(tb, migrations.FS)

	return Build(tb, &Parts{Pool: pool, Locales: Locales(tb, defaultLocale, names...)}), pool
}

func WithMailbox(tb testing.TB, defaultLocale string, names ...string) (*gin.Engine, *pgxpool.Pool, *mailbox.Mailbox) {
	tb.Helper()

	sent := testkit.Mailbox()
	pool := testkit.Pool(tb, migrations.FS)

	return Build(tb, &Parts{Pool: pool, Locales: Locales(tb, defaultLocale, names...), Sender: sent}), pool, sent
}

func Renderer(tb testing.TB, locales *locale.Set) *view.Renderer {
	tb.Helper()

	assets, err := vite.Load(Public(tb, locales), testkit.DiscardLogs())
	if err != nil {
		tb.Fatal(err)
	}

	renderer, err := templates.Parse(&config.Config{}, assets, nil, locales)
	if err != nil {
		tb.Fatal(err)
	}

	return renderer
}
