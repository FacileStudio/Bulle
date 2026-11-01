package computers

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

// boxKey is the keypair the API logs into a box with: Path is the private half
// handed to `ssh -i`, Public is the one-line authorized_keys entry the box has
// to trust.
//
// It is a pair, not a path, because boite authorizes only the key it generates
// for the instance plus the extra keys it is handed at create time (or through
// BOITE_CONTROL_PUBKEY on the worker). An API that shows up with a private key
// nobody authorized gets `Permission denied (publickey,password)` from every
// connection, and the box's own generated key lives on the worker's disk where
// the API container cannot reach it.
type boxKey struct {
	Path   string
	Public string
}

// loadBoxKey resolves BULLE_BOX_KEY. An inline PEM is staged to a 0600 file, a
// path is read from disk, and an empty value generates a fresh ed25519 pair, so
// an API deployed without the secret still opens its own boxes instead of
// failing every run at the preflight probe.
func loadBoxKey(value string) (boxKey, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return generateBoxKey()
	}
	if !strings.Contains(text, "PRIVATE KEY") {
		content, err := os.ReadFile(text)
		if err != nil {
			return boxKey{}, fmt.Errorf("failed to read the box key %q: %w", text, err)
		}
		text = strings.TrimSpace(string(content))
	}
	return stageBoxKey(text)
}

// stageBoxKey validates a private key and writes it where ssh can read it. An
// unreadable key is an error rather than a silent fallback: starting with a key
// that does not parse only trades a clear failure at boot for a confusing one
// per run.
func stageBoxKey(text string) (boxKey, error) {
	signer, err := ssh.ParsePrivateKey([]byte(text))
	if err != nil {
		return boxKey{}, fmt.Errorf("failed to parse the box key: %w", err)
	}
	path, err := stagePrivateKey(text)
	if err != nil {
		return boxKey{}, err
	}
	return boxKey{Path: path, Public: publicKeyLine(signer.PublicKey())}, nil
}

func generateBoxKey() (boxKey, error) {
	_, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return boxKey{}, fmt.Errorf("failed to generate the box key: %w", err)
	}
	block, err := ssh.MarshalPrivateKey(private, "")
	if err != nil {
		return boxKey{}, fmt.Errorf("failed to encode the box key: %w", err)
	}
	return stageBoxKey(string(pem.EncodeToMemory(block)))
}

func publicKeyLine(key ssh.PublicKey) string {
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
}

func stagePrivateKey(text string) (string, error) {
	file, err := os.CreateTemp("", "bulle-box-key-*")
	if err != nil {
		return "", fmt.Errorf("failed to stage the box key: %w", err)
	}
	defer file.Close()
	if err := file.Chmod(0o600); err != nil {
		return "", fmt.Errorf("failed to protect the box key: %w", err)
	}
	if _, err := file.WriteString(strings.TrimSpace(text) + "\n"); err != nil {
		return "", fmt.Errorf("failed to stage the box key: %w", err)
	}
	return file.Name(), nil
}
