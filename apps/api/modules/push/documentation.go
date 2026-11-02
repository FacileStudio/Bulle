package push

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

var Documentation = documentation.Module{
	Name:        "push",
	Description: "Push notifications through a self-hosted ntfy server.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/push",
			Summary:      "Fetch the push subscription",
			Description:  "Returns whether push is configured and, if so, the ntfy server, the user's private topic and the URL to subscribe to.",
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
			Description:  "Sends a generic test notification to the user's topic.",
			Auth:         "session cookie or bearer token",
			RequestBody:  EmptyRequest{},
			ResponseBody: EmptyResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
				{Status: 409, Code: "already_exists", Description: "Push is not configured on this server."},
			},
		},
	},
}
