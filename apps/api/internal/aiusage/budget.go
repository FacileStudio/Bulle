package aiusage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	troncerrors "github.com/FacileStudio/tronc/errors"
	"gorm.io/gorm"
)

// ErrBudget is the cause of every refusal Budget.Check gives.
var ErrBudget = errors.New("aiusage: the monthly AI allowance is used up")

// Status is where a user stands against their monthly budget. A LimitUSD of 0
// means no budget applies.
type Status struct {
	LimitUSD float64   `json:"limit_usd"`
	SpentUSD float64   `json:"spent_usd"`
	ResetsAt time.Time `json:"resets_at"`
	Exempt   bool      `json:"exempt"`
}

// Budget caps each user's spend on the server's provider keys per UTC month.
// Plan, when set, may raise the limit for a user above DefaultUSD.
type Budget struct {
	DB         *gorm.DB
	DefaultUSD float64
	Exempt     []string
	Plan       func(ctx context.Context, userID int64, fallback float64) float64
	Now        func() time.Time
}

// MonthStart is the first instant of the UTC month holding now.
func MonthStart(now time.Time) time.Time {
	now = now.UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// Status reads the user's limit and what they spent this month on the server's
// keys: the runs of their sessions plus the calls apps made for them, never
// the ones that used their own key. A nil Budget has no limit.
func (b *Budget) Status(ctx context.Context, userID int64) (Status, error) {
	if b == nil {
		return Status{}, nil
	}
	start := MonthStart(b.Now())
	status := Status{LimitUSD: b.DefaultUSD, ResetsAt: start.AddDate(0, 1, 0)}
	if b.Plan != nil {
		status.LimitUSD = b.Plan(ctx, userID, b.DefaultUSD)
	}
	var rows []struct {
		Email string
		Spent float64
	}
	err := b.DB.WithContext(ctx).Raw(
		`SELECT u.email,
			(SELECT COALESCE(SUM(r.cost_usd), 0) FROM runs r JOIN sessions s ON s.id = r.session_id
				WHERE s.user_id = u.id AND NOT r.own_key AND r.started_at >= @since)::float8
			+ (SELECT COALESCE(SUM(a.cost_usd), 0) FROM ai_usage a
				WHERE a.user_id = u.id AND NOT a.own_key AND a.created_at >= @since)::float8 AS spent
		FROM users u WHERE u.id = @user`,
		map[string]any{"since": start, "user": userID},
	).Scan(&rows).Error
	if err != nil {
		return Status{}, err
	}
	if len(rows) == 0 {
		return Status{}, gorm.ErrRecordNotFound
	}
	status.SpentUSD = rows[0].Spent
	for _, email := range b.Exempt {
		status.Exempt = status.Exempt || strings.EqualFold(email, rows[0].Email)
	}
	return status, nil
}

// Check answers nil when userID may spend more on the server's keys, and the
// 429 refusal naming the reset date when the month's budget is used up. A nil
// Budget refuses nobody.
func (b *Budget) Check(ctx context.Context, userID int64) error {
	if b == nil {
		return nil
	}
	status, err := b.Status(ctx, userID)
	if err != nil {
		return troncerrors.Internal("failed to read the AI allowance", err)
	}
	if status.Exempt || status.LimitUSD <= 0 || status.SpentUSD < status.LimitUSD {
		return nil
	}
	return troncerrors.New("rate_limited", fmt.Sprintf("This month's AI allowance is used up. It resets on %s.", status.ResetsAt.Format("2 January")), ErrBudget)
}
