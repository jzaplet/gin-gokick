package profile

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	userdb "gokick/app/internal/user/db"
)

type Profiles struct {
	queries *userdb.Queries
}

func New(db userdb.DBTX) *Profiles {
	return &Profiles{queries: userdb.New(db)}
}

func (p *Profiles) Email(ctx context.Context, id uuid.UUID) (string, error) {
	email, err := p.queries.EmailByID(ctx, id)
	if err != nil {
		return "", fmt.Errorf("find the email of the user: %w", err)
	}

	return email, nil
}

func (p *Profiles) SetLocale(ctx context.Context, id uuid.UUID, locale string) error {
	if err := p.queries.UpdateLocale(ctx, userdb.UpdateLocaleParams{ID: id, Locale: locale}); err != nil {
		return fmt.Errorf("store the locale of the user: %w", err)
	}

	return nil
}
