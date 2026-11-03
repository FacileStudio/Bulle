package usage

// Totals sums the spend of a set of runs.
type Totals struct {
	Runs                int     `json:"runs"`
	InputTokens         int64   `json:"input_tokens"`
	OutputTokens        int64   `json:"output_tokens"`
	CacheReadTokens     int64   `json:"cache_read_tokens"`
	CacheCreationTokens int64   `json:"cache_creation_tokens"`
	CostUSD             float64 `json:"cost_usd"`
}

// Day is one calendar day of spend in the requested timezone.
type Day struct {
	Date string `json:"date"`
	Totals
}

// Group is the spend of one model or one channel.
type Group struct {
	Name string `json:"name"`
	Totals
}

// SummaryResponse is the usage summary for the authenticated user.
type SummaryResponse struct {
	Timezone string  `json:"timezone"`
	Today    Totals  `json:"today"`
	Week     Totals  `json:"week"`
	Month    Totals  `json:"month"`
	Days     []Day   `json:"days"`
	Models   []Group `json:"models"`
	Channels []Group `json:"channels"`
}
