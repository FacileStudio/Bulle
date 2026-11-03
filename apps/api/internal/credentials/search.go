package credentials

var searchEnvVars = map[string][]string{
	"parallel": {"PARALLEL_API_KEY"},
	"brave":    {"BRAVE_SEARCH_API_KEY"},
	"tavily":   {"TAVILY_API_KEY"},
	"exa":      {"EXA_API_KEY"},
}

var searchOrder = []string{"parallel", "brave", "tavily", "exa"}

// SearchEngines lists the web search engines a user or the instance can hold a
// key for. They are kept apart from Backends so they never reach the model
// picker.
func SearchEngines() []string {
	return append([]string(nil), searchOrder...)
}

func known() []string {
	return append(Backends(), searchOrder...)
}
