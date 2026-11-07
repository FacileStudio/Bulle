package memory

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

// EmptyResponse documents the bodiless 204 of a delete.
type EmptyResponse struct{}

var spaceQuery = []documentation.Field{{Name: "space_id", Type: "string", Description: "The Bulle space; the default space when omitted."}}

var (
	unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
	badPath         = documentation.Error{Status: 400, Code: "invalid_argument", Description: "The path is absolute, has a hidden or .. segment, or does not end in .md; or the content is empty or over 64 KB."}
	csrf            = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header, or a Mycelium space the caller may not change."}
	notFound        = documentation.Error{Status: 404, Code: "not_found", Description: "No such page in the space, or space_id names no space of the caller."}
	notConnected    = documentation.Error{Status: 412, Code: "failed_precondition", Description: "The user has no Mycelium connection."}
	unavailable     = documentation.Error{Status: 503, Code: "unavailable", Description: "Mycelium is not reachable."}
)

var Documentation = documentation.Module{
	Name:        "memory",
	Description: "The pages of the agent's long-term memory, stored in the space's Mycelium space, for the user to see, edit and delete.",
	Routes: []documentation.Route{
		{
			Method: "GET", Path: "/memory/pages", Summary: "List memory pages",
			Description: "Lists the markdown pages of the space's memory, profile.md first. available is false, with an empty list, when the user has no Mycelium connection.",
			Auth:        sessionAuth, QueryParams: spaceQuery, ResponseBody: ListResponse{},
			Errors: []documentation.Error{unauthenticated, notFound, unavailable},
		},
		{
			Method: "GET", Path: "/memory/pages/{path}", Summary: "Read a memory page",
			Description: "Reads one page's markdown. path is relative to the memory directory, such as people/ann.md, each segment URL-escaped.",
			Auth:        sessionAuth, QueryParams: spaceQuery, ResponseBody: PageResponse{},
			Errors: []documentation.Error{unauthenticated, badPath, notFound, notConnected, unavailable},
		},
		{
			Method: "PUT", Path: "/memory/pages/{path}", Summary: "Write a memory page",
			Description: "Creates or replaces one page with the given markdown, as memory_write does.",
			Auth:        sessionAuth, QueryParams: spaceQuery, RequestBody: WriteRequest{}, ResponseBody: PageResponse{},
			Errors: []documentation.Error{unauthenticated, badPath, csrf, notFound, notConnected, unavailable},
		},
		{
			Method: "DELETE", Path: "/memory/pages/{path}", Summary: "Delete a memory page",
			Description: "Deletes one page, as memory_delete does. Deleting a page that is already gone succeeds.",
			Auth:        sessionAuth, QueryParams: spaceQuery, ResponseBody: EmptyResponse{},
			Errors: []documentation.Error{unauthenticated, badPath, csrf, notFound, notConnected, unavailable},
		},
	},
}
