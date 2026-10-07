package failures_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/testkit"
	"gokick/app/internal/auth/failures"
	"gokick/migrations"
)

const window = 15 * time.Minute

func TestCountCountsTheNetworkWithinTheWindowOfItsScope(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	addFailures(t, pool, "login", "203.0.113.7/32", 9, 14*time.Minute)
	addFailures(t, pool, "login", "203.0.113.7/32", 5, 16*time.Minute)
	addFailures(t, pool, "login", "203.0.113.8/32", 3, time.Minute)
	addFailures(t, pool, "reset", "203.0.113.7/32", 4, time.Minute)

	count, err := failures.New(pool, "login").Count(t.Context(), netip.MustParsePrefix("203.0.113.7/32"), window)
	if err != nil || count != 9 {
		t.Errorf("count %d, error %v", count, err)
	}
}

func TestRecordStoresTheNetworkAndDeletesOldFailuresOfItsScope(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	addFailures(t, pool, "login", "198.51.100.1/32", 1, 16*time.Minute)
	addFailures(t, pool, "login", "198.51.100.2/32", 1, 14*time.Minute)
	addFailures(t, pool, "reset", "198.51.100.3/32", 1, 16*time.Minute)

	if err := failures.New(pool, "login").Record(t.Context(), netip.MustParsePrefix("2001:db8:1:2::/64"), window); err != nil {
		t.Fatal(err)
	}

	var left string
	if err := pool.QueryRow(t.Context(), "SELECT string_agg(scope || ' ' || network::text, ', ' ORDER BY scope, network) FROM auth_failures").Scan(&left); err != nil {
		t.Fatal(err)
	}

	if left != "login 198.51.100.2/32, login 2001:db8:1:2::/64, reset 198.51.100.3/32" {
		t.Errorf("left %s", left)
	}
}

func addFailures(t *testing.T, pool *pgxpool.Pool, scope, network string, count int, age time.Duration) {
	t.Helper()

	query := "INSERT INTO auth_failures (scope, network, failed_at) SELECT $1, $2, now() - $3::interval FROM generate_series(1, $4)"
	if _, err := pool.Exec(t.Context(), query, scope, network, age, count); err != nil {
		t.Fatal(err)
	}
}
