package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authdb "gokick/app/internal/auth/db"
	"gokick/app/internal/auth/session/clientinfo"
)

const CookieName = "__Host-session"

const Lifetime = 30 * 24 * time.Hour

const touchEvery = 5 * time.Minute

type Sessions struct {
	queries *authdb.Queries
}

func New(db authdb.DBTX) *Sessions {
	return &Sessions{queries: authdb.New(db)}
}

func (s *Sessions) Start(ctx context.Context, userID uuid.UUID, client clientinfo.Client, previous string) (string, error) {
	if err := s.queries.DeleteExpiredSessions(ctx); err != nil {
		return "", fmt.Errorf("delete expired sessions: %w", err)
	}

	token := rand.Text()

	params := authdb.CreateSessionParams{
		TokenHash: tokenHash(token),

		UserID: userID,

		ExpiresAt: time.Now().Add(Lifetime),

		UserAgent: clientinfo.UserAgent(client.UserAgent),

		LastSeenIp: clientinfo.Address(client.IP),
	}
	if _, err := s.queries.CreateSession(ctx, params); err != nil {
		return "", fmt.Errorf("create a session: %w", err)
	}

	if previous != "" {
		if err := s.End(ctx, previous); err != nil {
			return "", err
		}
	}

	return token, nil
}

func (s *Sessions) User(ctx context.Context, token, ip string) (uuid.UUID, bool, error) {
	found, err := s.queries.SessionByTokenHash(ctx, tokenHash(token))
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}

	if err != nil {
		return uuid.Nil, false, fmt.Errorf("find the session: %w", err)
	}

	seenFrom := clientinfo.Address(ip)
	if time.Since(found.LastSeenAt) < touchEvery && clientinfo.SameAddress(found.LastSeenIp, seenFrom) {
		return found.UserID, true, nil
	}

	if err := s.queries.TouchSession(ctx, authdb.TouchSessionParams{ID: found.ID, LastSeenIp: seenFrom}); err != nil {
		return uuid.Nil, false, fmt.Errorf("mark the session as seen: %w", err)
	}

	return found.UserID, true, nil
}

func (s *Sessions) End(ctx context.Context, token string) error {
	if err := s.queries.DeleteSession(ctx, tokenHash(token)); err != nil {
		return fmt.Errorf("delete the session: %w", err)
	}

	return nil
}

func (s *Sessions) EndOthers(ctx context.Context, userID uuid.UUID, token string, ids []uuid.UUID) (int64, error) {
	params := authdb.EndUserSessionsParams{UserID: userID, Ids: ids, CurrentHash: tokenHash(token)}

	ended, err := s.queries.EndUserSessions(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("end the chosen sessions: %w", err)
	}

	return ended, nil
}

func (s *Sessions) EndAllOthers(ctx context.Context, userID uuid.UUID, token string, filter Filter) (int64, error) {
	params := authdb.EndOtherUserSessionsParams{UserID: userID, CurrentHash: tokenHash(token), Ip: filter.IP}

	ended, err := s.queries.EndOtherUserSessions(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("end the other sessions: %w", err)
	}

	return ended, nil
}

func tokenHash(token string) []byte {
	hash := sha256.Sum256([]byte(token))

	return hash[:]
}
