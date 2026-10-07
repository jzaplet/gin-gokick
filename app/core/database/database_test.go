package database_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"gokick/app/core/database"
	"gokick/app/core/testkit"
)

var migrations = fstest.MapFS{
	"20260101000000_create_notes.sql": {Data: []byte("-- +goose Up\nCREATE TABLE notes (id int PRIMARY KEY);\n\n-- +goose Down\nDROP TABLE notes;\n")},

	"20260102000000_add_body.sql": {Data: []byte("-- +goose Up\nALTER TABLE notes ADD COLUMN body text;\n\n-- +goose Down\nALTER TABLE notes DROP COLUMN body;\n")},
}

func TestMigrateAppliesNewMigrationsOnce(t *testing.T) {
	pool, err := database.Open(t.Context(), testkit.Database(t))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)

	logger, buf := testkit.CaptureLogs()

	if err := database.Migrate(t.Context(), pool, migrations, logger); err != nil {
		t.Fatal(err)
	}

	if err := database.Migrate(t.Context(), pool, migrations, logger); err != nil {
		t.Fatal(err)
	}

	entries := testkit.LogEntries(t, buf)
	if len(entries) != 2 || entries[0]["migration"] != "20260101000000_create_notes.sql" || entries[1]["migration"] != "20260102000000_add_body.sql" || entries[0]["duration_ms"] == nil {
		t.Errorf("log %v", entries)
	}

	if _, err := pool.Exec(t.Context(), "INSERT INTO notes (id, body) VALUES (1, 'x')"); err != nil {
		t.Errorf("the migrated table: %v", err)
	}
}

func TestMigrateStopsAtABrokenMigration(t *testing.T) {
	pool, err := database.Open(t.Context(), testkit.Database(t))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(pool.Close)

	broken := fstest.MapFS{"20260101000000_broken.sql": {Data: []byte("-- +goose Up\nCREATE TABLE;\n")}}

	if err := database.Migrate(t.Context(), pool, broken, testkit.DiscardLogs()); err == nil || strings.Contains(err.Error(), "20260101000000") == false {
		t.Errorf("got %v", err)
	}

	if err := database.Migrate(t.Context(), pool, fstest.MapFS{}, testkit.DiscardLogs()); err == nil {
		t.Error("migrated without migrations")
	}
}
