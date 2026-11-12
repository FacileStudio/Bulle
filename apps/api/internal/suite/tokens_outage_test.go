package suite

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FacileStudio/porte"
)

type outageRegistre struct {
	srv     *httptest.Server
	missing atomic.Int64
	refuse  atomic.Bool
	hits    atomic.Int64
}

func newOutageRegistre(t *testing.T, missing int64) *outageRegistre {
	t.Helper()
	waits := exchangeWaits
	exchangeWaits = []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond, time.Millisecond}
	t.Cleanup(func() { exchangeWaits = waits })
	r := &outageRegistre{}
	r.missing.Store(missing)
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(rw http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(rw).Encode(map[string]string{"token_endpoint": r.srv.URL + "/token"})
	})
	mux.HandleFunc("/token", func(rw http.ResponseWriter, _ *http.Request) {
		r.hits.Add(1)
		if r.missing.Add(-1) >= 0 {
			http.Error(rw, "404 page not found", http.StatusNotFound)
			return
		}
		if r.refuse.Load() {
			rw.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(rw).Encode(map[string]string{"error": "access_denied"})
			return
		}
		_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "minted", "expires_in": 600})
	})
	r.srv = httptest.NewServer(mux)
	t.Cleanup(r.srv.Close)
	return r
}

func (r *outageRegistre) ListByUser(_ context.Context, userID int64) ([]porte.StoredIdentity, error) {
	return []porte.StoredIdentity{{UserID: userID, Provider: r.srv.URL, Tokens: porte.TokenSet{AccessToken: "user", Expiry: time.Now().Add(time.Hour)}}}, nil
}

func (r *outageRegistre) Save(context.Context, porte.StoredIdentity) error { return nil }

func (r *outageRegistre) tokens(apps ...string) *Tokens {
	cfg := Config{Issuer: r.srv.URL, ClientID: "bulle", ClientSecret: "s3cret", MCP: map[string]MCPConfig{}}
	for _, app := range apps {
		cfg.MCP[app] = MCPConfig{URL: "https://" + app + ".test/mcp", Resource: "https://" + app + ".test/mcp"}
	}
	return NewTokens(cfg, r, nil)
}

func TestAnExchangeRidesOutARegistreRedeploy(t *testing.T) {
	registre := newOutageRegistre(t, 2)
	token, err := registre.tokens(AppAgenda).ForMCP(context.Background(), 1, AppAgenda)
	if err != nil || token != "minted" || registre.hits.Load() != 3 {
		t.Fatalf("two 404s with no OAuth error must be waited out: %q, %v, %d hits", token, err, registre.hits.Load())
	}
}

func TestARegistreThatStaysDownCostsOneSeriesOfAttempts(t *testing.T) {
	registre := newOutageRegistre(t, 1<<30)
	tokens := registre.tokens(AppAgenda, AppCourrier)
	_, err := tokens.ForMCP(context.Background(), 1, AppAgenda)
	if !transientExchange(err) || registre.hits.Load() != int64(len(exchangeWaits)+1) {
		t.Fatalf("a Registre that stays down must be tried %d times, then reported as transient: %v, %d hits", len(exchangeWaits)+1, err, registre.hits.Load())
	}
	if _, err := tokens.ForMCP(context.Background(), 1, AppCourrier); !transientExchange(err) || registre.hits.Load() != int64(len(exchangeWaits)+2) {
		t.Fatalf("the next app right after must try once, not wait again: %v, %d hits", err, registre.hits.Load())
	}
	registre.missing.Store(0)
	if token, err := tokens.ForMCP(context.Background(), 1, AppCourrier); err != nil || token != "minted" {
		t.Fatalf("a Registre back must answer at once: %q, %v", token, err)
	}
}

func TestParallelExchangesShareOneSeries(t *testing.T) {
	registre := newOutageRegistre(t, 1<<30)
	apps := []string{AppAgenda, AppCourrier, AppNuage, AppOpus, AppSablier, AppPlume, AppVision, AppEcho}
	tokens := registre.tokens(apps...)
	var wg sync.WaitGroup
	for i, app := range apps {
		wg.Go(func() {
			if _, err := tokens.ForMCP(context.Background(), int64(i+1), app); !transientExchange(err) {
				t.Errorf("%s: %v", app, err)
			}
		})
	}
	wg.Wait()
	if most := int64(len(exchangeWaits) + 2*len(apps)); registre.hits.Load() > most {
		t.Fatalf("parallel exchanges must share one series of attempts: %d hits, at most %d", registre.hits.Load(), most)
	}
}

func TestARealRefusalIsNotRetried(t *testing.T) {
	registre := newOutageRegistre(t, 0)
	registre.refuse.Store(true)
	_, err := registre.tokens(AppJournal).ForMCP(context.Background(), 1, AppJournal)
	if !refusedAccess(err) || transientExchange(err) || registre.hits.Load() != 1 {
		t.Fatalf("access_denied is final: %v, %d hits", err, registre.hits.Load())
	}
	if !errors.Is(absence(err), ErrNoAccess) {
		t.Fatalf("a refusal is no access, not an outage: %v", absence(err))
	}
}
