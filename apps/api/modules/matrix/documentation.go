package matrix

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
	"github.com/FacileStudio/Bulle/apps/api/internal/matrixbot"
)

// Documentation describes the matrix routes for the API reference.
var Documentation = documentation.Module{
	Name:        "matrix",
	Description: "The user's chat agent in Prisme. Bulle provisions it automatically: it joins every channel of the user's Prisme spaces and answers when mentioned, in a thread.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/matrix",
			Summary:      "Read the chat agent status",
			Description:  "Returns whether chat agents are available on this server, the agent's Matrix user id, each room it sits in with the Bulle space that room works in, and its live status: off, connecting, running, or the last error.",
			Auth:         "session cookie or bearer token",
			ResponseBody: matrixbot.Status{},
			Errors:       []documentation.Error{{Status: 401, Code: "unauthenticated", Description: "No valid session."}},
		},
	},
}
