package ideas

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

const sessionAuth = "session cookie or bearer token"

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
var tooMany = documentation.Error{Status: 429, Code: "rate_limited", Description: "More than 3 manual batches today in the space's timezone, or too many active runs."}

var Documentation = documentation.Module{
	Name:        "ideas",
	Description: "Bulle's proactive suggestions: a daily batch of idea cards per space, the goals that steer it, and its settings.",
	Routes: []documentation.Route{
		{
			Method: "GET", Path: "/ideas", Summary: "Show the Ideas page",
			Description:  "Lists the space's new and saved ideas (newest batch first), its goals (open first), its settings, whether a batch is running and when the last one finished. Creates the settings row with defaults the first time, so the space starts getting a daily batch.",
			Auth:         sessionAuth,
			QueryParams:  []documentation.Field{{Name: "space_id", Type: "string", Description: "The space; the default space when omitted."}},
			ResponseBody: ListResponse{},
			Errors:       []documentation.Error{unauthenticated, spaceNotFound},
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
			Method: "POST", Path: "/ideas/generate", Summary: "Think of more ideas",
			Description: "Starts a manual batch for the space. The batch runs in the background with read-only tools.",
			Auth:        sessionAuth, RequestBody: GenerateRequest{}, ResponseBody: GenerateResponse{},
			Errors: []documentation.Error{unauthenticated, csrf, spaceNotFound, busy, tooMany},
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
