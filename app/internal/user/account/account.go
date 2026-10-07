package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"gokick/app/core/password"
	userdb "gokick/app/internal/user/db"
)

var ErrEmailTaken = errors.New("an account with this email exists")

var ErrLoginFailed = errors.New("the email or the password is wrong")

const uniqueViolation = "23505"

const FailureFloor = 300 * time.Millisecond

type User struct {
	ID uuid.UUID

	Locale string
}

type Accounts struct {
	queries *userdb.Queries

	params password.Params
}

func New(db userdb.DBTX, params password.Params) *Accounts {
	return &Accounts{queries: userdb.New(db), params: params}
}

func (a *Accounts) Register(ctx context.Context, email, plain, locale string) (uuid.UUID, error) {
	hash, err := a.params.Hash(plain)
	if err != nil {
		return uuid.Nil, fmt.Errorf("hash the password: %w", err)
	}

	user, err := a.queries.CreateUser(ctx, userdb.CreateUserParams{Email: email, PasswordHash: hash, Locale: locale})
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
		return uuid.Nil, ErrEmailTaken
	}

	if err != nil {
		return uuid.Nil, fmt.Errorf("create the user: %w", err)
	}

	return user.ID, nil
}

func (a *Accounts) Authenticate(ctx context.Context, email, plain string) (user User, rehash bool, err error) {
	start := time.Now()

	user, rehash, err = a.authenticate(ctx, email, plain)
	if errors.Is(err, ErrLoginFailed) {
		waitUntil(ctx, start.Add(FailureFloor))
	}

	return user, rehash, err
}

func (a *Accounts) authenticate(ctx context.Context, email, plain string) (User, bool, error) {
	found, err := a.queries.UserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		if _, hashErr := a.params.Hash(plain); hashErr != nil {
			return User{}, false, fmt.Errorf("hash the password of an unknown email: %w", hashErr)
		}

		return User{}, false, ErrLoginFailed
	}

	if err != nil {
		return User{}, false, fmt.Errorf("find the user: %w", err)
	}

	ok, rehash, err := a.params.Verify(plain, found.PasswordHash)
	if err != nil {
		return User{}, false, fmt.Errorf("verify the password: %w", err)
	}

	if ok == false {
		return User{}, false, ErrLoginFailed
	}

	return User{ID: found.ID, Locale: found.Locale}, rehash, nil
}

func waitUntil(ctx context.Context, deadline time.Time) {
	timer := time.NewTimer(time.Until(deadline))
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-ctx.Done():
	}
}

func (a *Accounts) Rehash(ctx context.Context, id uuid.UUID, plain string) error {
	hash, err := a.params.Hash(plain)
	if err != nil {
		return fmt.Errorf("rehash the password: %w", err)
	}

	if err := a.queries.UpdatePasswordHash(ctx, userdb.UpdatePasswordHashParams{
		ID: id,

		PasswordHash: hash,
	}); err != nil {
		return fmt.Errorf("store the new hash: %w", err)
	}

	return nil
}

func (a *Accounts) Delete(ctx context.Context, id uuid.UUID) error {
	if err := a.queries.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("delete the user: %w", err)
	}

	return nil
}
