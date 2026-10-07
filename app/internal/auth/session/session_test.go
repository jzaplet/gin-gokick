package session_test

import (
	"crypto/sha256"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"gokick/app/core/listing"
	"gokick/app/core/testkit"
	authdb "gokick/app/internal/auth/db"
	"gokick/app/internal/auth/session"
	"gokick/app/internal/auth/session/clientinfo"
	userdb "gokick/app/internal/user/db"
	"gokick/migrations"
)

func TestStartStoresOnlyTheHashOfTheToken(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	userID := newUser(t, pool)

	token, err := session.New(pool).Start(t.Context(), userID, clientinfo.Client{}, "")
	if err != nil {
		t.Fatal(err)
	}

	var owner uuid.UUID

	var expires time.Time
	if err := pool.QueryRow(t.Context(), "SELECT user_id, expires_at FROM sessions WHERE token_hash = $1", tokenHash(token)).Scan(&owner, &expires); err != nil {
		t.Fatal(err)
	}

	if left := time.Until(expires); len(token) != 26 || owner != userID || left < session.Lifetime-time.Minute || left > session.Lifetime {
		t.Errorf("token %q of %s expires in %s", token, owner, left)
	}

	if strings.HasPrefix(session.CookieName, "__Host-") == false {
		t.Errorf("cookie %s", session.CookieName)
	}
}

func TestStartStoresTheAddressAndTheCleanedUserAgent(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	client := clientinfo.Client{UserAgent: "Mozilla\xff\x00 " + strings.Repeat("é", 600), IP: "::ffff:203.0.113.7"}

	token, err := session.New(pool).Start(t.Context(), newUser(t, pool), client, "")
	if err != nil {
		t.Fatal(err)
	}

	stored := "SELECT count(*) FROM sessions WHERE token_hash = $1 AND user_agent = $2 AND last_seen_ip = '203.0.113.7'"
	if testkit.Count(t, pool, stored, tokenHash(token), clientinfo.UserAgent(client.UserAgent)) != 1 {
		t.Error("the session did not store the address and the cleaned user agent")
	}
}

func TestStartDeletesExpiredSessions(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	userID := newUser(t, pool)
	addSession(t, pool, userID, "expired", -time.Minute)
	addSession(t, pool, userID, "live", time.Hour)

	if _, err := session.New(pool).Start(t.Context(), userID, clientinfo.Client{}, ""); err != nil {
		t.Fatal(err)
	}

	if n := testkit.Count(t, pool, "SELECT count(*) FROM sessions WHERE token_hash = $1", tokenHash("expired")); n != 0 || testkit.Count(t, pool, "SELECT count(*) FROM sessions") != 2 {
		t.Error("the start kept the expired session or deleted a live one")
	}
}

func TestStartEndsThePreviousSession(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	userID := newUser(t, pool)
	addSession(t, pool, userID, "previous", time.Hour)
	addSession(t, pool, userID, "another-device", time.Hour)

	token, err := session.New(pool).Start(t.Context(), userID, clientinfo.Client{}, "previous")
	if err != nil {
		t.Fatal(err)
	}

	byHash := "SELECT count(*) FROM sessions WHERE token_hash = $1"
	if testkit.Count(t, pool, byHash, tokenHash("previous")) != 0 || testkit.Count(t, pool, byHash, tokenHash(token)) != 1 || testkit.Count(t, pool, "SELECT count(*) FROM sessions") != 2 {
		t.Error("the start did not replace exactly the previous session")
	}
}

func TestStartKeepsThePreviousSessionWhenTheNewOneFails(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	userID := newUser(t, pool)
	addSession(t, pool, userID, "previous", time.Hour)

	if _, err := pool.Exec(t.Context(), "ALTER TABLE sessions ADD CONSTRAINT blocked CHECK (false) NOT VALID"); err != nil {
		t.Fatal(err)
	}

	if _, err := session.New(pool).Start(t.Context(), userID, clientinfo.Client{}, "previous"); err == nil {
		t.Fatal("a blocked session started")
	}

	if n := testkit.Count(t, pool, "SELECT count(*) FROM sessions WHERE token_hash = $1", tokenHash("previous")); n != 1 {
		t.Error("the failed start ended the previous session")
	}
}

