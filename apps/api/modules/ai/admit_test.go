package ai_test

import (
	"context"
	"crypto/sha256"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/modules/ai"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/go-chi/chi/v5"
)

const freeMember = "free@facile.studio"

type twoUsers struct{}

func (twoUsers) IdentityForEmail(_ context.Context, email string) (authcontext.Identity, error) {
	switch email {
	case member:
		return authcontext.Identity{UserID: 7, Email: email}, nil
	case freeMember:
		return authcontext.Identity{UserID: 8, Email: email}, nil
	}
	return authcontext.Identity{}, troncerrors.NotFound("user not found")
}

// paidMember pays for user 7 only.
type paidMember struct{}

func (paidMember) CheckAIIn(_ context.Context, userID int64, _ string) error {
	if userID == 7 {
		return nil
	}
	return troncerrors.New("payment_required", "upgrade to use AI", nil)
}

func TestAFreeUserNeverSpendsAPaidUsersDecideBudget(t *testing.T) {
	var calls atomic.Int32
	sum := sha256.Sum256([]byte(serviceKey))
	router := chi.NewRouter()
	ai.RegisterRoutes(router, ai.Config{
		Decider: stubJev(t, &calls), Callers: map[string][]byte{"sablier": sum[:]}, Users: twoUsers{}, Access: paidMember{},
		Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	srv := httptest.NewServer(router)
	t.Cleanup(srv.Close)
	for range 40 {
		if res := post(t, srv, map[string]string{"Authorization": "Bearer " + serviceKey, "X-Bulle-Acting-For": freeMember}, question); res.StatusCode != http.StatusPaymentRequired {
			t.Fatalf("a free user gets 402 before any limit: %d", res.StatusCode)
		}
	}
	for i := range 30 {
		if res := post(t, srv, map[string]string{"Authorization": "Bearer " + serviceKey, "X-Bulle-Acting-For": member}, question); res.StatusCode != http.StatusOK {
			t.Fatalf("call %d of the paying user: %d", i, res.StatusCode)
		}
	}
}

// paidSpace pays for nobody but the members of sp-paid.
type paidSpace struct{}

func (paidSpace) CheckAIIn(_ context.Context, _ int64, spaceID string) error {
	if spaceID == "sp-paid" {
		return nil
	}
	return troncerrors.New("payment_required", "upgrade to use AI", nil)
}

func TestCompleteGrantsThePlanOfTheSpaceItIsFor(t *testing.T) {
	writer := &fakeCompleter{answer: "ok"}
	srv := completeServer(t, writer, paidSpace{})
	if res, raw := complete(t, srv, toileWrite, `{"prompt":"Summarize."}`); res.StatusCode != http.StatusPaymentRequired {
		t.Fatalf("no space and no personal plan: status %d: %s", res.StatusCode, raw)
	}
	if res, raw := complete(t, srv, toileWrite, `{"prompt":"Summarize.","space_id":" sp-paid "}`); res.StatusCode != http.StatusOK || writer.calls() != 1 {
		t.Fatalf("a paid space: status %d: %s", res.StatusCode, raw)
	}
}
