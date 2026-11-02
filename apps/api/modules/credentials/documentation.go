package credentials

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

type EmptyRequest struct{}

type EmptyResponse struct{}

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}
var badRequest = documentation.Error{Status: 400, Code: "invalid_argument", Description: "Unknown backend, malformed payload, or an empty api_key."}
var unavailable = documentation.Error{Status: 503, Code: "unavailable", Description: "The server has no BULLE_CREDENTIALS_KEY, so keys cannot be stored."}

var Documentation = documentation.Module{
	Name:        "credentials",
	Description: "Per-user provider API keys, encrypted at rest. Keys are write-only: no endpoint returns one.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/credentials",
			Summary:      "List provider keys",
			Description:  "Lists the four backends with the last four characters of the user's key and whether the key comes from the user, the instance environment, or is missing.",
			Auth:         sessionAuth,
			ResponseBody: []CredentialResponse{},
			Errors:       []documentation.Error{unauthenticated},
		},
		{
			Method:       "PUT",
			Path:         "/credentials/{backend}",
			Summary:      "Save a provider key",
			Description:  "Encrypts and stores the user's key for a backend, replacing any previous one.",
			Auth:         sessionAuth,
			RequestBody:  PutCredentialRequest{},
			ResponseBody: CredentialResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, badRequest, unavailable},
		},
		{
			Method:       "DELETE",
			Path:         "/credentials/{backend}",
			Summary:      "Remove a provider key",
			Description:  "Deletes the user's key for a backend. The instance key, if any, applies again.",
			Auth:         sessionAuth,
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, badRequest},
		},
		{
			Method:       "POST",
			Path:         "/credentials/{backend}/test",
			Summary:      "Test a provider key",
			Description:  "Makes a minimal call to the vendor with the key that would be used and reports whether it was accepted.",
			Auth:         sessionAuth,
			RequestBody:  EmptyRequest{},
			ResponseBody: TestCredentialResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, badRequest},
		},
	},
}
