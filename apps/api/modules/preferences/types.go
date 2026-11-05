package preferences

type EmptyResponse struct{}

// PreferencesResponse is what the settings page reads: the user's own choices,
// the models the instance offers, and the instance default an empty model falls
// back to.
type PreferencesResponse struct {
	Model        string   `json:"model"`
	CLIConfig    string   `json:"cli_config"`
	Models       []string `json:"models"`
	DefaultModel string   `json:"default_model"`
}

// UpdatePreferencesRequest is a partial write: a field left out is left alone,
// so the model picker and the CLI editor save without overwriting each other.
type UpdatePreferencesRequest struct {
	Model     *string `json:"model"`
	CLIConfig *string `json:"cli_config"`
}

// CLIConfigResponse is what the CLI reads at startup.
type CLIConfigResponse struct {
	Config string `json:"config"`
}
