package preferences

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}
var badRequest = documentation.Error{Status: 400, Code: "invalid_argument", Description: "A model the instance does not offer, or a CLI config the CLI would refuse to load."}

var Documentation = documentation.Module{
	Name: "preferences",
	Description: "The signed-in user's own defaults: the model new sessions start on and the CLI settings their machines read. " +
		"An empty model means the instance default. The saved CLI config is one layer, above the shipped defaults and below ~/.config/bulle/config.yml, " +
		"and may not carry server.url.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/preferences",
			Summary:      "Read preferences",
			Description:  "Returns the user's default model and CLI config, the models the instance offers, and the instance default an empty model falls back to.",
			Auth:         sessionAuth,
			ResponseBody: PreferencesResponse{},
			Errors:       []documentation.Error{unauthenticated},
		},
		{
			Method:       "PUT",
			Path:         "/preferences",
			Summary:      "Save preferences",
			Description:  "Writes the fields the request carries and leaves the others alone. A model must be one the instance offers; an empty one restores the instance default. The CLI config is validated as YAML with the same strictness the CLI loads it with.",
			Auth:         sessionAuth,
			RequestBody:  UpdatePreferencesRequest{},
			ResponseBody: PreferencesResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, badRequest},
		},
		{
			Method:       "GET",
			Path:         "/cli/config",
			Summary:      "Read the CLI config",
			Description:  "Returns the saved CLI settings as the YAML document the CLI merges under its own file. Read by the CLI at startup with a login token.",
			Auth:         sessionAuth,
			ResponseBody: CLIConfigResponse{},
			Errors:       []documentation.Error{unauthenticated},
		},
	},
}
