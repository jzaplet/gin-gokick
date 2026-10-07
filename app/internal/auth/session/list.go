package session

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"gokick/app/core/listing"
	authdb "gokick/app/internal/auth/db"
)

//tsgen:assets/app/Auth/types/SessionSort.ts SessionSort union
type Sort string

const SortCreatedAt Sort = "createdAt"

const SortLastSeenAt Sort = "lastSeenAt"

var Grid = listing.Grid[Sort]{Sorts: []Sort{SortLastSeenAt, SortCreatedAt}, Direction: listing.Descending, PerPage: 25}

type Filter struct {
	IP string
}

type Listed = authdb.ListUserSessionsRow

type Page struct {
	Sessions []Listed

	Total int64

	Others int64
}

func (s *Sessions) List(
	ctx context.Context,
	userID uuid.UUID,
	token string,
	query listing.Query[Sort],
	filter Filter,
) (Page, error) {
	listed, err := s.queries.ListUserSessions(ctx, authdb.ListUserSessionsParams{
		CurrentHash: tokenHash(token),

		UserID: userID,

		SortBy: string(query.Sort),

		Descending: query.Direction == listing.Descending,

		RowOffset: query.Offset(),

		RowLimit: query.Limit(),

		Ip: filter.IP,
	})
	if err != nil {
		return Page{}, fmt.Errorf("list the sessions: %w", err)
	}

	counted, err := s.queries.CountUserSessions(ctx, authdb.CountUserSessionsParams{
		CurrentHash: tokenHash(token),

		UserID: userID,

		Ip: filter.IP,
	})
	if err != nil {
		return Page{}, fmt.Errorf("count the sessions: %w", err)
	}

	return Page{Sessions: listed, Total: counted.Total, Others: counted.Others}, nil
}
