package suite

import (
	"context"

	"github.com/FacileStudio/porte"
)

type forgetOnRevoke struct {
	porte.SessionStore
	tokens *Tokens
}

func ForgetOnRevoke(store porte.SessionStore, tokens *Tokens) porte.SessionStore {
	return forgetOnRevoke{SessionStore: store, tokens: tokens}
}

func (f forgetOnRevoke) Delete(ctx context.Context, tokenHash string) error {
	if session, err := f.Find(ctx, tokenHash); err == nil {
		f.tokens.Forget(session.UserID)
	}
	return f.SessionStore.Delete(ctx, tokenHash)
}

func (f forgetOnRevoke) DeleteByUser(ctx context.Context, userID int64) (int64, error) {
	f.tokens.Forget(userID)
	return f.SessionStore.DeleteByUser(ctx, userID)
}
