package testkit

import (
	"bytes"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/testkit/dotenv"
	"gokick/app/core/testkit/logs"
	"gokick/app/core/testkit/mailbox"
	"gokick/app/core/testkit/postgres"
	"gokick/app/core/testkit/web"
	"gokick/app/core/testkit/webroot"
)

func Request(tb testing.TB, method, path string) *http.Request {
	tb.Helper()

	return web.Request(tb, method, path)
}

func JSONRequest(tb testing.TB, method, path, body string) *http.Request {
	tb.Helper()

	return web.JSONRequest(tb, method, path, body)
}

func Serve(handler http.Handler, req *http.Request) *httptest.ResponseRecorder {
	return web.Serve(handler, req)
}

func Markup(res *httptest.ResponseRecorder) string {
	return web.Markup(res)
}

func MarkupOf(html string) string {
	return web.MarkupOf(html)
}

func Cookie(tb testing.TB, res *httptest.ResponseRecorder, name string) *http.Cookie {
	tb.Helper()

	return web.Cookie(tb, res, name)
}

func Nonce(tb testing.TB, res *httptest.ResponseRecorder) string {
	tb.Helper()

	return web.Nonce(tb, res)
}

func DiscardLogs() *slog.Logger {
	return logs.Discard()
}

func CaptureLogs() (*slog.Logger, *bytes.Buffer) {
	return logs.Capture()
}

func LogEntries(tb testing.TB, buf *bytes.Buffer) []map[string]any {
	tb.Helper()

	return logs.Entries(tb, buf)
}

func Mailbox() *mailbox.Mailbox {
	return &mailbox.Mailbox{}
}

func LoadEnv(tb testing.TB) {
	tb.Helper()
	dotenv.Load(tb)
}

func Database(tb testing.TB) string {
	tb.Helper()

	return postgres.Schema(tb)
}

func Pool(tb testing.TB, migrations fs.FS) *pgxpool.Pool {
	tb.Helper()

	return postgres.Pool(tb, migrations)
}

func Count(tb testing.TB, pool *pgxpool.Pool, query string, args ...any) int {
	tb.Helper()

	return postgres.Count(tb, pool, query, args...)
}

func Public(tb testing.TB, names ...string) fstest.MapFS {
	tb.Helper()

	return webroot.Public(tb, names...)
}

func TemplateLiterals(tb testing.TB, templates fs.FS, funcs ...string) []string {
	tb.Helper()

	return webroot.TemplateLiterals(tb, templates, funcs...)
}
