package preferences

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}
var badRequest = documentation.Error{Status: 400, Code: "invalid_argument", Description: "A model the current backend does not offer."}

var Documentation = documentation.Module{
	Name: "preferences",
	Description: "The signed-in user's own default model. An empty model means the instance default, and a saved one is returned " +
		"by GET /llm so every new session, in the web app and in the CLI, starts on it. The choices come from the same catalogue " +
		"the CLI's /model lists (GET /models).",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/preferences",
			Summary:      "Read the default model",
			Description:  "Returns the user's default model, the models the current backend offers, and the instance default an empty model falls back to.",
			Auth:         sessionAuth,
			ResponseBody: PreferencesResponse{},
			Errors:       []documentation.Error{unauthenticated},
		},
		{
			Method:       "PUT",
			Path:         "/preferences",
			Summary:      "Save the default model",
			Description:  "Writes the model named in the request. It must be the instance default or one the backend advertises; an empty string restores the instance default.",
			Auth:         sessionAuth,
			RequestBody:  UpdatePreferencesRequest{},
			ResponseBody: PreferencesResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, badRequest},
		},
	},
}
