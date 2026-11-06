package artifacts

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

// EmptyResponse documents the bodiless 204 of a delete.
type EmptyResponse struct{}

var spaceQuery = []documentation.Field{{Name: "space_id", Type: "string", Description: "The Bulle space; the default space when omitted."}}

var (
	unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
	csrf            = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}
	notFound        = documentation.Error{Status: 404, Code: "not_found", Description: "No such artifact in the space, or space_id names no space of the caller."}
	notConnected    = documentation.Error{Status: 412, Code: "failed_precondition", Description: "The user has no Mycelium connection."}
	unavailable     = documentation.Error{Status: 503, Code: "unavailable", Description: "Mycelium is not reachable."}
)

var Documentation = documentation.Module{
	Name:        "artifacts",
	Description: "Markdown artifacts the agent writes, stored in the space's Mycelium space.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/artifacts",
			Summary:      "List artifacts",
			Description:  "Lists the markdown artifacts of a space, newest first. available is false, with an empty list, when the user has no Mycelium connection.",
			Auth:         sessionAuth,
			QueryParams:  spaceQuery,
			ResponseBody: ListResponse{},
			Errors:       []documentation.Error{unauthenticated, notFound, unavailable},
		},
		{
			Method:       "GET",
			Path:         "/artifacts/{id}",
			Summary:      "Read an artifact",
			Description:  "Reads one artifact with its markdown, frontmatter removed.",
			Auth:         sessionAuth,
			QueryParams:  spaceQuery,
			ResponseBody: DetailResponse{},
			Errors:       []documentation.Error{unauthenticated, notFound, notConnected, unavailable},
		},
		{
			Method:       "DELETE",
			Path:         "/artifacts/{id}",
			Summary:      "Delete an artifact",
			Description:  "Deletes one artifact from Mycelium.",
			Auth:         sessionAuth,
			QueryParams:  spaceQuery,
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound, notConnected, unavailable},
		},
	},
}
