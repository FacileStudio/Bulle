package suite

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
)

const (
	capsuleKeyBytes = 32
	capsuleIVBytes  = 12
	maxCapsuleText  = 64 << 10
	capsuleTimeout  = 15 * time.Second
)

var capsuleExpiries = []string{"1h", "24h", "7d", "30d"}

var capsuleIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,128}$`)

// Sealed is a capsule Bulle created: the link the reader opens, whose part
// after # is the key Capsule never sees, and the token that revokes it.
type Sealed struct {
	ID          string     `json:"id"`
	URL         string     `json:"url"`
	DeleteToken string     `json:"delete_token"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	BurnOnRead  bool       `json:"burn_after_read"`
}

// Sealer encrypts text in Capsule's format on Bulle's side and stores only
// the ciphertext in Capsule. dial is where Bulle reaches Capsule and site is
// the public origin the link points at.
type Sealer struct {
	dial   string
	site   string
	client *http.Client
}

// NewSealer is the sealer for Capsule's configured MCP server, or nil when
// Capsule is not configured.
func NewSealer(server MCPConfig, client *http.Client) *Sealer {
	if !server.Enabled() {
		return nil
	}
	parsed, err := url.Parse(server.URL)
	if err != nil || server.Site() == "" {
		return nil
	}
	if client == nil {
		client = &http.Client{Timeout: capsuleTimeout}
	}
	return &Sealer{dial: parsed.Scheme + "://" + parsed.Host, site: server.Site(), client: client}
}

// Seal encrypts text under a fresh key and stores the ciphertext in Capsule.
func (s *Sealer) Seal(ctx context.Context, text, expiresIn string, burn bool) (Sealed, error) {
	if err := checkCapsuleInput(text, expiresIn); err != nil {
		return Sealed{}, err
	}
	ciphertext, key, err := sealCapsuleText(text)
	if err != nil {
		return Sealed{}, err
	}
	created, err := s.store(ctx, ciphertext, expiresIn, burn)
	if err != nil {
		return Sealed{}, err
	}
	created.URL = s.site + "/" + created.ID + "#" + key
	created.BurnOnRead = burn
	return created, nil
}

func checkCapsuleInput(text, expiresIn string) error {
	switch {
	case strings.TrimSpace(text) == "":
		return errors.New("capsule text is empty")
	case len(text) > maxCapsuleText:
		return fmt.Errorf("capsule text is over %d bytes", maxCapsuleText)
	case expiresIn != "" && !slices.Contains(capsuleExpiries, expiresIn):
		return errors.New("expires_in must be one of 1h, 24h, 7d, 30d")
	}
	return nil
}

// sealCapsuleText is Capsule's browser encryption: AES-256-GCM, the 12-byte
// IV in front of the ciphertext, standard base64, and the raw key in base64url
// without padding.
func sealCapsuleText(text string) (string, string, error) {
	key := make([]byte, capsuleKeyBytes)
	iv := make([]byte, capsuleIVBytes)
	if _, err := rand.Read(key); err != nil {
		return "", "", err
	}
	if _, err := rand.Read(iv); err != nil {
		return "", "", err
	}
	aead, err := capsuleAEAD(key)
	if err != nil {
		return "", "", err
	}
	sealed := aead.Seal(append([]byte{}, iv...), iv, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(sealed), base64.RawURLEncoding.EncodeToString(key), nil
}

func capsuleAEAD(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (s *Sealer) store(ctx context.Context, ciphertext, expiresIn string, burn bool) (Sealed, error) {
	body, err := json.Marshal(map[string]any{"content": ciphertext, "burn_after_read": burn, "expires_in": expiresIn})
	if err != nil {
		return Sealed{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.dial+"/api/pastes", bytes.NewReader(body))
	if err != nil {
		return Sealed{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return Sealed{}, errors.New("capsule could not be reached")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return Sealed{}, fmt.Errorf("capsule refused the capsule (HTTP %d)", resp.StatusCode)
	}
	var created Sealed
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4096)).Decode(&created); err != nil {
		return Sealed{}, errors.New("capsule answered something unreadable")
	}
	if !capsuleIDPattern.MatchString(created.ID) {
		return Sealed{}, errors.New("capsule answered no usable id")
	}
	return created, nil
}
