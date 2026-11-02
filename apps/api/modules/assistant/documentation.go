package assistant

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "assistant",
	Description: "Assistant profile: persona, avatar, archetype, and style traits.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/assistant",
			Summary:      "Fetch the assistant profile",
			Description:  "Returns the user's assistant profile, defaulting to empty fields if none exists.",
			Auth:         "session cookie or bearer token",
			ResponseBody: ProfileResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
			},
		},
		{
			Method:       "PUT",
			Path:         "/assistant",
			Summary:      "Update the assistant profile",
			Description:  "Updates or creates the assistant profile for the authenticated user.",
			Auth:         "session cookie or bearer token",
			RequestBody:  UpdateProfileRequest{},
			ResponseBody: ProfileResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "bad_request", Description: "Invalid JSON body."},
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
			},
		},
	},
}
