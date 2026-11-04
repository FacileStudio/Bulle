// Package hardware keeps the BOHS-01 devices each user paired and the
// credentials those devices present to the gateway.
package hardware

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// TokenPrefix starts every device token, so a leaked one is recognisable.
const TokenPrefix = "bohs_tok_"

// IDPrefix starts every device id.
const IDPrefix = "hw_"

const (
	TransportBLE  = "ble"
	TransportWiFi = "wifi"
)

// ErrNotFound reports a device that does not exist or belongs to someone else.
var ErrNotFound = errors.New("hardware: device not found")

// Device is one paired BOHS-01 device. LastTransport is empty until the
// device first connects.
type Device struct {
	ID            string
	UserID        int64
	Name          string
	DeviceClass   string
	HardwareID    string
	Capabilities  string
	LastSeenAt    *time.Time
	LastTransport string
	CreatedAt     time.Time
}

// NewID answers a fresh device id: IDPrefix and 32 lowercase hex characters.
func NewID() string {
	return IDPrefix + store.NewID()
}

// MintToken answers a fresh device token and the hash the database keeps.
// The plaintext leaves the server once, in the claim response.
func MintToken() (token, hash string) {
	token = TokenPrefix + store.NewID()
	return token, HashToken(token)
}

// HashToken answers the lowercase SHA-256 hex digest stored for token.
func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// IsToken reports whether value has the shape of a device token.
func IsToken(value string) bool {
	rest, ok := strings.CutPrefix(value, TokenPrefix)
	if !ok || len(rest) != 32 {
		return false
	}
	_, err := hex.DecodeString(rest)
	return err == nil && strings.ToLower(rest) == rest
}
