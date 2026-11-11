package ai

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

// Documentation describes the decide, summarize and complete routes for the API reference.
var Documentation = documentation.Module{
	Name:        "ai",
	Description: "Decisions, summaries and completions other Facile apps ask Bulle's AI for, so no other app holds an AI key.",
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
		{
			Method:  "POST",
			Path:    "/ai/summarize",
			Summary: "Summarize search results",
			Description: "Summarizes 1 to 10 search results for a query with the acting user's AI, citing results as [n] by their 1-based position. " +
				"Source text is treated as untrusted data; the model must answer strict JSON, and any citation of a position that was not sent is dropped. " +
				"citations lists the positions the summary cites. The call is logged with the caller, the user, the source count and the model, never the query. " +
				"Rate limited per calling service and acting user.",
			Auth:         "service key (Authorization: Bearer, listed in BULLE_AI_CALLERS) plus X-Bulle-Acting-For; session cookies and user tokens are refused",
			RequestBody:  SummarizeRequest{},
			ResponseBody: SummarizeResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "No query or one over 500 characters, no sources or more than 10, a source without an http(s) url, or a title or snippet over its limit."},
				{Status: 401, Code: "unauthenticated", Description: "No service key, an unknown one, or a session cookie on the request."},
				{Status: 402, Code: "payment_required", Description: "The acting user's plan lacks ai.access; the error carries upgrade_url."},
				{Status: 403, Code: "permission_denied", Description: "X-Bulle-Acting-For is missing or names no Bulle user."},
				{Status: 413, Code: "resource_exhausted", Description: "The body is larger than 32 KiB."},
				{Status: 429, Code: "rate_limited", Description: "Too many summaries for this user from this service; retry after the Retry-After delay."},
				{Status: 503, Code: "unavailable", Description: "No model is configured, or it did not answer valid JSON. Show no summary."},
			},
		},
		{
			Method:  "POST",
			Path:    "/ai/complete",
			Summary: "Complete a prompt for an app's user",
			Description: "Answers one prompt with the user's AI on the default model. input is untrusted data and goes to the model in a delimited block it is told never to obey. " +
				"With schema (a JSON Schema without remote $ref) the answer is one validated JSON value in json; otherwise it is text. " +
				"Limits: system 4 KiB, prompt 8 KiB, input 64 KiB, schema 8 KiB, max_tokens 1 to 4096 (default 1024). " +
				"Rate limited per app and user, and capped at 200 completions per user per UTC day. Logged with the app, the user, the model and the sizes, never the text.",
			Auth:         "Registre token-exchange JWT for audience bulle with scope bulle.write and an act claim naming the calling app; session cookies are refused",
			RequestBody:  CompleteRequest{},
			ResponseBody: CompleteResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "No prompt or one over 8 KiB, max_tokens out of range, text that is not UTF-8, or a schema that is not a valid local JSON Schema."},
				{Status: 401, Code: "unauthenticated", Description: "No token or an invalid one."},
				{Status: 402, Code: "payment_required", Description: "The user's plan lacks ai.access; the error carries upgrade_url."},
				{Status: 403, Code: "permission_denied", Description: "The token lacks bulle.write, has no act claim, or is a session cookie."},
				{Status: 413, Code: "resource_exhausted", Description: "system, input or schema over its limit, or a body over 96 KiB."},
				{Status: 429, Code: "rate_limited", Description: "Too many completions for this user from this app, or the daily cap is reached."},
				{Status: 503, Code: "unavailable", Description: "No model is configured, it did not answer, or its answer did not match the schema."},
			},
		},
	},
}
