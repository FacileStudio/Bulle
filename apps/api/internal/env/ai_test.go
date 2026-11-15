package env

import "testing"

func TestLoadAIBudget(t *testing.T) {
	cases := []struct {
		name    string
		budget  string
		exempt  string
		want    float64
		wantErr bool
	}{
		{"default", "", "", 5, false},
		{"explicit", "12.5", "", 12.5, false},
		{"disabled", "0", "", 0, false},
		{"negative", "-1", "", 0, true},
		{"garbage", "lots", "", 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv("BULLE_AI_MONTHLY_BUDGET_USD", c.budget)
			t.Setenv("BULLE_AI_BUDGET_EXEMPT", c.exempt)
			ai, err := loadAI()
			if c.wantErr != (err != nil) || ai.MonthlyBudgetUSD != c.want {
				t.Fatalf("budget %v, err %v", ai.MonthlyBudgetUSD, err)
			}
		})
	}
}

func TestLoadAIBudgetExemptEmailsAreLowercased(t *testing.T) {
	t.Setenv("BULLE_AI_BUDGET_EXEMPT", " Owner@Facile.Studio , b@x.test ")
	ai, err := loadAI()
	if err != nil || len(ai.BudgetExempt) != 2 || ai.BudgetExempt[0] != "owner@facile.studio" {
		t.Fatalf("%v, %v", ai.BudgetExempt, err)
	}
}
