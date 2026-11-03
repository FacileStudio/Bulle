package spaces

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const spaceAuth = "session cookie or bearer token"

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}

var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}

var invalid = documentation.Error{Status: 400, Code: "invalid_argument", Description: "A missing or too long name, a too long icon, or a negative id."}

var notFound = documentation.Error{Status: 404, Code: "not_found", Description: "The space does not exist or belongs to another user."}

var lastSpace = documentation.Error{Status: 409, Code: "already_exists", Description: "The space is the caller's last one."}

var spaceID = []documentation.Field{{Name: "id", Type: "string", Description: "Space id."}}

var Documentation = documentation.Module{
	Name:        "spaces",
	Description: "Work contexts. Each space has its own chats, jobs and default suite accounts. Every user has at least one.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/spaces",
			Summary:      "List spaces",
			Description:  "Lists the caller's spaces, oldest first. The first one is the default, and it is created on first use.",
			Auth:         spaceAuth,
			ResponseBody: []SpaceResponse{},
			Errors:       []documentation.Error{unauthenticated},
		},
		{
			Method:       "POST",
			Path:         "/spaces",
			Summary:      "Create a space",
			Description:  "Creates a space. name is required; the other fields default to empty, which leaves each app's own default.",
			Auth:         spaceAuth,
			RequestBody:  SpaceRequest{},
			ResponseBody: SpaceResponse{},
			Status:       201,
			Errors:       []documentation.Error{unauthenticated, csrf, invalid},
		},
		{
			Method:       "PUT",
			Path:         "/spaces/{id}",
			Summary:      "Update a space",
			Description:  "Changes the fields present in the body and keeps the others.",
			Auth:         spaceAuth,
			PathParams:   spaceID,
			RequestBody:  SpaceRequest{},
			ResponseBody: SpaceResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, invalid, notFound},
		},
		{
			Method:       "DELETE",
			Path:         "/spaces/{id}",
			Summary:      "Delete a space",
			Description:  "Deletes the space with every session and job in it. The caller's last space cannot be deleted.",
			Auth:         spaceAuth,
			PathParams:   spaceID,
			ResponseBody: EmptyResponse{},
			Status:       204,
			Errors:       []documentation.Error{unauthenticated, csrf, notFound, lastSpace},
		},
	},
}
