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

	"github.com/FacileStudio/Bulle/apps/api/internal/disclosure"
	"github.com/FacileStudio/Bulle/apps/api/internal/prisme"
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
// per-bot crypto schemas, the store and runner a turn goes through, the
// credentials key that seals agent tokens, and the Prisme instance that
// provisions the agents.
type Config struct {
	SQL         *sql.DB
	DatabaseURL string
	Store       store.Store
	Runner      Runner
	Key         []byte
	Prisme      *prisme.Client
	Screen      *disclosure.Screen
	Logger      *slog.Logger
}

// Manager owns every running bot, one per user.
type Manager struct {
	cfg      Config
	repo     repo
	seal     *sealer
	mu       sync.Mutex
	cancel   map[int64]context.CancelFunc
	status   map[int64]string
	stopLoop context.CancelFunc
	wg       sync.WaitGroup
}

// New builds a manager; a nil Key leaves bots disabled.
func New(cfg Config) (*Manager, error) {
	seal, err := newSealer(cfg.Key)
	if err != nil {
		return nil, err
	}
	return &Manager{cfg: cfg, repo: repo{db: cfg.SQL}, seal: seal, cancel: map[int64]context.CancelFunc{}, status: map[int64]string{}}, nil
}

// Start provisions every user's agent in Prisme and brings its bot up, then
// keeps them in step. Without Prisme or a credentials key there is no bot.
func (m *Manager) Start() {
	if m.seal == nil || m.cfg.Prisme == nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.stopLoop = cancel
	m.mu.Unlock()
	m.wg.Add(1)
	go m.provisionLoop(ctx)
}

// Stop shuts every bot down and waits for them.
func (m *Manager) Stop() {
	m.mu.Lock()
	if m.stopLoop != nil {
		m.stopLoop()
	}
	for _, cancel := range m.cancel {
		cancel()
	}
	m.mu.Unlock()
	m.wg.Wait()
}

func (m *Manager) running(userID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, live := m.cancel[userID]
	return live
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
	token, err := m.seal.open(bot.UserID, bot.Token, bot.Nonce)
	if err != nil {
		return err
	}
	db, err := m.cryptoDB(ctx, bot.UserID, bot.DeviceID)
	if err != nil {
		return err
	}
	defer func() { _ = db.RawDB.Close() }()
	if err := m.dropStale(ctx, bot.UserID, bot.DeviceID); err != nil {
		m.cfg.Logger.Warn("matrix: dropping stale crypto stores", slog.Int64("user_id", bot.UserID), slog.Any("error", err))
	}
	adapter, err := chat.NewMatrix(chat.MatrixConfig{
		Homeserver: bot.Homeserver,
		UserID:     bot.User,
		Token:      token,
		DeviceID:   bot.DeviceID,
		PickleKey:  m.seal.pickleKey(bot.UserID),
		Database:   db,
	})
	if err != nil {
		return err
	}
	defer func() { _ = adapter.Close() }()
	m.setStatus(bot.UserID, "running")
	err = chat.Run(ctx, adapter, m.responder(bot, adapter), chat.Router{Open: true, MaxAge: maxAge})
	if err == nil {
		err = fmt.Errorf("sync ended")
	}
	return err
}
