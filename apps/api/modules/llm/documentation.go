package llm

import (
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
)

var methods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

var Documentation = documentation.Module{
	Name:        "llm",
	Description: "Authenticated proxy to model vendors. The server injects the user's provider key.",
	Routes:      routes(),
}

func routes() []documentation.Route {
	out := make([]documentation.Route, 0, len(methods)+1)
	out = append(out, documentation.Route{
		Method:       "GET",
		Path:         "/llm",
		Summary:      "Model setup",
		Description:  "Returns the backend this instance proxies, its default model, the models a client may switch to, and the user's assistant persona as system prompt text. Clients take their provider settings from here and hold none locally.",
		Auth:         "bearer token",
		ResponseBody: Config{},
		Errors: []documentation.Error{
			{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
		},
	})
	for _, method := range methods {
		out = append(out, documentation.Route{
			Method:       method,
			Path:         "/llm/{backend}/*",
			Summary:      "Proxy a model call",
			Description:  "Forwards the request to the vendor backend (anthropic, openai, google or openrouter) with the user's key and streams the response without buffering. The Bulle token is never forwarded.",
			Auth:         "bearer token",
			RequestBody:  ProxyRequest{},
			ResponseBody: ProxyResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthorized", Description: "The request carries no valid session."},
				{Status: 404, Code: "not_found", Description: "The backend is not supported."},
				{Status: 409, Code: "no_key", Description: "No key is configured for the backend."},
			},
		})
	}
	return out
}
