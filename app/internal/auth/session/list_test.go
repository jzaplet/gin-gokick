package session_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"gokick/app/core/listing"
	"gokick/app/core/testkit"
	"gokick/app/internal/auth/session"
	"gokick/migrations"
)

func TestListShowsTheLiveSessionsOfTheUserAndMarksTheCurrentOne(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	jan := newUser(t, pool)
	current := addSession(t, pool, jan, "current", time.Hour)
	phone := addSession(t, pool, jan, "phone", time.Hour)
	addSession(t, pool, jan, "expired", -time.Minute)
	addSession(t, pool, newUserWith(t, pool, "eva@example.com"), "eva", time.Hour)

	query := listing.Query[session.Sort]{
		Page: 1,

		PerPage: 25,

		Sort: session.SortCreatedAt,

		Direction: listing.Ascending,
	}

	found, err := session.New(pool).List(t.Context(), jan, "current", query, session.Filter{})

	listed := found.Sessions
	if err != nil || found.Total != 2 || found.Others != 1 || len(listed) != 2 || listed[0].ID != current || listed[0].IsCurrent == false || listed[1].ID != phone || listed[1].IsCurrent {
		t.Errorf("got %+v, %v", found, err)
	}
}

func TestListSortsAndPages(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	jan := newUser(t, pool)
	ids := map[uuid.UUID]string{}

	for token, times := range map[string][2]string{
		"a": {"3 hours", "2 hours"},

		"b": {"2 hours", "1 hour"},

		"c": {"1 hour", "3 hours"},
	} {
		id := addSession(t, pool, jan, token, time.Hour)

		ids[id] = token
		if _, err := pool.Exec(t.Context(), "UPDATE sessions SET created_at = now() - $2::interval, last_seen_at = now() - $3::interval WHERE id = $1", id, times[0], times[1]); err != nil {
			t.Fatal(err)
		}
	}

	for _, tc := range []struct {
		page, perPage int

		sort session.Sort

		direction listing.Direction

		want string
	}{
		{1, 3, session.SortCreatedAt, listing.Ascending, "abc"},
		{1, 3, session.SortCreatedAt, listing.Descending, "cba"},
		{1, 3, session.SortLastSeenAt, listing.Ascending, "cab"},
		{1, 3, session.SortLastSeenAt, listing.Descending, "bac"},
		{2, 2, session.SortCreatedAt, listing.Ascending, "c"},
		{3, 2, session.SortCreatedAt, listing.Ascending, ""},
	} {
		query := listing.Query[session.Sort]{Page: tc.page, PerPage: tc.perPage, Sort: tc.sort, Direction: tc.direction}
		found, err := session.New(pool).List(t.Context(), jan, "", query, session.Filter{})

		tokens := make([]string, len(found.Sessions))
		for i, listed := range found.Sessions {
			tokens[i] = ids[listed.ID]
		}

		if got := strings.Join(tokens, ""); got != tc.want || found.Total != 3 || err != nil {
			t.Errorf("%+v: got %q of %d, %v, want %q", query, got, found.Total, err, tc.want)
		}
	}
}

func TestListFiltersByAPartOfTheAddressOfTheLastRequest(t *testing.T) {
	pool := testkit.Pool(t, migrations.FS)
	jan := newUser(t, pool)
	ids := map[uuid.UUID]string{}

	for i, token := range []string{"a", "b", "c", "d"} {
		id := addSession(t, pool, jan, token, time.Hour)
		ids[id] = token

		ip := map[string]any{"a": "203.0.113.7", "b": "198.51.100.4", "c": "2001:db8::1", "d": nil}[token]
		if _, err := pool.Exec(t.Context(), "UPDATE sessions SET created_at = now() - make_interval(hours => $2), last_seen_ip = $3::inet WHERE id = $1", id, 4-i, ip); err != nil {
			t.Fatal(err)
		}
	}

	query := listing.Query[session.Sort]{
		Page: 1,

		PerPage: 25,

		Sort: session.SortCreatedAt,

		Direction: listing.Ascending,
	}

	for ip, want := range map[string]string{
		"": "abcd",

		"203.0": "a",

		"2001:DB8": "c",

		"1": "abc",

		"10.0.0": "",
	} {
		found, err := session.New(pool).List(t.Context(), jan, "", query, session.Filter{IP: ip})

		tokens := make([]string, len(found.Sessions))
		for i, listed := range found.Sessions {
			tokens[i] = ids[listed.ID]
		}

		if got := strings.Join(tokens, ""); got != want || found.Total != int64(len(want)) || err != nil {
			t.Errorf("%q: got %q of %d, %v, want %q", ip, got, found.Total, err, want)
		}
	}
}
