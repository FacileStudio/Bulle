package erasure

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/FacileStudio/porte/directory"
	"gorm.io/gorm"
)

type Hooks struct {
	EraseComputer func(ctx context.Context, userID int64) error
	HaltBot       func(userID int64)
	ForgetTokens  func(userID int64)
}

type Deletions struct {
	orm      *gorm.DB
	hooks    Hooks
	receiver *directory.Deletions
}

func New(orm *gorm.DB, client directory.Client, erase func(context.Context, string, string) error, hooks Hooks, logger *slog.Logger) *Deletions {
	d := &Deletions{orm: orm, hooks: hooks}
	d.receiver = &directory.Deletions{
		Client:          client,
		OnPersonDeleted: d.ErasePerson,
		Erase:           erase,
		State:           state{orm: orm},
		Logger:          logger,
	}
	return d
}

func (d *Deletions) Run(ctx context.Context) { d.receiver.Run(ctx) }

func (d *Deletions) Handler() http.Handler { return d.receiver.Handler() }

func (d *Deletions) ErasePerson(ctx context.Context, provider, subject string) error {
	var ids []int64
	err := d.orm.WithContext(ctx).Raw(
		`SELECT user_id FROM porte_identities WHERE provider = ? AND subject = ?`, provider, subject).Scan(&ids).Error
	if err != nil || len(ids) == 0 {
		return err
	}
	userID := ids[0]
	if d.hooks.HaltBot != nil {
		d.hooks.HaltBot(userID)
	}
	if d.hooks.EraseComputer != nil {
		if err := d.hooks.EraseComputer(ctx, userID); err != nil {
			return err
		}
	}
	if d.hooks.ForgetTokens != nil {
		d.hooks.ForgetTokens(userID)
	}
	return d.orm.WithContext(ctx).Exec(`DELETE FROM users WHERE id = ?`, userID).Error
}

type state struct{ orm *gorm.DB }

func (s state) Load(ctx context.Context) ([]byte, error) {
	var rows []struct{ State []byte }
	err := s.orm.WithContext(ctx).Raw(`SELECT state FROM directory_deletion_state WHERE id = 1`).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}
	return rows[0].State, nil
}

func (s state) Save(ctx context.Context, data []byte) error {
	return s.orm.WithContext(ctx).Exec(
		`INSERT INTO directory_deletion_state (id, state, updated_at) VALUES (1, ?, now())
		 ON CONFLICT (id) DO UPDATE SET state = EXCLUDED.state, updated_at = now()`, data).Error
}
