package ai_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	troncerrors "github.com/FacileStudio/tronc/errors"
)

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
	srv := completeServer(t, nil, paidMember{}, withDecider(stubJev(t, &calls)))
	for range 40 {
		if res := post(t, srv, bearer(freeToken), question); res.StatusCode != http.StatusPaymentRequired {
			t.Fatalf("a free user gets 402 before any limit: %d", res.StatusCode)
		}
	}
	for i := range 30 {
		if res := post(t, srv, bearer(toileWrite), question); res.StatusCode != http.StatusOK {
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
