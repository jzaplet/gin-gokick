package db_test

import (
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/testkit"
	authdb "gokick/app/internal/auth/db"
	userdb "gokick/app/internal/user/db"
	"gokick/migrations"
)

const checkViolation = "23514"

func TestDeletingAUserEndsTheirSessions(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	userID := newUser(t, pool)
	newSession(t, authdb.New(pool), userID, "live", time.Hour)

	if _, err := pool.Exec(t.Context(), "DELETE FROM users WHERE id = $1", userID); err != nil {
		t.Fatal(err)
	}

	if count := testkit.Count(t, pool, "SELECT count(*) FROM sessions"); count != 0 {
		t.Errorf("%d sessions left", count)
	}
}

func TestCreateSessionRefusesAHashOfTheWrongLength(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	params := authdb.CreateSessionParams{
		TokenHash: tokenHash("short")[:31],

		UserID: newUser(t, pool),

		ExpiresAt: time.Now().Add(time.Hour),
	}

	_, err := authdb.New(pool).CreateSession(t.Context(), params)
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok == false || pgErr.Code != checkViolation {
		t.Errorf("got %v", err)
	}
}

func newUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()

	user, err := userdb.New(pool).CreateUser(t.Context(), userdb.CreateUserParams{
		Email: "jan@example.com",

		PasswordHash: "hash",

		Locale: "cs_CZ",
	})
	if err != nil {
		t.Fatal(err)
	}

	return user.ID
}

func newSession(t *testing.T, queries *authdb.Queries, userID uuid.UUID, token string, lifetime time.Duration) authdb.Session {
	t.Helper()

	params := authdb.CreateSessionParams{
		TokenHash: tokenHash(token),

		UserID: userID,

		ExpiresAt: time.Now().Add(lifetime),
	}

	session, err := queries.CreateSession(t.Context(), params)
	if err != nil {
		t.Fatal(err)
	}

	return session
}

func tokenHash(token string) []byte {
	hash := sha256.Sum256([]byte(token))

	return hash[:]
}
