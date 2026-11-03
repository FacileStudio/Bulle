package push

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "push",
	Description: "Push notifications: native Android through FCM, browsers through Web Push, desktop apps through a live event stream, and an optional self-hosted ntfy topic.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/push",
			Summary:      "Fetch the push subscription",
			Description:  "Returns which channels are on: native (FCM), web_push_key (the VAPID public key browsers subscribe with) and, when ntfy is on, the server, the user's private topic and the URL to subscribe to.",
			Auth:         "session cookie or bearer token",
			ResponseBody: StatusResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
			},
		},
		{
			Method:       "POST",
			Path:         "/push/test",
			Summary:      "Send a test notification",
			Description:  "Sends a generic test notification to every channel and device of the user.",
			Auth:         "session cookie or bearer token",
			RequestBody:  EmptyRequest{},
			ResponseBody: EmptyResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
				{Status: 409, Code: "already_exists", Description: "Push is not configured on this server."},
			},
		},
		{
			Method:       "GET",
			Path:         "/push/stream",
			Summary:      "Stream the user's live events",
			Description:  "Server-sent events for every open app of the user: `notification` (title, body, path), `session.created` and `session.updated` (session_id, space_id, title, channel, updated_at, status), `session.deleted` (session_id), `run.started` and `run.finished` (session_id, run_id, channel, status on finish, session_status), `approval.pending` and `approval.decided` (session_id, run_id, call_id, tool or approved, session_status). A comment line every 15 seconds keeps proxies open. The server ends the stream when the app falls behind, and the app refetches its sessions on every reconnect.",
			Auth:         "session cookie or bearer token",
			ResponseBody: StreamNotification{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
			},
		},
		{
			Method:       "POST",
			Path:         "/push/devices",
			Summary:      "Register a device for notifications",
			Description:  "Records an FCM token (kind fcm) or a browser Web Push subscription (kind web, with p256dh and auth). Registering a token another account owns moves it to the caller.",
			Auth:         "session cookie or bearer token",
			RequestBody:  DeviceRequest{},
			ResponseBody: EmptyResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid", Description: "The token or a browser key is missing."},
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
				{Status: 409, Code: "already_exists", Description: "This server cannot send to that kind of device."},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/push/devices",
			Summary:      "Forget a device",
			Description:  "Stops notifying one of the caller's devices.",
			Auth:         "session cookie or bearer token",
			RequestBody:  ForgetDeviceRequest{},
			ResponseBody: EmptyResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid", Description: "The token is missing."},
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
			},
		},
	},
}
