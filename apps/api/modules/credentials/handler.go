package credentials

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/internal/credentials"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

const maxKeyBytes = 8 << 10

var vendorURLs = map[string]string{
	"anthropic":  "https://api.anthropic.com/v1/models?limit=1",
	"openai":     "https://api.openai.com/v1/models",
	"google":     "https://generativelanguage.googleapis.com/v1beta/models?pageSize=1",
	"openrouter": "https://openrouter.ai/api/v1/auth/key",
}

type Handler struct {
	store  credentials.Store
	client *http.Client
	urls   map[string]string
}

func NewHandler(store credentials.Store, urls map[string]string) *Handler {
	if urls == nil {
		urls = vendorURLs
	}
	return &Handler{store: store, client: &http.Client{Timeout: 15 * time.Second}, urls: urls}
}

func (h *Handler) userID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	identity, ok := authcontext.IdentityFromContext(r.Context())
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("not authenticated"))
		return 0, false
	}
	return identity.UserID, true
}

func (h *Handler) backend(w http.ResponseWriter, r *http.Request) (string, bool) {
	backend := chi.URLParam(r, "backend")
	if !credentials.Valid(backend) {
		httpjson.WriteError(w, troncerrors.Invalid("unknown backend"))
		return "", false
	}
	return backend, true
}

func (h *Handler) responses(ctx context.Context, userID int64) ([]CredentialResponse, error) {
	entries, err := h.store.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	byBackend := make(map[string]credentials.Entry, len(entries))
	for _, entry := range entries {
		byBackend[entry.Backend] = entry
	}
	names := credentials.Backends()
	out := make([]CredentialResponse, 0, len(names))
	for _, backend := range names {
		item := CredentialResponse{Backend: backend, Source: "none"}
		if entry, ok := byBackend[backend]; ok {
			updated := entry.UpdatedAt
			item.Last4, item.UpdatedAt, item.Source = entry.Last4, &updated, "user"
		} else if credentials.EnvKey(backend) != "" {
			item.Source = "instance"
		}
		out = append(out, item)
	}
	return out, nil
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	out, err := h.responses(r.Context(), userID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to list credentials", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, out)
}

func (h *Handler) Put(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	backend, ok := h.backend(w, r)
	if !ok {
		return
	}
	var body PutCredentialRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxKeyBytes)).Decode(&body); err != nil {
		httpjson.WriteError(w, troncerrors.Invalid("invalid json payload"))
		return
	}
	key := strings.TrimSpace(body.APIKey)
	if key == "" {
		httpjson.WriteError(w, troncerrors.Invalid("api_key is required"))
		return
	}
	if err := h.store.Put(r.Context(), userID, backend, key); err != nil {
		if errors.Is(err, credentials.ErrDisabled) {
			httpjson.WriteError(w, troncerrors.Unavailable(err.Error()))
			return
		}
		httpjson.WriteError(w, troncerrors.Internal("failed to save credential", err))
		return
	}
	out, err := h.responses(r.Context(), userID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to read credential", err))
		return
	}
	for _, item := range out {
		if item.Backend == backend {
			httpjson.WriteJSON(w, http.StatusOK, item)
			return
		}
	}
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	backend, ok := h.backend(w, r)
	if !ok {
		return
	}
	if err := h.store.Delete(r.Context(), userID, backend); err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to delete credential", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, EmptyResponse{})
}

func (h *Handler) Test(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}
	backend, ok := h.backend(w, r)
	if !ok {
		return
	}
	key, err := h.store.Resolve(r.Context(), userID, backend)
	if errors.Is(err, credentials.ErrNoKey) {
		httpjson.WriteJSON(w, http.StatusOK, TestCredentialResponse{Error: "no key for " + backend})
		return
	}
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to resolve credential", err))
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, h.probe(r.Context(), backend, key))
}

func (h *Handler) probe(ctx context.Context, backend, key string) TestCredentialResponse {
	vendorURL, ok := h.urls[backend]
	if !ok {
		return TestCredentialResponse{Error: "no key test is available for " + backend}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, vendorURL, nil)
	if err != nil {
		return TestCredentialResponse{Error: "invalid vendor url"}
	}
	switch backend {
	case "anthropic":
		req.Header.Set("x-api-key", key)
		req.Header.Set("anthropic-version", "2023-06-01")
	case "google":
		req.Header.Set("x-goog-api-key", key)
	default:
		req.Header.Set("Authorization", "Bearer "+key)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return TestCredentialResponse{Error: "vendor unreachable"}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return TestCredentialResponse{OK: true}
	}
	return TestCredentialResponse{Error: fmt.Sprintf("vendor rejected the key (HTTP %d)", resp.StatusCode)}
}
