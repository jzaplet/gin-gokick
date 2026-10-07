package failures

import (
	"context"
	"fmt"
	"net/netip"
	"time"

	authdb "gokick/app/internal/auth/db"
)

type Store struct {
	queries *authdb.Queries

	scope string
}

func New(db authdb.DBTX, scope string) *Store {
	return &Store{queries: authdb.New(db), scope: scope}
}

func (s *Store) Count(ctx context.Context, network netip.Prefix, window time.Duration) (int64, error) {
	failures, err := s.queries.CountFailures(ctx, authdb.CountFailuresParams{
		Scope: s.scope,

		Network: network,

		Period: window,
	})
	if err != nil {
		return 0, fmt.Errorf("count failures: %w", err)
	}

	return failures, nil
}

func (s *Store) Record(ctx context.Context, network netip.Prefix, window time.Duration) error {
	if err := s.queries.DeleteOldFailures(ctx, authdb.DeleteOldFailuresParams{
		Scope: s.scope,

		Period: window,
	}); err != nil {
		return fmt.Errorf("delete old failures: %w", err)
	}

	if err := s.queries.RecordFailure(ctx, authdb.RecordFailureParams{Scope: s.scope, Network: network}); err != nil {
		return fmt.Errorf("record a failure: %w", err)
	}

	return nil
}
