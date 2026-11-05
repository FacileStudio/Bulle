package preferences

import "github.com/FacileStudio/Bulle/apps/api/internal/catalog"

// PreferencesResponse is what the settings page reads: the user's own default,
// the catalogue the current backend offers, and the instance default an empty
// model falls back to.
type PreferencesResponse struct {
	Model        string          `json:"model"`
	Catalog      []catalog.Model `json:"catalog"`
	DefaultModel string          `json:"default_model"`
}

// UpdatePreferencesRequest writes the default model. A nil field leaves the
// saved one alone; an empty string restores the instance default.
type UpdatePreferencesRequest struct {
	Model *string `json:"model"`
}
