// Package aiusage records what the AI calls made through Bulle cost and caps
// what each user may spend on the server's provider keys per UTC month.
package aiusage

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/nacelle"
	"gorm.io/gorm"
)

const (
	SourceComplete  = "complete"
	SourceSummarize = "summarize"
	SourceDecide    = "decide"
)

// Call is what one completion used: the model that answered, the usage the
// backend reported and whether the user's own provider key paid for it.
type Call struct {
	Model  string
	Usage  nacelle.Usage
	OwnKey bool
}

// Ledger writes ai_usage rows.
type Ledger struct {
	db     *gorm.DB
	logger *slog.Logger
}

// NewLedger builds the Ledger over db.
func NewLedger(db *gorm.DB, logger *slog.Logger) *Ledger {
	return &Ledger{db: db, logger: logger}
}

// Record writes one row for a call made for userID by caller. A failed write
// is logged, never returned: the user's request has already been served. A nil
// Ledger records nothing.
func (l *Ledger) Record(ctx context.Context, userID int64, source, caller string, call Call) {
	if l == nil {
		return
	}
	u := call.Usage
	err := l.db.WithContext(context.WithoutCancel(ctx)).Exec(
		`INSERT INTO ai_usage (user_id, source, caller, model, input_tokens, output_tokens,
			cache_read_tokens, cache_creation_tokens, cost_usd, own_key)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, source, caller, call.Model, u.InputTokens, u.OutputTokens, u.CacheReadTokens, u.CacheCreationTokens, u.Cost, call.OwnKey,
	).Error
	if err != nil {
		l.logger.Error("recording ai usage", "user", userID, "source", source, "error", err)
	}
}
