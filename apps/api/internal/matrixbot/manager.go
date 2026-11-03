// Package matrixbot runs each user's Matrix bot on the server: a teammate in
// every space's linked room that answers only when mentioned, in a thread,
// with one Bulle session per thread.
package matrixbot

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.mau.fi/util/dbutil"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/core/chat"
)

const (
	retryAfter = 30 * time.Second
	maxAge     = 10 * time.Minute
)

// Runner starts a run on a session and signals when it changes.
type Runner interface {
	Start(ctx context.Context, userID int64, sessionID, text string) (string, error)
	Subscribe(runID string) (<-chan struct{}, func())
}

// Config is what the manager needs: the database for its own tables and the
// per-bot crypto schemas, the store and runner a turn goes through, and the
// credentials key that seals passwords.
type Config struct {
	SQL         *sql.DB
	DatabaseURL string
	Store       store.Store
	Runner      Runner
	Key         []byte
	Logger      *slog.Logger
}

// Manager owns every running bot, one per user.
type Manager struct {
	cfg    Config
	repo   repo
	seal   *sealer
	mu     sync.Mutex
	cancel map[int64]context.CancelFunc
	status map[int64]string
	wg     sync.WaitGroup
}

// New builds a manager; a nil Key leaves bots disabled.
func New(cfg Config) (*Manager, error) {
	seal, err := newSealer(cfg.Key)
	if err != nil {
		return nil, err
	}
	return &Manager{cfg: cfg, repo: repo{db: cfg.SQL}, seal: seal, cancel: map[int64]context.CancelFunc{}, status: map[int64]string{}}, nil
}

// Start brings up every enabled bot.
func (m *Manager) Start(ctx context.Context) {
	if m.seal == nil {
		return
	}
	bots, err := m.repo.enabled(ctx)
	if err != nil {
		m.cfg.Logger.Error("matrix: listing bots", slog.Any("error", err))
		return
	}
	for _, bot := range bots {
		m.launch(bot)
	}
}

// Stop shuts every bot down and waits for them.
func (m *Manager) Stop() {
	m.mu.Lock()
	for _, cancel := range m.cancel {
		cancel()
	}
	m.mu.Unlock()
	m.wg.Wait()
}

// reload restarts one user's bot from what is stored.
func (m *Manager) reload(ctx context.Context, userID int64) error {
	m.halt(userID)
	bot, found, err := m.repo.get(ctx, userID)
	if err != nil || !found || !bot.Enabled {
		return err
	}
	m.launch(bot)
	return nil
}

func (m *Manager) halt(userID int64) {
	m.mu.Lock()
	cancel := m.cancel[userID]
	delete(m.cancel, userID)
	m.status[userID] = "off"
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (m *Manager) launch(bot Bot) {
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancel[bot.UserID] = cancel
	m.mu.Unlock()
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.keepRunning(ctx, bot)
	}()
}

func (m *Manager) setStatus(userID int64, status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, live := m.cancel[userID]; live || status == "off" {
		m.status[userID] = status
	}
}

func (m *Manager) statusOf(userID int64) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.status[userID]; ok {
		return s
	}
	return "off"
}

// keepRunning runs the bot and brings it back after a failure, so a
// homeserver that is down for a minute does not leave the room silent.
func (m *Manager) keepRunning(ctx context.Context, bot Bot) {
	for ctx.Err() == nil {
		m.setStatus(bot.UserID, "connecting")
		err := m.runOnce(ctx, bot)
		if ctx.Err() != nil {
			return
		}
		m.setStatus(bot.UserID, "error: "+err.Error())
		m.cfg.Logger.Warn("matrix: bot stopped", slog.Int64("user_id", bot.UserID), slog.Any("error", err))
		select {
		case <-ctx.Done():
		case <-time.After(retryAfter):
		}
	}
}

func (m *Manager) runOnce(ctx context.Context, bot Bot) error {
	password, err := m.seal.open(bot.UserID, bot.Password, bot.Nonce)
	if err != nil {
		return err
	}
	db, err := m.cryptoDB(ctx, bot.UserID)
	if err != nil {
		return err
	}
	adapter, err := chat.NewMatrix(chat.MatrixConfig{
		Homeserver: bot.Homeserver,
		UserID:     bot.User,
		Password:   password,
		PickleKey:  m.seal.pickleKey(bot.UserID),
		Database:   db,
		SelfSign:   true,
	})
	if err != nil {
		return err
	}
	defer func() { _ = adapter.Close() }()
	m.setStatus(bot.UserID, "running")
	err = chat.Run(ctx, adapter, m.responder(bot.UserID), chat.Router{Allow: bot.Allow, MaxAge: maxAge})
	if err == nil {
		err = fmt.Errorf("sync ended")
	}
	return err
}

// cryptoDB opens the bot's own schema, so mautrix's tables for one bot never
// meet another's.
func (m *Manager) cryptoDB(ctx context.Context, userID int64) (*dbutil.Database, error) {
	schema := fmt.Sprintf("matrix_u%d", userID)
	if _, err := m.cfg.SQL.ExecContext(ctx, "CREATE SCHEMA IF NOT EXISTS "+schema); err != nil {
		return nil, fmt.Errorf("matrix: creating %s: %w", schema, err)
	}
	conf, err := pgx.ParseConfig(m.cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("matrix: database url: %w", err)
	}
	conf.RuntimeParams["search_path"] = schema
	return dbutil.NewWithDB(stdlib.OpenDB(*conf), "postgres")
}
