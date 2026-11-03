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
			Description:  "Returns session summaries ordered by latest update time, optionally only those of one space.",
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
			Path:         "/sessions/{id}/messages",
			Summary:      "Send a message to start an assistant run",
			Description:  "Enqueues a user message and starts a live run in the session.",
			Auth:         sessionAuth,
			RequestBody:  SendMessageRequest{},
			ResponseBody: SendMessageResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Message cannot be empty."},
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
			Description:  "Submits a human decision on a gated tool invocation.",
			Auth:         sessionAuth,
			RequestBody:  DecideRequest{},
			ResponseBody: EmptyResponse{},
			Errors:       []documentation.Error{unauthenticated, csrf, notFound},
		},
	},
}
