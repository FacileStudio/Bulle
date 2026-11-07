package suite

import (
	"context"
	"fmt"

	"github.com/FacileStudio/porte"
)

func (t *Tokens) subject(ctx context.Context, userID int64, force bool) (string, error) {
	flight := fmt.Sprintf("subject/%d/%t", userID, force)
	value, err := t.flights.do(flight, func() (string, error) {
		return t.loadSubject(ctx, userID, force)
	})
	if err != nil {
		return "", err
	}
	return value, nil
}

func (t *Tokens) loadSubject(ctx context.Context, userID int64, force bool) (string, error) {
	identity, err := t.identity(ctx, userID)
	if err != nil {
		return "", err
	}
	tokens := identity.Tokens
	if !force && tokens.AccessToken != "" && t.now().Add(cacheMargin).Before(tokens.Expiry) {
		return tokens.AccessToken, nil
	}
	if tokens.RefreshToken == "" {
		return "", ErrNotConnected
	}
	refreshed, err := t.reg.Refresh(ctx, tokens.RefreshToken)
	if err != nil {
		return "", err
	}
	identity.Tokens = refreshed
	if err := t.ids.Save(ctx, identity); err != nil {
		return "", fmt.Errorf("suite: saving the refreshed token: %w", err)
	}
	return refreshed.AccessToken, nil
}

func (t *Tokens) identity(ctx context.Context, userID int64) (porte.StoredIdentity, error) {
	stored, err := t.ids.ListByUser(ctx, userID)
	if err != nil {
		return porte.StoredIdentity{}, fmt.Errorf("suite: reading identities: %w", err)
	}
	for _, identity := range stored {
		if identity.Provider == t.cfg.Issuer && identity.PasswordHash == "" {
			return identity, nil
		}
	}
	return porte.StoredIdentity{}, ErrNotConnected
}
