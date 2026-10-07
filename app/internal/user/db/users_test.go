package db_test

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"gokick/app/core/testkit"
	userdb "gokick/app/internal/user/db"
	"gokick/migrations"
)

func TestEmailsIgnoreCase(t *testing.T) {
	queries := newQueries(t)

	created, err := queries.CreateUser(t.Context(), userdb.CreateUserParams{
		Email: "Jan@Example.com",

		PasswordHash: "hash",

		Locale: "cs_CZ",
	})
	if err != nil {
		t.Fatal(err)
	}

	found, err := queries.UserByEmail(t.Context(), "JAN@example.COM")
	if err != nil {
		t.Fatal(err)
	}

	if created.Email != "jan@example.com" || found.ID != created.ID {
		t.Errorf("created %+v, found %+v", created, found)
	}
}

func TestUpdatePasswordHashChangesOnlyThatUser(t *testing.T) {
	queries := newQueries(t)

	jan, err := queries.CreateUser(t.Context(), userdb.CreateUserParams{
		Email: "jan@example.com",

		PasswordHash: "old",

		Locale: "cs_CZ",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := queries.CreateUser(t.Context(), userdb.CreateUserParams{
		Email: "eva@example.com",

		PasswordHash: "old",

		Locale: "cs_CZ",
	}); err != nil {
		t.Fatal(err)
	}

	if err := queries.UpdatePasswordHash(t.Context(), userdb.UpdatePasswordHashParams{
		ID: jan.ID,

		PasswordHash: "new",
	}); err != nil {
		t.Fatal(err)
	}

	for email, want := range map[string]string{"jan@example.com": "new", "eva@example.com": "old"} {
		if user, err := queries.UserByEmail(t.Context(), email); err != nil || user.PasswordHash != want {
			t.Errorf("%s: %+v %v", email, user, err)
		}
	}
}

func TestUpdateLocaleChangesOnlyThatUser(t *testing.T) {
	queries := newQueries(t)

	jan, err := queries.CreateUser(t.Context(), userdb.CreateUserParams{
		Email: "jan@example.com",

		PasswordHash: "hash",

		Locale: "cs_CZ",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := queries.CreateUser(t.Context(), userdb.CreateUserParams{
		Email: "eva@example.com",

		PasswordHash: "hash",

		Locale: "cs_CZ",
	}); err != nil {
		t.Fatal(err)
	}

	if err := queries.UpdateLocale(t.Context(), userdb.UpdateLocaleParams{
		ID: jan.ID,

		Locale: "en_US",
	}); err != nil {
		t.Fatal(err)
	}

	for email, want := range map[string]string{"jan@example.com": "en_US", "eva@example.com": "cs_CZ"} {
		if user, err := queries.UserByEmail(t.Context(), email); err != nil || user.Locale != want {
			t.Errorf("%s: %+v %v", email, user, err)
		}
	}
}

func TestDeleteUserDeletesOnlyThatUser(t *testing.T) {
	queries := newQueries(t)

	jan, err := queries.CreateUser(t.Context(), userdb.CreateUserParams{
		Email: "jan@example.com",

		PasswordHash: "hash",

		Locale: "cs_CZ",
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := queries.CreateUser(t.Context(), userdb.CreateUserParams{
		Email: "eva@example.com",

		PasswordHash: "hash",

		Locale: "cs_CZ",
	}); err != nil {
		t.Fatal(err)
	}

	if err := queries.DeleteUser(t.Context(), jan.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := queries.UserByEmail(t.Context(), "jan@example.com"); errors.Is(err, pgx.ErrNoRows) == false {
		t.Errorf("jan: %v", err)
	}

	if _, err := queries.UserByEmail(t.Context(), "eva@example.com"); err != nil {
		t.Errorf("eva: %v", err)
	}
}

func newQueries(t *testing.T) *userdb.Queries {
	t.Helper()

	return userdb.New(testkit.Pool(t, migrations.FS))
}
