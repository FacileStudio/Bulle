package artifacts

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

// EmptyResponse documents the bodiless 204 of a delete.
type EmptyResponse struct{}

var locateQuery = []documentation.Field{
	{Name: "mycelium_space", Type: "string", Description: "The Mycelium tree that holds the artifact, as the list answers it; looked up among every readable artifact when omitted."},
	{Name: "space_id", Type: "string", Description: "The Bulle space whose tree wins when the id is in more than one tree."},
}

var (
	invalid         = documentation.Error{Status: 400, Code: "invalid_argument", Description: "The content is empty, the title is over 200 characters, or encre refuses the document; the message carries encre's reason."}
	tooLarge        = documentation.Error{Status: 413, Code: "resource_exhausted", Description: "The content is over 1 MB."}
	unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
	csrf            = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}
	notFound        = documentation.Error{Status: 404, Code: "not_found", Description: "No such artifact the caller can read, or space_id names no space of the caller."}
	notConnected    = documentation.Error{Status: 412, Code: "failed_precondition", Description: "The user has no Mycelium connection."}
	unavailable     = documentation.Error{Status: 503, Code: "unavailable", Description: "Mycelium is not reachable."}
)

var Documentation = documentation.Module{
	Name:        "artifacts",
	Description: "The user's Mycelium markdown artifacts across every tree they can read, the agent's included.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/artifacts",
			Summary:      "List artifacts",
			Description:  "Lists every markdown artifact the user can read in Mycelium, newest first: their personal tree (the common tree for an admin, mycelium_space common) and each of their spaces, with the machine that wrote it. available is false, with an empty list, when the user has no Mycelium connection.",
			Auth:         sessionAuth,
			ResponseBody: ListResponse{},
			Errors:       []documentation.Error{unauthenticated, notFound, unavailable},
		},
		{
			Method:       "GET",
			Path:         "/artifacts/{id}",
			Summary:      "Read an artifact",
			Description:  "Reads one artifact with its markdown, frontmatter removed.",
			Auth:         sessionAuth,
			QueryParams:  locateQuery,
			ResponseBody: DetailResponse{},
			Errors:       []documentation.Error{unauthenticated, notFound, notConnected, unavailable},
		},
		{
			Method:       "PUT",
			Path:         "/artifacts/{id}",
			Summary:      "Update an artifact",
			Description:  "Replaces one artifact's markdown, and its title when one is given, after encre validated the document. The answer is the saved artifact, frontmatter removed.",
			Auth:         sessionAuth,
			QueryParams:  locateQuery,
			RequestBody:  UpdateRequest{},
			ResponseBody: DetailResponse{},
			Errors:       []documentation.Error{invalid, unauthenticated, csrf, notFound, notConnected, tooLarge, unavailable},
		},
		{
			Method:       "DELETE",
			Path:         "/artifacts/{id}",
			Summary:      "Delete an artifact",
			Description:  "Deletes one artifact from Mycelium.",
			Auth:         sessionAuth,
			QueryParams:  locateQuery,
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound, notConnected, unavailable},
		},
	},
}
