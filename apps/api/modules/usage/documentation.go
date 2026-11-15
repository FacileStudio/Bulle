package usage

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

// Documentation describes the usage module's routes.
var Documentation = documentation.Module{
	Name:        "usage",
	Description: "Token and cost summary of the user's runs over the last 30 days.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/usage",
			Summary:      "Summarize token use and cost",
			Description:  "Totals for today, the last 7 days and the last 30 days, a per-day series, and breakdowns by model and by channel. Calls other apps make through Bulle's AI count as the Apps channel. budget gives the monthly AI budget on the server's keys: limit_usd (0 means none), spent_usd, resets_at and exempt. Days follow the tz query parameter (IANA name, default UTC).",
			Auth:         "session cookie or bearer token",
			ResponseBody: SummaryResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "bad_request", Description: "tz is not an IANA timezone."},
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
			},
		},
	},
}
