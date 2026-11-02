package suite

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

type StatusResponse struct {
	Apps []suite.AppStatus `json:"apps"`
}

var Documentation = documentation.Module{
	Name:        "suite",
	Description: "The user's access to Courrier, Agenda, Opus and Sablier through Bulle.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/suite/status",
			Summary:      "Connected apps",
			Description:  "Probes each configured suite app with the user's delegated token. State is connected, no_account (the user never signed in to that app), reconnect (sign in to Bulle again) or unavailable.",
			Auth:         "session cookie or bearer token",
			ResponseBody: StatusResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "No valid session."},
			},
		},
	},
}
