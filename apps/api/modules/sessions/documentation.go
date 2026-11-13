package sessions

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}

var notFound = documentation.Error{Status: 404, Code: "not_found", Description: "The requested session or run was not found."}

var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}

var spaceNotFound = documentation.Error{Status: 404, Code: "not_found", Description: "space_id names no space of the caller."}

var conflict = documentation.Error{Status: 409, Code: "conflict", Description: "The session already has an active run."}

var tooMany = documentation.Error{Status: 429, Code: "too_many_requests", Description: "Too many concurrent active runs for this user."}

var Documentation = documentation.Module{
	Name:        "sessions",
	Description: "Agent conversations, runs, and event streaming. Every session and run belongs to the caller.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/sessions",
			Summary:      "List sessions for the current user",
			Description:  "Returns session summaries ordered by latest update time, optionally only those of one space. Each carries status: running while a run is in flight, waiting while a tool call waits for approval, idle otherwise. unread is true when the session changed since the caller last marked it read; a job's session starts unread, a Matrix thread never is.",
			Auth:         sessionAuth,
			QueryParams:  []documentation.Field{{Name: "space_id", Type: "string", Description: "Only the sessions of this space."}},
			ResponseBody: []SessionSummary{},
			Errors:       []documentation.Error{unauthenticated},
		},
		{
			Method:       "POST",
			Path:         "/sessions",
			Summary:      "Create a new chat session",
			Description:  "Initialises a new conversation in space_id, or in the user's default space when it is omitted. Cookie callers must send X-Facile-CSRF.",
			Auth:         sessionAuth,
			RequestBody:  CreateSessionRequest{},
			ResponseBody: CreateSessionResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, spaceNotFound},
		},
		{
			Method:      "GET",
			Path:        "/sessions/search",
			Summary:     "Search past chats",
			Description: "Matches q, ignoring case, against the title and the user and assistant text of the caller's web and matrix chats in space_id, or in the default space when it is omitted. Tool calls and tool output are not searched, and ideas, memory, cron and guest sessions never match. Answers at most 20 results, newest first, each with a snippet around the first matching message text (empty when only the title matched).",
			Auth:        sessionAuth,
			QueryParams: []documentation.Field{
				{Name: "q", Type: "string", Description: "What to look for, at least 2 characters."},
				{Name: "space_id", Type: "string", Description: "The space to search."},
			},
			ResponseBody: SearchResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "q is shorter than 2 characters."},
				unauthenticated, spaceNotFound,
			},
		},
		{
			Method:       "GET",
			Path:         "/sessions/{id}",
			Summary:      "Get session details and message transcript",
			Description:  "Returns session metadata, completed messages, and active_run_id if a run is currently in progress.",
			Auth:         sessionAuth,
			ResponseBody: SessionDetailResponse{},
			Errors:       []documentation.Error{unauthenticated, notFound},
		},
		{
			Method:       "DELETE",
			Path:         "/sessions/{id}",
			Summary:      "Delete a chat session",
			Description:  "Deletes the session and all associated runs and messages.",
			Auth:         sessionAuth,
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound},
		},
		{
			Method:       "POST",
			Path:         "/sessions/{id}/read",
			Summary:      "Mark a chat session read",
			Description:  "Records that the caller has read the session up to now, so unread turns false until a new message arrives. Every open app gets session.read on the live stream. Cookie callers must send X-Facile-CSRF.",
			Auth:         sessionAuth,
			RequestBody:  EmptyRequest{},
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound},
		},
		{
			Method:       "POST",
			Path:         "/sessions/{id}/messages",
			Summary:      "Send a message to start an assistant run",
			Description:  "Enqueues a user message and starts a live run in the session. attachments lists up to 10 ids from POST /attachments that belong to the user; text may be empty when attachments are given. page {app, id, title, url} names the docs, toile or feuilles page a chat embedded in that app has open; the model reads it in front of the message and the transcript never shows it.",
			Auth:         sessionAuth,
			RequestBody:  SendMessageRequest{},
			ResponseBody: SendMessageResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Message cannot be empty."},
				{Status: 400, Code: "invalid_argument", Description: "page names no docs, toile or feuilles page, or has no id or https url."},
				unauthenticated, csrf, notFound, conflict, tooMany,
			},
		},
		{
			Method:       "GET",
			Path:         "/runs/{id}/events",
			Summary:      "Stream run events over Server-Sent Events",
			Description:  "Replays events starting after last_event_id, then streams live events until completion.",
			Auth:         sessionAuth,
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, notFound},
		},
		{
			Method:       "POST",
			Path:         "/runs/{id}/stop",
			Summary:      "Stop an active assistant run",
			Description:  "Signals the live run to cancel execution.",
			Auth:         sessionAuth,
			RequestBody:  EmptyRequest{},
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound},
		},
		{
			Method:       "POST",
			Path:         "/runs/{id}/approvals/{call}",
			Summary:      "Approve or deny an outward tool call",
			Description:  "Submits a human decision on a gated tool invocation, or on a confirmation an MCP server asked for mid-call (Courrier before it sends mail), which is answered to that server. args and as are refused with 400: no call can be edited on its card. Pass remember: true with an approval to let the same tool run unasked for the rest of the run; it is ignored for server confirmations, deletions, file saves and browser or web fetch tools, which always ask.",
			Auth:         sessionAuth,
			RequestBody:  DecideRequest{},
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound},
		},
	},
}
