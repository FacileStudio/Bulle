package ai

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

// Documentation describes the decide route for the API reference.
var Documentation = documentation.Module{
	Name:        "ai",
	Description: "Decisions other Facile services ask Bulle's decision model for, so no other app holds an AI key.",
	Routes: []documentation.Route{
		{
			Method:  "POST",
			Path:    "/ai/decide",
			Summary: "Answer choice questions",
			Description: "Sends one state and up to 8 choice questions (2 to 64 options each, state at most 32 KiB) to Jev and returns every answer. " +
				"X-Bulle-Acting-For must name an existing Bulle user; the call is logged with the caller, the user, the question count, the model and the usage. Rate limited per calling service.",
			Auth:         "service key (Authorization: Bearer, listed in BULLE_AI_CALLERS) plus X-Bulle-Acting-For; session cookies and user tokens are refused",
			RequestBody:  DecideRequest{},
			ResponseBody: DecideResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "No state, no questions, more than 8, or a question that is not a choice with 2 to 64 options."},
				{Status: 401, Code: "unauthenticated", Description: "No service key, an unknown one, or a session cookie on the request."},
				{Status: 403, Code: "permission_denied", Description: "X-Bulle-Acting-For is missing or names no Bulle user."},
				{Status: 413, Code: "resource_exhausted", Description: "The state is larger than 32 KiB or the body larger than 128 KiB."},
				{Status: 429, Code: "rate_limited", Description: "Too many calls from this service; retry after the Retry-After delay."},
				{Status: 503, Code: "unavailable", Description: "No decision model is configured, or it did not answer. Treat it as no answer."},
			},
		},
	},
}
