package auth

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
	"github.com/FacileStudio/porte"
)

var Documentation = documentation.Module{
	Name:        "auth",
	Description: "Authentication: porte sessions over Registre single sign-on. No password routes exist.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/auth/config",
			Summary:      "Sign-in configuration for the client",
			Description:  "Reports sso_only and oidc_enabled.",
			Auth:         "public",
			ResponseBody: porte.ConfigResponse{},
			Errors:       []documentation.Error{},
		},
		{
			Method:       "POST",
			Path:         "/auth/logout",
			Summary:      "End the current session",
			Description:  "Idempotent. Cookie callers must send X-Facile-CSRF.",
			Auth:         "session cookie or bearer token",
			ResponseBody: porte.LogoutResponse{},
			Errors: []documentation.Error{
				{Status: 403, Code: "forbidden", Description: "A cookie-authenticated request without the X-Facile-CSRF header."},
			},
		},
		{
			Method:      "GET",
			Path:        "/auth/oidc",
			Summary:     "Start the OIDC flow",
			Description: "Redirects to Registre with PKCE.",
			Auth:        "public",
			Errors:      []documentation.Error{},
		},
		{
			Method:      "GET",
			Path:        "/auth/oidc/callback",
			Summary:     "OIDC callback",
			Description: "Exchanges the code and opens a session. A failure, including an unverified email, redirects to the login page with ?error=<reason>.",
			Auth:        "public",
			Errors:      []documentation.Error{},
		},
		{
			Method:       "POST",
			Path:         "/auth/oidc/device/exchange",
			Summary:      "Trade a CLI device token for a Bulle session",
			Description:  "Verifies an access token issued to OIDC_CLI_AUDIENCE and mints a session for the matching identity. Mounted only when OIDC_CLI_AUDIENCE is set. No account is created here.",
			Auth:         "public",
			RequestBody:  porte.DeviceExchangeRequest{},
			ResponseBody: porte.ExchangeResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or a missing access_token."},
				{Status: 401, Code: "unauthenticated", Description: "The token failed verification, or its subject holds no account here."},
			},
		},
		{
			Method:       "GET",
			Path:         "/auth/me",
			Summary:      "Return the signed-in user",
			Auth:         "session cookie or bearer token",
			ResponseBody: MeResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "No valid session."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
	},
}
