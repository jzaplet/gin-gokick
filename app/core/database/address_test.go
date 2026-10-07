package database_test

import (
	"errors"
	"net/url"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/database"
)

func TestPgxReadsTheAddressBackWhole(t *testing.T) {
	address := database.Address{
		Host: "::1",

		Port: 54320,

		Name: "shop",

		Username: "shop",

		Password: "pass@word/word:word?word#word&word=word%",

		Params: url.Values{"sslmode": {"disable"}, "pool_max_conns": {"7"}},
	}

	cfg, err := pgxpool.ParseConfig(address.URL())
	if err != nil {
		t.Fatal(err)
	}

	conn := cfg.ConnConfig
	if conn.Host != "::1" || conn.Port != 54320 || conn.Database != "shop" || conn.User != "shop" || conn.Password != address.Password || conn.TLSConfig != nil || cfg.MaxConns != 7 {
		t.Errorf("pgx read %s:%d/%s as %s with %q, TLS %v, %d conns", conn.Host, conn.Port, conn.Database, conn.User, conn.Password, conn.TLSConfig, cfg.MaxConns)
	}
}

func TestParseParams(t *testing.T) {
	for value, want := range map[string]url.Values{
		"": {},

		"?sslmode=disable&pool_max_conns=10": {"sslmode": {"disable"}, "pool_max_conns": {"10"}},
	} {
		if got, err := database.ParseParams(value); err != nil || reflect.DeepEqual(got, want) == false {
			t.Errorf("%q: got %v, %v", value, got, err)
		}
	}

	for _, value := range []string{"sslmode=disable", "?sslmode=%zz", "?sslmode=disable;pool_max_conns=10"} {
		if _, err := database.ParseParams(value); errors.Is(err, database.ErrInvalidParams) == false {
			t.Errorf("%q: got %v", value, err)
		}
	}
}
