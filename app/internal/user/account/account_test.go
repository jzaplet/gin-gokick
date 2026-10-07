package account_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/password"
	"gokick/app/core/testkit"
	"gokick/app/internal/user/account"
	"gokick/migrations"
)

const secret = "correct horse battery"

var cheap = password.Params{Memory: 64, Time: 1, Threads: 1}

func TestRegisterStoresTheEmailInLowerCaseAndAHash(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)

	id, err := account.New(pool, cheap).Register(t.Context(), "Jan@Example.com", secret, "cs_CZ")
	if err != nil {
		t.Fatal(err)
	}

	email, hash := stored(t, pool)
	if ok, rehash, err := cheap.Verify(secret, hash); id.Version() != 7 || email != "jan@example.com" || ok == false || rehash || err != nil {
		t.Errorf("id %s, email %s, hash %s: ok %t, rehash %t, error %v", id, email, hash, ok, rehash, err)
	}

	if testkit.Count(t, pool, "SELECT count(*) FROM users WHERE locale = 'cs_CZ'") != 1 {
		t.Error("the locale is not stored")
	}
}

func TestRegisterRefusesATakenEmail(t *testing.T) {
	accounts := account.New(testkit.Pool(t, migrations.FS), cheap)
	if _, err := accounts.Register(t.Context(), "jan@example.com", secret, "cs_CZ"); err != nil {
		t.Fatal(err)
	}

	if _, err := accounts.Register(t.Context(), "JAN@example.com", secret, "cs_CZ"); errors.Is(err, account.ErrEmailTaken) == false {
		t.Errorf("got %v", err)
	}
}

func TestAuthenticateFindsTheAccount(t *testing.T) {
	accounts := account.New(testkit.Pool(t, migrations.FS), cheap)

	registered, err := accounts.Register(t.Context(), "jan@example.com", secret, "en_US")
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()

	found, rehash, err := accounts.Authenticate(t.Context(), "JAN@example.com", secret)
	if want := (account.User{ID: registered, Locale: "en_US"}); err != nil || found != want || rehash {
		t.Errorf("found %+v, want %+v, rehash %t, error %v", found, want, rehash, err)
	}

	if time.Since(start) >= account.FailureFloor {
		t.Error("a login that succeeded waited for the floor of a failure")
	}
}

func TestAuthenticateRefusesWrongCredentialsNoSoonerThanTheFloor(t *testing.T) {
	accounts := account.New(testkit.Pool(t, migrations.FS), cheap)
	if _, err := accounts.Register(t.Context(), "jan@example.com", secret, "cs_CZ"); err != nil {
		t.Fatal(err)
	}

	for name, attempt := range map[string][2]string{
		"a wrong password": {"jan@example.com", secret + "!"},

		"an unknown email": {"eva@example.com", secret},
	} {
		start := time.Now()

		_, _, err := accounts.Authenticate(t.Context(), attempt[0], attempt[1])
		if elapsed := time.Since(start); errors.Is(err, account.ErrLoginFailed) == false || elapsed < account.FailureFloor {
			t.Errorf("%s: %v after %s", name, err, elapsed)
		}
	}
}

func TestAuthenticateAsksForARehashThatRehashStores(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	if _, err := account.New(pool, cheap).Register(t.Context(), "jan@example.com", secret, "cs_CZ"); err != nil {
		t.Fatal(err)
	}

	stronger := password.Params{Memory: 128, Time: 1, Threads: 1}
	accounts := account.New(pool, stronger)

	user, rehash, err := accounts.Authenticate(t.Context(), "jan@example.com", secret)
	if err != nil || rehash == false {
		t.Fatalf("rehash %t, error %v", rehash, err)
	}

	if _, hash := stored(t, pool); strings.HasPrefix(hash, "$argon2id$v=19$m=64,t=1,p=1$") == false {
		t.Errorf("the authentication itself changed the hash to %s", hash)
	}

	if err := accounts.Rehash(t.Context(), user.ID, secret); err != nil {
		t.Fatal(err)
	}

	_, hash := stored(t, pool)
	if ok, rehash, err := stronger.Verify(secret, hash); ok == false || rehash || err != nil || strings.HasPrefix(hash, "$argon2id$v=19$m=128,t=1,p=1$") == false {
		t.Errorf("hash %s: ok %t, rehash %t, error %v", hash, ok, rehash, err)
	}
}

func TestAuthenticateFailsOnABrokenHash(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	if _, err := pool.Exec(t.Context(), "INSERT INTO users (email, password_hash, locale) VALUES ('jan@example.com', 'broken', 'cs_CZ')"); err != nil {
		t.Fatal(err)
	}

	_, _, err := account.New(pool, cheap).Authenticate(t.Context(), "jan@example.com", secret)
	if err == nil || errors.Is(err, account.ErrLoginFailed) || strings.HasPrefix(err.Error(), "verify the password: ") == false {
		t.Errorf("got %v", err)
	}
}

func stored(t *testing.T, pool *pgxpool.Pool) (email, hash string) {
	t.Helper()

	if err := pool.QueryRow(t.Context(), "SELECT email, password_hash FROM users").Scan(&email, &hash); err != nil {
		t.Fatal(err)
	}

	return email, hash
}
