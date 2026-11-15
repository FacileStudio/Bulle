package plans

import (
	"context"

	"github.com/FacileStudio/porte/forfait"
)

// AIMonthlyBudget is the feature that raises a person's monthly AI budget in
// dollars above the instance default. -1 lifts the budget.
const AIMonthlyBudget = "ai.monthly_budget_usd"

// MonthlyBudget is the monthly AI budget in dollars for the user: the most
// generous of fallback and the ai.monthly_budget_usd of every identity they
// hold, and 0, no budget, when a plan sets -1. It answers fallback before the
// first entitlements snapshot or when the plan cannot be read.
func (g *Gate) MonthlyBudget(ctx context.Context, userID int64, fallback float64) float64 {
	if !g.Active() {
		return fallback
	}
	subjects, err := g.subjects(ctx, userID)
	if err != nil {
		return fallback
	}
	best := fallback
	for _, subject := range subjects {
		limit := g.store.Limit(forfait.Scope{Subject: subject}, AIMonthlyBudget)
		if limit < 0 {
			return 0
		}
		best = max(best, float64(limit))
	}
	return best
}

func (g *Gate) subjects(ctx context.Context, userID int64) ([]string, error) {
	var subjects []string
	if err := g.orm.WithContext(ctx).Raw(`SELECT subject FROM porte_identities WHERE user_id = ? AND provider = ?`,
		userID, g.issuer).Scan(&subjects).Error; err != nil {
		return nil, err
	}
	if len(subjects) == 0 {
		subjects = []string{""}
	}
	return subjects, nil
}