func TestEndDeletesOnlyThatSession(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	userID := newUser(t, pool)
	addSession(t, pool, userID, "this-device", time.Hour)
	addSession(t, pool, userID, "another-device", time.Hour)

	if err := session.New(pool).End(t.Context(), "this-device"); err != nil {
		t.Fatal(err)
	}

	if n := testkit.Count(t, pool, "SELECT count(*) FROM sessions WHERE token_hash = $1", tokenHash("another-device")); n != 1 || testkit.Count(t, pool, "SELECT count(*) FROM sessions") != 1 {
		t.Error("the end did not delete exactly its own session")
	}
}

func TestUserFindsTheOwnerOfALiveSession(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	userID := newUser(t, pool)
	addSession(t, pool, userID, "live", time.Hour)

	owner, ok, err := session.New(pool).User(t.Context(), "live", "")
	if owner != userID || ok == false || err != nil {
		t.Errorf("got %s, %t, %v", owner, ok, err)
	}
}

func TestUserFindsNoExpiredOrUnknownSession(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	addSession(t, pool, newUser(t, pool), "expired", -time.Minute)

	for _, token := range []string{"expired", "unknown"} {
		if owner, ok, err := session.New(pool).User(t.Context(), token, ""); owner != uuid.Nil || ok || err != nil {
			t.Errorf("%s: got %s, %t, %v", token, owner, ok, err)
		}
	}
}

func TestUserFailsWhenTheDatabaseFails(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	pool.Close()

	if _, ok, err := session.New(pool).User(t.Context(), "live", ""); ok || err == nil {
		t.Errorf("got %t, %v", ok, err)
	}
}

func TestUserMarksTheSessionAsSeenEveryFiveMinutesOrFromANewAddress(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	userID := newUser(t, pool)
	addSession(t, pool, userID, "live", time.Hour)
	addSession(t, pool, userID, "expired", -time.Minute)

	seen := "SELECT count(*) FROM sessions WHERE token_hash = $1 AND last_seen_at > now() - interval '1 minute' AND last_seen_ip = $2"

	for _, tc := range []struct {
		seenAgo, storedIP, ip, wantIP string

		touched int
	}{
		{"4 minutes", "192.0.2.1", "192.0.2.1", "192.0.2.1", 0},
		{"6 minutes", "192.0.2.1", "192.0.2.1", "192.0.2.1", 1},
		{"4 minutes", "192.0.2.1", "::ffff:203.0.113.7", "203.0.113.7", 1},
		{"4 minutes", "192.0.2.1", "", "192.0.2.1", 0},
		{"6 minutes", "192.0.2.1", "", "192.0.2.1", 1},
	} {
		if _, err := pool.Exec(t.Context(), "UPDATE sessions SET last_seen_at = now() - $1::interval, last_seen_ip = $2", tc.seenAgo, tc.storedIP); err != nil {
			t.Fatal(err)
		}

		for _, token := range []string{"live", "expired"} {
			if _, _, err := session.New(pool).User(t.Context(), token, tc.ip); err != nil {
				t.Fatal(err)
			}
		}

		if live, expired := testkit.Count(t, pool, seen, tokenHash("live"), tc.wantIP), testkit.Count(t, pool, seen, tokenHash("expired"), tc.wantIP); live != tc.touched || expired != 0 {
			t.Errorf("%+v: %d live and %d expired sessions marked as seen", tc, live, expired)
		}
	}
}

func TestEndOthersEndsOnlyTheChosenSessionsOfTheUserButNeverTheCurrentOne(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	jan := newUser(t, pool)
	current := addSession(t, pool, jan, "current", time.Hour)
	phone := addSession(t, pool, jan, "phone", time.Hour)
	addSession(t, pool, jan, "tablet", time.Hour)
	eva := addSession(t, pool, newUserWith(t, pool, "eva@example.com"), "eva", time.Hour)

	ended, err := session.New(pool).EndOthers(t.Context(), jan, "current", []uuid.UUID{current, phone, eva})

	if ended != 1 || err != nil || remaining(t, pool, "current", "phone", "tablet", "eva") != "current tablet eva" {
		t.Errorf("ended %d, %v, left %s", ended, err, remaining(t, pool, "current", "phone", "tablet", "eva"))
	}
}

