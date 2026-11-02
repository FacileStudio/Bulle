package credentials

import "time"

type CredentialResponse struct {
	Backend   string     `json:"backend"`
	Last4     string     `json:"last4"`
	UpdatedAt *time.Time `json:"updated_at"`
	Source    string     `json:"source"`
}

type PutCredentialRequest struct {
	APIKey string `json:"api_key"`
}

type TestCredentialResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}
