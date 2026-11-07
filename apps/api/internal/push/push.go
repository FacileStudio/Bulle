// Package push delivers generic notifications to a self-hosted ntfy server.
package push

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	clickScheme   = "studio.facile.bulle://"
	topicPrefix   = "bulle-"
	topicHexChars = 32
	sendTimeout   = 5 * time.Second
)

// Config holds the ntfy server and the secret that derives per-user topics.
type Config struct {
	URL         string
	Token       string
	TopicSecret string
}

// Notification is a generic message. Title and Body must never carry model
// output or tool arguments.
type Notification struct {
	Title string
	Body  string
	Path  string
}

// Sender posts notifications to ntfy. A nil Sender is a safe no-op.
type Sender struct {
	cfg    Config
	client *http.Client
	logger *slog.Logger
}

// New returns a Sender, or nil when the URL or the topic secret is empty.
func New(cfg Config, logger *slog.Logger) *Sender {
	cfg.URL = strings.TrimRight(cfg.URL, "/")
	if cfg.URL == "" || cfg.TopicSecret == "" {
		return nil
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Sender{cfg: cfg, client: &http.Client{Timeout: sendTimeout}, logger: logger}
}

// Enabled reports whether notifications are delivered.
func (s *Sender) Enabled() bool { return s != nil }

// Server returns the ntfy server URL without a trailing slash.
func (s *Sender) Server() string {
	if s == nil {
		return ""
	}
	return s.cfg.URL
}

// Topic returns the stable, unguessable topic of a user.
func (s *Sender) Topic(userID int64) string {
	if s == nil {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(s.cfg.TopicSecret))
	mac.Write([]byte(strconv.FormatInt(userID, 10)))
	return topicPrefix + hex.EncodeToString(mac.Sum(nil))[:topicHexChars]
}

// SubscribeURL returns the URL a client subscribes to for a user.
func (s *Sender) SubscribeURL(userID int64) string {
	if s == nil {
		return ""
	}
	return s.cfg.URL + "/" + s.Topic(userID)
}

// Notify posts a notification. Failures are logged and never returned.
func (s *Sender) Notify(ctx context.Context, userID int64, n Notification) {
	if s == nil {
		return
	}
	if err := s.post(ctx, userID, n); err != nil {
		s.logger.Warn("push notification failed", slog.Any("error", err))
	}
}

func (s *Sender) post(ctx context.Context, userID int64, n Notification) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.SubscribeURL(userID), strings.NewReader(n.Body))
	if err != nil {
		return err
	}
	req.Header.Set("Title", n.Title)
	req.Header.Set("Click", clickScheme+n.Path)
	req.Header.Set("Priority", "default")
	if s.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.Token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy answered %d", resp.StatusCode)
	}
	return nil
}
