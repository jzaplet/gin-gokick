package postgres

import (
	"errors"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"gokick/app/core/database"
)

func env(t *testing.T) {
	t.Helper()
	t.Setenv("DB_HOST", "db.gokick.local")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_NAME", "gokick")
	t.Setenv("DB_USERNAME", "app")
	t.Setenv("DB_PASSWORD", "secret")
	t.Setenv("DB_PARAMS", "?sslmode=disable&pool_max_conns=10")
}

func TestReadsTheParts(t *testing.T) {
	env(t)

	got, err := Read()

	want := database.Address{
		Host: "db.gokick.local",

		Port: 5432,

		Name: "gokick",

		Username: "app",

		Password: "secret",

		Params: url.Values{"sslmode": {"disable"}, "pool_max_conns": {"10"}},
	}
	if err != nil || reflect.DeepEqual(got, want) == false {
		t.Errorf("got %+v, %v", got, err)
	}
}

func TestRequiresTheParts(t *testing.T) {
	for _, key := range []string{"DB_HOST", "DB_NAME", "DB_USERNAME", "DB_PASSWORD"} {
		t.Run(key, func(t *testing.T) {
			env(t)
			t.Setenv(key, "")

			if _, err := Read(); errors.Is(err, ErrMissing) == false || strings.HasPrefix(err.Error(), key+" ") == false {
				t.Errorf("got %v", err)
			}
		})
	}
}

func TestRejectsAnInvalidPortOrParams(t *testing.T) {
	for key, value := range map[string]string{"DB_PORT": "0", "DB_PARAMS": "sslmode=disable"} {
		t.Run(key, func(t *testing.T) {
			env(t)
			t.Setenv(key, value)

			if _, err := Read(); err == nil || strings.HasPrefix(err.Error(), key+" ") == false {
				t.Errorf("got %v", err)
			}
		})
	}
}
