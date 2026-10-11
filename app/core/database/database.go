package database

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"gokick/app/core/logging"
)

const (
	pingTimeout = 5 * time.Second

	retryDelay = time.Second
)

func Open(ctx context.Context, dsn string, wait time.Duration, tracer pgx.QueryTracer) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	cfg.ConnConfig.Tracer = tracer

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("database: %w", err)
	}

	deadline := time.Now().Add(wait)

	for err = ping(ctx, pool); err != nil; err = ping(ctx, pool) {
		if ctx.Err() != nil || time.Now().After(deadline) {
			pool.Close()

			return nil, fmt.Errorf("database: %w", err)
		}

		time.Sleep(retryDelay)
	}

	return pool, nil
}

func ping(ctx context.Context, pool *pgxpool.Pool) error {
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	return pool.Ping(pingCtx)
}

func Migrate(ctx context.Context, pool *pgxpool.Pool, migrations fs.FS, logger *slog.Logger) (err error) {
	db := stdlib.OpenDBFromPool(pool)
	defer func() { err = errors.Join(err, db.Close()) }()

	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations, goose.WithSessionLocker(locker), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	results, err := provider.Up(ctx)
	for _, result := range results {
		if result.Error == nil {
			logger.LogAttrs(ctx, slog.LevelInfo, "migration applied", slog.String(logging.KeyMigration, result.Source.Path), logging.DurationMs(result.Duration))
		}
	}

	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	return nil
}
