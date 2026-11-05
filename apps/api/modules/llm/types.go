package llm

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/catalog"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ProxyRequest struct{}

type ProxyResponse struct{}

// Config is the model setup the instance offers: the backend every call goes
// through, the default model, the models a client may switch to, and the
// signed-in user's assistant persona as system prompt text.
type Config struct {
	Backend string   `json:"backend"`
	Model   string   `json:"model"`
	Models  []string `json:"models"`
	Persona string   `json:"persona,omitempty"`
}

// Offer is the instance's model setup plus the lookups that fill each
// request: the user's persona, and the model they chose in the web settings.
// Either may be nil.
type Offer struct {
	Config
	Persona  func(ctx context.Context, userID int64) string
	ModelFor func(ctx context.Context, userID int64) string
	Catalog  func(ctx context.Context, userID int64) []catalog.Model
}
