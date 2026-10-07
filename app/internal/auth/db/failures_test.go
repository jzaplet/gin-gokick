package db_test

import (
	"net/netip"
	"testing"

	"gokick/app/core/testkit"
	authdb "gokick/app/internal/auth/db"
	"gokick/migrations"
)

func TestRecordFailureRefusesHostBits(t *testing.T) {
	params := authdb.RecordFailureParams{Scope: "login", Network: netip.MustParsePrefix("2001:db8:1:2::abcd/64")}
	if err := authdb.New(testkit.Pool(t, migrations.FS)).RecordFailure(t.Context(), params); err == nil {
		t.Error("a network with host bits was stored")
	}
}
