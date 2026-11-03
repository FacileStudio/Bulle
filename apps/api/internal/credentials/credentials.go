package credentials

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	ErrNoKey          = errors.New("no key for backend")
	ErrUnknownBackend = errors.New("unknown backend")
	ErrDisabled       = errors.New("credential storage is disabled: BULLE_CREDENTIALS_KEY is not set")
)

type Store interface {
	Put(ctx context.Context, userID int64, backend, plaintext string) error
	Delete(ctx context.Context, userID int64, backend string) error
	List(ctx context.Context, userID int64) ([]Entry, error)
	Resolve(ctx context.Context, userID int64, backend string) (string, error)
}

type Entry struct {
	Backend   string
	Last4     string
	UpdatedAt time.Time
}

var envVars = map[string][]string{
	"anthropic":  {"ANTHROPIC_API_KEY"},
	"openai":     {"OPENAI_API_KEY"},
	"google":     {"GEMINI_API_KEY", "GOOGLE_API_KEY"},
	"openrouter": {"OPENROUTER_API_KEY"},
}

var backendOrder = []string{"anthropic", "openai", "google", "openrouter"}

func Backends() []string {
	return append([]string(nil), backendOrder...)
}

func Valid(backend string) bool {
	return len(EnvVars(backend)) > 0
}

func EnvVars(backend string) []string {
	if names, ok := envVars[backend]; ok {
		return append([]string(nil), names...)
	}
	return append([]string(nil), searchEnvVars[backend]...)
}

func EnvKey(backend string) string {
	for _, name := range EnvVars(backend) {
		if value := os.Getenv(name); value != "" {
			return value
		}
	}
	return ""
}

func ParseKey(encoded string) ([]byte, error) {
	encoded = strings.TrimSpace(encoded)
	if encoded == "" {
		return nil, nil
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("BULLE_CREDENTIALS_KEY must be base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("BULLE_CREDENTIALS_KEY must decode to 32 bytes, got %d", len(key))
	}
	return key, nil
}

type row struct {
	Ciphertext []byte
	Nonce      []byte
	Last4      string
	UpdatedAt  time.Time
}

type repo interface {
	upsert(ctx context.Context, userID int64, backend string, r row) error
	remove(ctx context.Context, userID int64, backend string) error
	list(ctx context.Context, userID int64) (map[string]row, error)
	get(ctx context.Context, userID int64, backend string) (row, bool, error)
}

type Service struct {
	repo repo
	aead cipher.AEAD
}

func newService(r repo, masterKey []byte) (*Service, error) {
	s := &Service{repo: r}
	if len(masterKey) == 0 {
		return s, nil
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("credentials key: %w", err)
	}
	s.aead, err = cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("credentials key: %w", err)
	}
	return s, nil
}

func aad(userID int64, backend string) []byte {
	return []byte(strconv.FormatInt(userID, 10) + ":" + backend)
}

func lastFour(secret string) string {
	if utf8.RuneCountInString(secret) < 8 {
		return ""
	}
	runes := []rune(secret)
	return string(runes[len(runes)-4:])
}

func (s *Service) Put(ctx context.Context, userID int64, backend, plaintext string) error {
	if !Valid(backend) {
		return ErrUnknownBackend
	}
	if s.aead == nil {
		return ErrDisabled
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return fmt.Errorf("credentials nonce: %w", err)
	}
	ciphertext := s.aead.Seal(nil, nonce, []byte(plaintext), aad(userID, backend))
	return s.repo.upsert(ctx, userID, backend, row{Ciphertext: ciphertext, Nonce: nonce, Last4: lastFour(plaintext)})
}

func (s *Service) Delete(ctx context.Context, userID int64, backend string) error {
	if !Valid(backend) {
		return ErrUnknownBackend
	}
	return s.repo.remove(ctx, userID, backend)
}

func (s *Service) List(ctx context.Context, userID int64) ([]Entry, error) {
	rows, err := s.repo.list(ctx, userID)
	if err != nil {
		return nil, err
	}
	entries := make([]Entry, 0, len(rows))
	for _, backend := range known() {
		if r, ok := rows[backend]; ok {
			entries = append(entries, Entry{Backend: backend, Last4: r.Last4, UpdatedAt: r.UpdatedAt})
		}
	}
	return entries, nil
}

func (s *Service) Resolve(ctx context.Context, userID int64, backend string) (string, error) {
	if !Valid(backend) {
		return "", ErrUnknownBackend
	}
	if s.aead != nil {
		r, found, err := s.repo.get(ctx, userID, backend)
		if err != nil {
			return "", err
		}
		if found {
			plain, err := s.aead.Open(nil, r.Nonce, r.Ciphertext, aad(userID, backend))
			if err != nil {
				return "", fmt.Errorf("credentials: cannot decrypt stored %s key", backend)
			}
			return string(plain), nil
		}
	}
	if key := EnvKey(backend); key != "" {
		return key, nil
	}
	return "", ErrNoKey
}

type memory struct {
	mu   sync.Mutex
	rows map[string]row
}

func memoryKey(userID int64, backend string) string {
	return string(aad(userID, backend))
}

func NewMemory(masterKey []byte) (*Service, error) {
	return newService(&memory{rows: map[string]row{}}, masterKey)
}

func (m *memory) upsert(_ context.Context, userID int64, backend string, r row) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	r.UpdatedAt = time.Now().UTC()
	m.rows[memoryKey(userID, backend)] = r
	return nil
}

func (m *memory) remove(_ context.Context, userID int64, backend string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rows, memoryKey(userID, backend))
	return nil
}

func (m *memory) list(_ context.Context, userID int64) (map[string]row, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]row{}
	for _, backend := range known() {
		if r, ok := m.rows[memoryKey(userID, backend)]; ok {
			out[backend] = r
		}
	}
	return out, nil
}

func (m *memory) get(_ context.Context, userID int64, backend string) (row, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.rows[memoryKey(userID, backend)]
	return r, ok, nil
}

func (m *memory) swap(userID int64, a, b string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ka, kb := memoryKey(userID, a), memoryKey(userID, b)
	m.rows[ka], m.rows[kb] = m.rows[kb], m.rows[ka]
}
