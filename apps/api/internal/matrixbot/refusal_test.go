package matrixbot

import (
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

func TestRefusalCarriesUpgradeLink(t *testing.T) {
	err := troncerrors.New("payment_required", "AI features need a paid plan.", plans.ErrAIAccess).
		WithUpgradeURL("https://sso.example.test/spaces")
	if got, want := refusal(err), "AI features need a paid plan. https://sso.example.test/spaces"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
