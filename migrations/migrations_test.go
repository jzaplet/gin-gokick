package migrations_test

import (
	"testing"

	"gokick/app/core/testkit"
	"gokick/migrations"
)

func TestUsersTable(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)

	var id string
	if err := pool.QueryRow(t.Context(), "INSERT INTO users (email, password_hash, locale) VALUES ('jan@example.com', 'hash', 'cs_CZ') RETURNING id").Scan(&id); err != nil {
		t.Fatal(err)
	}

	if len(id) != 36 || id[14] != '7' {
		t.Errorf("id %q is not a UUIDv7", id)
	}

	if _, err := pool.Exec(t.Context(), "INSERT INTO users (email, password_hash, locale) VALUES ('jan@example.com', 'hash', 'cs_CZ')"); err == nil {
		t.Error("the same e-mail was stored twice")
	}

	if _, err := pool.Exec(t.Context(), "INSERT INTO users (email, password_hash, locale) VALUES ('Jan@example.com', 'hash', 'cs_CZ')"); err == nil {
		t.Error("an e-mail with capitals was stored")
	}

	if _, err := pool.Exec(t.Context(), "INSERT INTO users (email, password_hash, locale) VALUES ('eva@example.com', 'hash', 'cs')"); err == nil {
		t.Error("a locale other than xx_YY was stored")
	}

	if _, err := pool.Exec(t.Context(), "INSERT INTO users (email, password_hash) VALUES ('ida@example.com', 'hash')"); err == nil {
		t.Error("a user without a locale was stored")
	}
}