func TestEndAllOthersEndsEveryLiveSessionOfTheUserButTheCurrentOne(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)

	jan := newUser(t, pool)
	for _, token := range []string{"current", "phone", "tablet"} {
		addSession(t, pool, jan, token, time.Hour)
	}

	addSession(t, pool, jan, "expired", -time.Minute)
	addSession(t, pool, newUserWith(t, pool, "eva@example.com"), "eva", time.Hour)

	ended, err := session.New(pool).EndAllOthers(t.Context(), jan, "current", session.Filter{})

	if ended != 2 || err != nil || remaining(t, pool, "current", "phone", "tablet", "expired", "eva") != "current expired eva" {
		t.Errorf("ended %d, %v, left %s", ended, err, remaining(t, pool, "current", "phone", "tablet", "expired", "eva"))
	}
}

func TestEndAllOthersEndsOnlyTheSessionsTheFilterLists(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)

	jan := newUser(t, pool)
	for token, ip := range map[string]any{
		"current": "203.0.113.7",

		"phone": "203.0.113.9",

		"tablet": "198.51.100.4",

		"unknown": nil,
	} {
		seenFrom(t, pool, addSession(t, pool, jan, token, time.Hour), ip)
	}

	seenFrom(t, pool, addSession(t, pool, newUserWith(t, pool, "eva@example.com"), "eva", time.Hour), "203.0.113.1")

	filter := session.Filter{IP: "203.0"}
	query := listing.Query[session.Sort]{
		Page: 1,

		PerPage: 25,

		Sort: session.SortCreatedAt,

		Direction: listing.Ascending,
	}

	found, err := session.New(pool).List(t.Context(), jan, "current", query, filter)
	if found.Total != 2 || found.Others != 1 || err != nil {
		t.Fatalf("listed %d, %d others, %v", found.Total, found.Others, err)
	}

	ended, err := session.New(pool).EndAllOthers(t.Context(), jan, "current", filter)

	if ended != found.Others || err != nil || remaining(t, pool, "current", "phone", "tablet", "unknown", "eva") != "current tablet unknown eva" {
		t.Errorf("ended %d, %v, left %s", ended, err, remaining(t, pool, "current", "phone", "tablet", "unknown", "eva"))
	}
}

func seenFrom(t *testing.T, pool *pgxpool.Pool, id uuid.UUID, ip any) {
	t.Helper()

	if _, err := pool.Exec(t.Context(), "UPDATE sessions SET last_seen_ip = $2::inet WHERE id = $1", id, ip); err != nil {
		t.Fatal(err)
	}
}

func remaining(t *testing.T, pool *pgxpool.Pool, tokens ...string) string {
	t.Helper()
	var left []string

	for _, token := range tokens {
		if testkit.Count(t, pool, "SELECT count(*) FROM sessions WHERE token_hash = $1", tokenHash(token)) == 1 {
			left = append(left, token)
		}
	}

	return strings.Join(left, " ")
}

func newUser(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	t.Helper()

	return newUserWith(t, pool, "jan@example.com")
}

func newUserWith(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()

	user, err := userdb.New(pool).CreateUser(t.Context(), userdb.CreateUserParams{
		Email: email,

		PasswordHash: "hash",

		Locale: "cs_CZ",
	})
	if err != nil {
		t.Fatal(err)
	}

	return user.ID
}

func addSession(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID, token string, lifetime time.Duration) uuid.UUID {
	t.Helper()

	params := authdb.CreateSessionParams{
		TokenHash: tokenHash(token),

		UserID: userID,

		ExpiresAt: time.Now().Add(lifetime),
	}

	created, err := authdb.New(pool).CreateSession(t.Context(), params)
	if err != nil {
		t.Fatal(err)
	}

	return created.ID
}

func tokenHash(token string) []byte {
	hash := sha256.Sum256([]byte(token))

	return hash[:]
}
