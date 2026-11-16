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
				"The acting user (the token's user) must have ai.access in their plan, checked before any rate limit is spent; the call is logged with the caller, the user, the question count, the model and the usage. Rate limited per app and acting user. " +
				"judges (optional) is [\"jev\"], the default, or [\"jev\",\"clef\"]. With both, the server asks Jev and Cloudflare's clef (OpenRouter's Decisions API) at the same time and answers, per question, " +
				"judges (each judge's choice, confidence, probabilities, or its error) and agree. choice is set only when both judges pick the same option and is empty otherwise; confidence is the lower of the two, " +
				"or 0 when a judge gave no answer; probabilities is empty. The top-level judges map gives each judge's model, usage and error: \"not configured\" when the server holds no key for it, \"did not answer\" when its call failed. " +
				"A judge that is down never falls back to the other one: agree is false, and the call still answers 200. Only when no judge at all is configured does it answer 503. " +
				"usage is the sum of both calls and model is Jev's. A two-judge call counts as two calls against the rate limit and is recorded as two rows in the usage ledger, one per judge, so it spends the user's monthly budget twice. " +
				"Without judges, or with [\"jev\"], the response is exactly the single-judge one: no agree, no judges. " +
				"Thresholds are per judge: calibrate a floor on the pair, not on one judge. clef reads only the start of a long text state on Workers AI, so put what matters first and keep a two-judge state short.",
			Auth:         "Registre token-exchange JWT for audience bulle with scope bulle.write and an act claim naming the calling app; session cookies and tokens without act are refused",
			RequestBody:  DecideRequest{},
			ResponseBody: DecideResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "No state, no questions, more than 8, a question that is not a choice with 2 to 64 options, or judges other than [\"jev\"] and [\"jev\",\"clef\"]."},
				{Status: 401, Code: "unauthenticated", Description: "No token or an invalid one."},
				{Status: 403, Code: "permission_denied", Description: "The token lacks bulle.write, has no act claim, or is a session cookie."},
				{Status: 413, Code: "resource_exhausted", Description: "The state is larger than 32 KiB or the body larger than 128 KiB."},
				{Status: 402, Code: "payment_required", Description: "The acting user's plan lacks ai.access; the error carries upgrade_url."},
				{Status: 429, Code: "rate_limited", Description: "Too many calls from this app for this user (a two-judge call counts as two); retry after the Retry-After delay, or the user's monthly AI budget on the server's keys is used up (the message says when it resets)."},
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
				"Rate limited per app and acting user.",
			Auth:         "Registre token-exchange JWT for audience bulle with scope bulle.write and an act claim naming the calling app; session cookies and tokens without act are refused",
			RequestBody:  SummarizeRequest{},
			ResponseBody: SummarizeResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "No query or one over 500 characters, no sources or more than 10, a source without an http(s) url, or a title or snippet over its limit."},
				{Status: 401, Code: "unauthenticated", Description: "No token or an invalid one."},
				{Status: 402, Code: "payment_required", Description: "The acting user's plan lacks ai.access; the error carries upgrade_url."},
				{Status: 403, Code: "permission_denied", Description: "The token lacks bulle.write, has no act claim, or is a session cookie."},
				{Status: 413, Code: "resource_exhausted", Description: "The body is larger than 32 KiB."},
				{Status: 429, Code: "rate_limited", Description: "Too many summaries for this user from this app; retry after the Retry-After delay, or the user's monthly AI budget on the server's keys is used up (the message says when it resets)."},
				{Status: 503, Code: "unavailable", Description: "No model is configured, or it did not answer valid JSON. Show no summary."},
			},
		},
		{
			Method:  "POST",
			Path:    "/ai/complete",
			Summary: "Complete a prompt for an app's user",
			Description: "Answers one prompt with the user's AI on the default model. input is untrusted data and goes to the model in a delimited block it is told never to obey. " +
				"With schema (a JSON Schema without remote $ref) the answer is one validated JSON value in json; otherwise it is text. " +
				"space_id (optional, at most 128 bytes) names the Registre shared space the request is for: AI is granted when the user's own plan or, " +
				"for a member of that space, the space's plan has ai.access. " +
				"Limits: system 4 KiB, prompt 8 KiB, input 64 KiB, schema 8 KiB, max_tokens 1 to 4096 (default 1024). " +
				"Rate limited per app and user, and capped at 200 completions per user per UTC day. Logged with the app, the user, the model and the sizes, never the text.",
			Auth:         "Registre token-exchange JWT for audience bulle with scope bulle.write and an act claim naming the calling app; session cookies are refused",
			RequestBody:  CompleteRequest{},
			ResponseBody: CompleteResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "No prompt or one over 8 KiB, max_tokens out of range, text that is not UTF-8, or a schema that is not a valid local JSON Schema."},
				{Status: 401, Code: "unauthenticated", Description: "No token or an invalid one."},
				{Status: 402, Code: "payment_required", Description: "Neither the user's plan nor the plan of the space_id they are a member of has ai.access; the error carries upgrade_url."},
				{Status: 403, Code: "permission_denied", Description: "The token lacks bulle.write, has no act claim, or is a session cookie."},
				{Status: 413, Code: "resource_exhausted", Description: "system, input or schema over its limit, or a body over 96 KiB."},
				{Status: 429, Code: "rate_limited", Description: "Too many completions for this user from this app, the daily cap is reached, or the user's monthly AI budget on the server's keys is used up (the message says when it resets)."},
				{Status: 503, Code: "unavailable", Description: "No model is configured, it did not answer, or its answer did not match the schema."},
			},
		},
	},
}
