// Package plans enforces the plan a Registre person pays for. It feeds every
// directory snapshot to a porte/forfait Store, keeps the Store's state in one
// database row so a restart never fails open again, and answers whether a user
// may use Bulle's AI features.
package plans

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/forfait"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"gorm.io/gorm"
)

// AIAccess is the feature that turns Bulle's AI features on for a person.
const AIAccess = "ai.access"

// ErrAIAccess is the cause of every refusal the Gate gives.
var ErrAIAccess = errors.New("plans: ai.access is not part of this plan")

const refusalMessage = "AI features need a paid plan. Upgrade your plan to use the assistant."

// Gate is Bulle's view of the plans Registre serves. A nil Gate, or one that
// never received an entitlements snapshot, allows everything.
type Gate struct {
	orm    *gorm.DB
	issuer string
	logger *slog.Logger
	store  *forfait.Store
	active atomic.Bool
}

// New builds the Gate for the OIDC issuer Bulle signs users in with. The
// issuer is both the provider of the porte identities a person is read from
// and the origin of every upgrade link.
func New(orm *gorm.DB, issuer string, logger *slog.Logger) *Gate {
	store := forfait.NewStore(forfait.Feature{Key: AIAccess, Kind: forfait.Bool, Default: json.RawMessage("false")})
	store.Issuer = issuer
	store.Logger = logger
	return &Gate{orm: orm, issuer: issuer, logger: logger, store: store}
}

// Load restores the state the last Apply saved. Call it at boot, before the
// first pull. A row that cannot be imported is logged and the Gate fails
// closed at the feature defaults until the next snapshot, because a row means
// entitlements were authoritative once.
func (g *Gate) Load(ctx context.Context) error {
	var documents []string
	if err := g.orm.WithContext(ctx).Raw(`SELECT document FROM forfait_state WHERE id = 1`).
		Scan(&documents).Error; err != nil {
		return err
	}
	if len(documents) == 0 {
		return nil
	}
	if err := g.store.Import([]byte(documents[0])); err != nil {
		g.logger.Error("plans: saved entitlements could not be restored, enforcing defaults", slog.Any("error", err))
		g.store.Apply(directory.Snapshot{EntitlementsRevision: 1})
		g.active.Store(true)
		return nil
	}
	g.refresh([]byte(documents[0]))
	return nil
}

// Apply feeds one directory snapshot to the Store and saves the result. It is
// chained into the directory Syncer's Apply, so a failed save makes the
// Syncer pull and save again.
func (g *Gate) Apply(ctx context.Context, snapshot directory.Snapshot) error {
	g.store.Apply(snapshot)
	document, err := g.store.Export()
	if err != nil {
		return err
	}
	if err := g.orm.WithContext(ctx).Exec(`INSERT INTO forfait_state (id, document, updated_at)
		VALUES (1, ?, now())
		ON CONFLICT (id) DO UPDATE SET document = EXCLUDED.document, updated_at = now()`,
		string(document)).Error; err != nil {
		return err
	}
	g.refresh(document)
	return nil
}

// Active reports whether entitlements are authoritative: an entitlements
// snapshot was applied, now or before the last restart.
func (g *Gate) Active() bool {
	return g != nil && g.active.Load()
}

func (g *Gate) refresh(document []byte) {
	var state struct {
		Revision int64 `json:"entitlements_revision"`
	}
	if json.Unmarshal(document, &state) == nil && state.Revision > 0 {
		g.active.Store(true)
	}
}

// CheckAI answers nil when the user may use the AI features, and a 402
// envelope error (payment_required with an upgrade_url) when their plan lacks
// ai.access. The user is the person acting: the signed-in user, or the user a
// service call acts for. Every Registre identity the user holds is read and
// the most generous one wins; a user with none reads the free defaults. Before
// the first entitlements snapshot ever applied it answers nil.
func (g *Gate) CheckAI(ctx context.Context, userID int64) error {
	if !g.Active() {
		return nil
	}
	var subjects []string
	if err := g.orm.WithContext(ctx).Raw(`SELECT subject FROM porte_identities WHERE user_id = ? AND provider = ?`,
		userID, g.issuer).Scan(&subjects).Error; err != nil {
		return troncerrors.Internal("failed to read the user's plan", err)
	}
	if len(subjects) == 0 {
		subjects = []string{""}
	}
	for _, subject := range subjects {
		if g.store.Enabled(forfait.Scope{Subject: subject}, AIAccess) {
			return nil
		}
	}
	return troncerrors.New("payment_required", refusalMessage, ErrAIAccess).
		WithUpgradeURL(forfait.UpgradeURL(g.issuer, ""))
}

// Plan names the tier the user's entitlements put them on: "paid" when ai.access
// is on, "free" when it is off, and "" while entitlements are not authoritative
// or cannot be read.
func (g *Gate) Plan(ctx context.Context, userID int64) string {
	if !g.Active() {
		return ""
	}
	err := g.CheckAI(ctx, userID)
	if err == nil {
		return "paid"
	}
	if errors.Is(err, ErrAIAccess) {
		return "free"
	}
	return ""
}

// Require is middleware for routes that spend AI. It runs after
// authentication and answers the CheckAI refusal for the signed-in user, 401
// when nobody is signed in.
func (g *Gate) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := authcontext.IdentityFromContext(r.Context())
		if !ok {
			httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
			return
		}
		if err := g.CheckAI(r.Context(), identity.UserID); err != nil {
			httpjson.WriteError(w, err)
			return
		}
		next.ServeHTTP(w, r)
	})
}
