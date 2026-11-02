package suite

import (
	"context"
	"testing"

	"github.com/FacileStudio/porte"
)

type fakeSessions struct {
	porte.SessionStore
	user int64
}

func (f fakeSessions) Find(context.Context, string) (porte.Session, error) {
	return porte.Session{UserID: f.user}, nil
}
func (fakeSessions) Delete(context.Context, string) error               { return nil }
func (fakeSessions) DeleteByUser(context.Context, int64) (int64, error) { return 1, nil }

func TestRevokingSessionsForgetsCachedTokens(t *testing.T) {
	w := newWorld(t)
	w.addUser(1, future)
	tokens := w.tokens()
	store := ForgetOnRevoke(fakeSessions{user: 1}, tokens)
	ctx := context.Background()

	_, _ = tokens.For(ctx, 1, AppCourrier)
	_ = store.Delete(ctx, "hash")
	_, _ = tokens.For(ctx, 1, AppCourrier)
	_, _ = store.DeleteByUser(ctx, 1)
	_, _ = tokens.For(ctx, 1, AppCourrier)
	if got := w.count(&w.exchanges); got != 3 {
		t.Fatalf("logout and back-channel logout must each clear the cache, exchanges = %d", got)
	}
}
