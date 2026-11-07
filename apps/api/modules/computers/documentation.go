package computers

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}

var unavailable = documentation.Error{Status: 503, Code: "unavailable", Description: "The computer service is unreachable or busy."}

var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}

var Documentation = documentation.Module{
	Name:        "computers",
	Description: "The signed-in user's own computer: one microVM per account. Every route acts on the caller's computer only.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/computer",
			Summary:      "Read the computer's status and resources",
			Description:  "status is absent, creating, running, stopped or error. An absent computer is created on first start or first chat. network is open or offline.",
			Auth:         sessionAuth,
			ResponseBody: ComputerResponse{},
			Errors:       []documentation.Error{unauthenticated, unavailable},
		},
		{
			Method:       "POST",
			Path:         "/computer/start",
			Summary:      "Create or start the computer",
			Description:  "Returns at once; poll GET /computer until status is running. Cookie callers must send X-Facile-CSRF.",
			Auth:         sessionAuth,
			RequestBody:  EmptyRequest{},
			ResponseBody: ComputerResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, unavailable},
		},
		{
			Method:       "POST",
			Path:         "/computer/stop",
			Summary:      "Stop the computer",
			Description:  "Idempotent. Cookie callers must send X-Facile-CSRF.",
			Auth:         sessionAuth,
			RequestBody:  EmptyRequest{},
			ResponseBody: ComputerResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, unavailable},
		},
		{
			Method:       "POST",
			Path:         "/computer/snapshots",
			Summary:      "Snapshot the computer",
			Description:  "Stops the computer first, since snapshots are taken offline, then tags the disk. The computer stays stopped.",
			Auth:         sessionAuth,
			RequestBody:  TagRequest{},
			ResponseBody: ComputerResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or a malformed tag."},
				unauthenticated, csrf,
				{Status: 404, Code: "not_found", Description: "The user has no computer."},
				unavailable,
			},
		},
		{
			Method:       "POST",
			Path:         "/computer/rollback",
			Summary:      "Roll the computer back to a snapshot",
			Description:  "Stops the computer first, then restores the tagged snapshot. The computer stays stopped.",
			Auth:         sessionAuth,
			RequestBody:  TagRequest{},
			ResponseBody: ComputerResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or a malformed tag."},
				unauthenticated, csrf,
				{Status: 404, Code: "not_found", Description: "The user has no computer, or the snapshot does not exist."},
				unavailable,
			},
		},
		{
			Method:  "PUT",
			Path:    "/computer/network",
			Summary: "Choose whether the computer may reach the network",
			Description: "network is open (the default) or offline. Offline denies every outbound connection from the " +
				"computer on the host side, so installs and web access from its shell fail; web_fetch still works " +
				"because it runs outside the computer. The worker reads the mode at boot, so a running computer " +
				"restarts at once and a stopped one takes it on its next start. Cookie callers must send X-Facile-CSRF.",
			Auth:         sessionAuth,
			RequestBody:  NetworkRequest{},
			ResponseBody: ComputerResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or a network other than open or offline."},
				unauthenticated, csrf, unavailable,
			},
		},
	},
}
