package ideas

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token; another suite app uses an RFC 8693 token exchanged for audience bulle, with act naming the app and scope bulle.read (GET) or bulle.write (POST, PUT, PATCH, DELETE)"

// EmptyRequest documents a route that takes no body.
type EmptyRequest struct{}

// EmptyResponse documents a route that answers no body.
type EmptyResponse struct{}

var unauthenticated = documentation.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
var csrf = documentation.Error{Status: 403, Code: "permission_denied", Description: "A cookie-authenticated request without the X-Facile-CSRF header."}
var badRequest = documentation.Error{Status: 400, Code: "invalid_argument", Description: "Malformed payload or a value outside its limits."}
var spaceNotFound = documentation.Error{Status: 404, Code: "not_found", Description: "space_id names no space of the caller."}
var ideaNotFound = documentation.Error{Status: 404, Code: "not_found", Description: "No idea of the caller has this id."}
var goalNotFound = documentation.Error{Status: 404, Code: "not_found", Description: "No goal of the caller has this id."}
var closedIdea = documentation.Error{Status: 409, Code: "already_exists", Description: "The idea is done or expired, so its status no longer changes."}
var busy = documentation.Error{Status: 409, Code: "already_exists", Description: "A batch is already running for this space."}
var tooMany = documentation.Error{Status: 429, Code: "rate_limited", Description: "More than 3 manual batches today in the space's timezone (per space for Bulle's own, per user and app across spaces with ?app=), or too many active runs."}
var forbidden = documentation.Error{Status: 403, Code: "permission_denied", Description: "A bearer token without the bulle scope the method needs, or one that acts for nobody (no act claim)."}
var paymentRequired = documentation.Error{Status: 402, Code: "payment_required", Description: "The user's plan lacks ai.access; the error carries upgrade_url. Reads stay allowed."}
var badApp = documentation.Error{Status: 400, Code: "invalid_argument", Description: "app is not one of courrier, agenda, opus, sablier, mycelium, nuage, glouton or bulle."}
var runNotFound = documentation.Error{Status: 404, Code: "not_found", Description: "No run of the caller has this id."}

var spaceField = documentation.Field{Name: "space_id", Type: "string", Description: "The active space: a Bulle space id or the Registre shared space id it mirrors, checked against the user's spaces; the default space when omitted."}
var appField = documentation.Field{Name: "app", Type: "string", Description: "Only this app's ideas: courrier, agenda, opus, sablier, mycelium, nuage, glouton or bulle. Every app's when omitted."}

var Documentation = documentation.Module{
	Name:        "ideas",
	Description: "Bulle's proactive suggestions: a daily batch of idea cards per space, the goals that steer it, and its settings.",
	Routes: []documentation.Route{
		{
			Method: "GET", Path: "/ideas", Summary: "Show the Ideas page",
			Description:  "Lists the space's new and saved ideas (newest batch first), its goals (open first), its settings, whether a batch is running and when the last one finished. Creates the settings row with defaults the first time, so the space starts getting a daily batch.",
			Auth:         sessionAuth,
			QueryParams:  []documentation.Field{spaceField, appField},
			ResponseBody: ListResponse{},
			Errors:       []documentation.Error{unauthenticated, forbidden, badApp, spaceNotFound},
		},
		{
			Method: "PATCH", Path: "/ideas/{id}", Summary: "Save, dismiss or rate an idea",
			Description: "Changes an idea's status (saved, new or dismissed), its thumbs feedback, or the reason it was dismissed.",
			Auth:        sessionAuth, RequestBody: UpdateIdeaRequest{}, ResponseBody: IdeaResponse{},
			Errors: []documentation.Error{unauthenticated, csrf, badRequest, ideaNotFound, closedIdea},
		},
		{
			Method: "POST", Path: "/ideas/{id}/start", Summary: "Do an idea",
			Description: "Opens a web chat in the idea's space titled with the idea, sends its prompt (and draft) as the user's first message, and marks the idea done. Starting a done idea again answers its chat.",
			Auth:        sessionAuth, RequestBody: EmptyRequest{}, ResponseBody: StartResponse{},
			Errors: []documentation.Error{unauthenticated, csrf, ideaNotFound, tooMany},
		},
		{
			Method: "POST", Path: "/ideas/{id}/run", Summary: "Run an idea in the background",
			Description: "Starts the idea as a background agent run, for an app that shows the result in place: the run records a web chat in the idea's space (continue it in Bulle at /s/{session_id}), its writes wait for approval as in a chat, and GET /runs/{id} follows it. Running a done idea again answers its latest run.",
			Auth:        sessionAuth, QueryParams: []documentation.Field{spaceField}, RequestBody: EmptyRequest{}, ResponseBody: RunStartResponse{},
			Errors: []documentation.Error{unauthenticated, csrf, forbidden, paymentRequired, ideaNotFound, tooMany},
		},
		{
			Method: "GET", Path: "/runs/{id}", Summary: "Follow a background run",
			Description: "Answers the run's status (running, waiting while a call waits for approval, finished, stopped, error or interrupted), the last text the agent wrote, the calls that changed data with the id and url their answer named, and the pending approvals. Answer one with POST its decide_path and {\"approve\": true} or false; a bulle.write token may do it, no Bulle UI needed.",
			Auth:        sessionAuth, ResponseBody: RunResponse{},
			Errors: []documentation.Error{unauthenticated, forbidden, runNotFound},
		},
		{
			Method: "POST", Path: "/ideas/generate", Summary: "Think of more ideas",
			Description: "Starts a manual batch for the space. The batch runs in the background with read-only tools. With ?app=, the batch reads only that app's tools, follows that app's guidance and files every idea under it.",
			Auth:        sessionAuth, QueryParams: []documentation.Field{appField, spaceField}, RequestBody: GenerateRequest{}, ResponseBody: GenerateResponse{},
			Errors: []documentation.Error{unauthenticated, csrf, forbidden, badApp, paymentRequired, spaceNotFound, busy, tooMany},
		},
		{
			Method: "PUT", Path: "/ideas/settings", Summary: "Change the daily batch settings",
			Description: "Changes whether the space gets a daily batch, its local hour and timezone, the morning push, and the steering text.",
			Auth:        sessionAuth, RequestBody: SettingsRequest{}, ResponseBody: SettingsResponse{},
			Errors: []documentation.Error{unauthenticated, csrf, badRequest, spaceNotFound},
		},
		{
			Method: "POST", Path: "/ideas/goals", Summary: "Add a goal",
			Description: "Adds a goal to the space. A space holds at most 20 open goals.",
			Auth:        sessionAuth, RequestBody: CreateGoalRequest{}, ResponseBody: GoalResponse{},
			Errors: []documentation.Error{unauthenticated, csrf, badRequest, spaceNotFound},
		},
		{
			Method: "PUT", Path: "/ideas/goals/{id}", Summary: "Change a goal",
			Description: "Renames, recategorises, completes or reopens a goal.",
			Auth:        sessionAuth, RequestBody: UpdateGoalRequest{}, ResponseBody: GoalResponse{},
			Errors: []documentation.Error{unauthenticated, csrf, badRequest, goalNotFound},
		},
		{
			Method: "DELETE", Path: "/ideas/goals/{id}", Summary: "Delete a goal",
			Description: "Deletes a goal.",
			Auth:        sessionAuth, ResponseBody: EmptyResponse{},
			Errors: []documentation.Error{unauthenticated, csrf, goalNotFound},
		},
	},
}
