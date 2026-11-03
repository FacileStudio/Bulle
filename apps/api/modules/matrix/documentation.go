package matrix

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
	"github.com/FacileStudio/Bulle/apps/api/internal/matrixbot"
)

const sessionAuth = "session cookie or bearer token"

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}

// Documentation describes the matrix routes for the API reference.
var Documentation = documentation.Module{
	Name:        "matrix",
	Description: "The user's Matrix bot: a teammate in each space's linked room that answers when mentioned, in a thread. The password is write-only.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/matrix",
			Summary:      "Read the Matrix bot settings",
			Description:  "Returns the bot account (whether a password is set, never the password), the Matrix users allowed to call it, the room linked to each space, and the bot's live status: off, connecting, running, or the last error.",
			Auth:         sessionAuth,
			ResponseBody: matrixbot.Settings{},
			Errors:       []documentation.Error{unauthenticated},
		},
		{
			Method:       "PUT",
			Path:         "/matrix",
			Summary:      "Save the Matrix bot settings",
			Description:  "Saves the bot account, the allowlist and the space rooms, then restarts the bot. An empty password keeps the stored one; an empty room unlinks its space.",
			Auth:         sessionAuth,
			RequestBody:  matrixbot.Settings{},
			ResponseBody: matrixbot.Settings{},
			Errors: []documentation.Error{
				unauthenticated,
				{Status: 400, Code: "invalid_argument", Description: "A malformed homeserver, user id or room id, a space the user does not own, or an enabled bot with no password."},
				{Status: 503, Code: "unavailable", Description: "The server has no BULLE_CREDENTIALS_KEY, so the password cannot be stored."},
			},
		},
	},
}
