package env

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/FacileStudio/Bulle/core/jev"
)

// AI is what POST /api/ai/decide runs on: the decision model, whose key still
// comes from TYPESAFE_API_KEY (or OPENROUTER_API_KEY on the decisions
// endpoint). MonthlyBudgetUSD caps each user's spend on the server's provider
// keys per UTC month (0 disables it), and BudgetExempt lists the emails the cap
// never applies to.
type AI struct {
	Jev              jev.Config
	MonthlyBudgetUSD float64
	BudgetExempt     []string
}

const defaultMonthlyBudgetUSD = 5.0

func loadAI() (AI, error) {
	cfg := jev.Config{
		Endpoint: jev.Endpoint(strings.TrimSpace(os.Getenv("JEV_ENDPOINT"))),
		BaseURL:  strings.TrimSpace(os.Getenv("JEV_BASE_URL")),
		Model:    strings.TrimSpace(os.Getenv("JEV_MODEL")),
	}
	switch cfg.Endpoint {
	case "", jev.EndpointSystemOne, jev.EndpointDecisions:
	default:
		return AI{}, fmt.Errorf("JEV_ENDPOINT must be systemone or decisions")
	}
	if cfg.BaseURL != "" {
		parsed, err := url.Parse(cfg.BaseURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return AI{}, fmt.Errorf("JEV_BASE_URL must be an http or https URL")
		}
	}
	budget, err := monthlyBudget(os.Getenv("BULLE_AI_MONTHLY_BUDGET_USD"))
	if err != nil {
		return AI{}, err
	}
	exempt := csvOrDefault("BULLE_AI_BUDGET_EXEMPT", nil)
	for i, email := range exempt {
		exempt[i] = strings.ToLower(email)
	}
	return AI{Jev: cfg, MonthlyBudgetUSD: budget, BudgetExempt: exempt}, nil
}

func monthlyBudget(raw string) (float64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return defaultMonthlyBudgetUSD, nil
	}
	budget, err := strconv.ParseFloat(raw, 64)
	if err != nil || budget < 0 || math.IsNaN(budget) || math.IsInf(budget, 0) {
		return 0, fmt.Errorf("BULLE_AI_MONTHLY_BUDGET_USD must be a number of dollars, 0 or more")
	}
	return budget, nil
}
