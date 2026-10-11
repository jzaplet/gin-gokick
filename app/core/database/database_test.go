package database_test

import (
	"context"
	"io"
	"net"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"gokick/app/core/database"
	"gokick/app/core/testkit"
)

var migrations = fstest.MapFS{
	"20260101000000_create_notes.sql": {Data: []byte("-- +goose Up\nCREATE TABLE notes (id int PRIMARY KEY);\n\n-- +goose Down\nDROP TABLE notes;\n")},

	"20260102000000_add_body.sql": {Data: []byte("-- +goose Up\nALTER TABLE notes ADD COLUMN body text;\n\n-- +goose Down\nALTER TABLE notes DROP COLUMN body;\n")},
}

func TestMigrateAppliesNewMigrationsOnce(t *testing.T) {
	pool, err := database.Open(t.Context(), testkit.Database(t), 0, nil)
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
	pool, err := database.Open(t.Context(), testkit.Database(t), 0, nil)
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

func TestOpenWaitsForTheDatabase(t *testing.T) {
	t.Run("connects once the database starts", func(t *testing.T) {
		dsn, address, target := laterAddress(t)
		opened := make(chan error, 1)

		go func() {
			pool, err := database.Open(t.Context(), dsn, 10*time.Second, nil)
			if err == nil {
				pool.Close()
			}

			opened <- err
		}()

		time.Sleep(1500 * time.Millisecond)
		forward(t, address, target)

		if err := <-opened; err != nil {
			t.Errorf("got %v", err)
		}
	})

	t.Run("gives up after the wait", func(t *testing.T) {
		dsn, _, _ := laterAddress(t)
		started := time.Now()

		if _, err := database.Open(t.Context(), dsn, time.Second, nil); err == nil || time.Since(started) < time.Second {
			t.Errorf("gave up after %v with %v", time.Since(started), err)
		}
	})
}

func laterAddress(t *testing.T) (dsn, address, target string) {
	t.Helper()

	parsed, err := url.Parse(testkit.Database(t))
	if err != nil {
		t.Fatal(err)
	}

	var config net.ListenConfig

	listener, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	address = listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	target = parsed.Host
	parsed.Host = address

	return parsed.String(), address, target
}

func forward(t *testing.T, address, target string) {
	t.Helper()

	var config net.ListenConfig

	listener, err := config.Listen(t.Context(), "tcp", address)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = listener.Close() })

	go func() {
		var dialer net.Dialer

		for {
			client, err := listener.Accept()
			if err != nil {
				return
			}

			server, err := dialer.DialContext(context.Background(), "tcp", target)
			if err != nil {
				_ = client.Close()

				continue
			}

			go pipe(client, server)
			go pipe(server, client)
		}
	}()
}

func pipe(to, from net.Conn) {
	_, _ = io.Copy(to, from)
	_ = to.Close()
}
