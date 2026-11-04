package hardware_test

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/coder/websocket"
)

func TestMintTokenStoresOnlyTheSHA256(t *testing.T) {
	token, hash := hardware.MintToken()
	if !strings.HasPrefix(token, hardware.TokenPrefix) || len(token) != len(hardware.TokenPrefix)+32 {
		t.Fatalf("token %q does not have the bohs_tok_<32 hex> shape", token)
	}
	if !hardware.IsToken(token) {
		t.Fatalf("IsToken(%q) = false", token)
	}
	sum := sha256.Sum256([]byte(token))
	if hash != hex.EncodeToString(sum[:]) || strings.Contains(hash, token) {
		t.Fatalf("hash %q is not the SHA-256 hex of the token", hash)
	}
	other, _ := hardware.MintToken()
	if other == token {
		t.Fatal("two minted tokens are equal")
	}
	for _, bad := range []string{"", "bohs_tok_", "bohs_tok_" + strings.Repeat("A", 32), "tok_" + strings.Repeat("a", 32)} {
		if hardware.IsToken(bad) {
			t.Errorf("IsToken(%q) = true", bad)
		}
	}
}

type closer struct{ code atomic.Int32 }

func (c *closer) Close(code websocket.StatusCode, _ string) error {
	c.code.Store(int32(code))
	return nil
}

func TestRegistryClosesOnlyTheRevokedDevice(t *testing.T) {
	registry := hardware.NewRegistry()
	revoked, kept, gone := &closer{}, &closer{}, &closer{}
	registry.Add("hw_a", revoked)
	registry.Add("hw_b", kept)
	registry.Add("hw_a", gone)()
	if n := registry.Close("hw_a"); n != 1 {
		t.Fatalf("closed %d sockets, want 1", n)
	}
	deadline := time.Now().Add(time.Second)
	for revoked.code.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := websocket.StatusCode(revoked.code.Load()); got != hardware.StatusRevoked {
		t.Fatalf("close code = %d, want %d", got, hardware.StatusRevoked)
	}
	if kept.code.Load() != 0 || gone.code.Load() != 0 {
		t.Fatal("a socket of another device, or an untracked one, was closed")
	}
}

func TestRegistryReplacesTheOldestSocketPastTheCap(t *testing.T) {
	registry := hardware.NewRegistry()
	sockets := make([]*closer, hardware.MaxSocketsPerDevice+1)
	for i := range sockets {
		sockets[i] = &closer{}
		registry.Add("hw_a", sockets[i])
	}
	deadline := time.Now().Add(time.Second)
	for sockets[0].code.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if got := websocket.StatusCode(sockets[0].code.Load()); got != websocket.StatusPolicyViolation {
		t.Fatalf("oldest socket close code = %d, want %d", got, websocket.StatusPolicyViolation)
	}
	if n := registry.Close("hw_a"); n != hardware.MaxSocketsPerDevice {
		t.Fatalf("registry kept %d sockets, want %d", n, hardware.MaxSocketsPerDevice)
	}
}
