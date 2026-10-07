package postgres

import (
	"context"
	"crypto/rand"
	"fmt"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/database"
	"gokick/app/core/testkit/dotenv"
	"gokick/app/core/testkit/logs"
)

func Schema(tb testing.TB) string {
	tb.Helper()
	address := testAddress(tb)
	dsn := address.URL()

	conn, err := pgx.Connect(tb.Context(), dsn)
	if err != nil {
		unavailable(tb, err)
	}
	defer closeConn(tb, conn)

	schema := pgx.Identifier{"test_" + strings.ToLower(rand.Text())}.Sanitize()
	if _, err := conn.Exec(tb.Context(), "CREATE SCHEMA "+schema); err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(func() { dropSchema(tb, dsn, schema) })
	address.Params.Set("search_path", strings.Trim(schema, `"`))

	return address.URL()
}

func testAddress(tb testing.TB) database.Address {
	tb.Helper()

	dir, err := os.Getwd()
	if err != nil {
		tb.Fatal(err)
	}

	env, err := dotenv.Lookup(dir, "DB_TEST_NAME", "DB_HOST", "DB_PORT", "DB_USERNAME", "DB_PASSWORD", "DB_TEST_PARAMS")
	if err != nil {
		unavailable(tb, err)
	}

	for _, key := range []string{"DB_HOST", "DB_USERNAME", "DB_PASSWORD"} {
		if env[key] == "" {
			unavailable(tb, fmt.Errorf("%s is empty next to DB_TEST_NAME, see .env.example", key))
		}
	}

	port, err := database.ParsePort(env["DB_PORT"])
	if err != nil {
		tb.Fatalf("DB_PORT %v", err)
	}

	params, err := database.ParseParams(env["DB_TEST_PARAMS"])
	if err != nil {
		tb.Fatalf("DB_TEST_PARAMS %v", err)
	}

	return database.Address{
		Host: env["DB_HOST"],

		Port: port,

		Name: env["DB_TEST_NAME"],

		Username: env["DB_USERNAME"],

		Password: env["DB_PASSWORD"],

		Params: params,
	}
}

func Pool(tb testing.TB, migrations fs.FS) *pgxpool.Pool {
	tb.Helper()

	pool, err := database.Open(tb.Context(), Schema(tb))
	if err != nil {
		tb.Fatal(err)
	}

	tb.Cleanup(pool.Close)

	if err := database.Migrate(tb.Context(), pool, migrations, logs.Discard()); err != nil {
		tb.Fatal(err)
	}

	return pool
}

func Count(tb testing.TB, pool *pgxpool.Pool, query string, args ...any) int {
	tb.Helper()

	var n int
	if err := pool.QueryRow(tb.Context(), query, args...).Scan(&n); err != nil {
		tb.Fatal(err)
	}

	return n
}

func unavailable(tb testing.TB, err error) {
	tb.Helper()

	if os.Getenv("CI") != "" {
		tb.Fatalf("test database: %v", err)
	}

	tb.Skipf("test database unavailable, start it with docker compose up -d db: %v", err)
}

func dropSchema(tb testing.TB, dsn, schema string) {
	tb.Helper()

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		tb.Error(err)

		return
	}
	defer closeConn(tb, conn)

	if _, err := conn.Exec(ctx, "DROP SCHEMA "+schema+" CASCADE"); err != nil {
		tb.Error(err)
	}
}

func closeConn(tb testing.TB, conn *pgx.Conn) {
	tb.Helper()

	if err := conn.Close(context.Background()); err != nil {
		tb.Error(err)
	}
}
