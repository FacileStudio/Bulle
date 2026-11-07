package computers

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
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

// loadBoxKey resolves BULLE_BOX_KEY in the shapes a deployment stores it in: the
// PEM itself, the base64 of it on one line — which is how docs/deploy.md says to
// set it, because a multi-line PEM does not survive a KEY=value env file — or a
// path to either on disk. An empty value generates a fresh ed25519 pair.
func loadBoxKey(value string) (boxKey, error) {
	text := strings.TrimSpace(value)
	if text == "" {
		return generateBoxKey()
	}
	if material, ok := readKeyText(text); ok {
		return stageBoxKey(material)
	}
	content, err := os.ReadFile(text)
	if err != nil {
		return boxKey{}, fmt.Errorf("the box key is not a private key, the base64 of one, or a file that exists: %w", err)
	}
	if material, ok := readKeyText(string(content)); ok {
		return stageBoxKey(material)
	}
	return boxKey{}, fmt.Errorf("the file %q holds no private key", text)
}

// readKeyText recognizes both encodings of a private key and returns the PEM.
func readKeyText(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "-----BEGIN") && strings.Contains(text, "PRIVATE KEY") {
		return text, true
	}
	// A base64 blob is one line, but a deployment that wrapped it is read the
	// same way: whitespace carries no meaning in base64.
	compact := strings.Join(strings.Fields(text), "")
	if compact == "" || strings.ContainsAny(compact, "-#") {
		return "", false
	}
	for _, encoding := range []*base64.Encoding{
		base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding,
	} {
		decoded, err := encoding.DecodeString(compact)
		if err != nil {
			continue
		}
		if material := strings.TrimSpace(string(decoded)); strings.HasPrefix(material, "-----BEGIN") && strings.Contains(material, "PRIVATE KEY") {
			return material, true
		}
	}
	return "", false
}

// stageBoxKey validates a private key and writes it where ssh can read it.
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

func stagePrivateKey(text string) (path string, err error) {
	file, err := os.CreateTemp("", "bulle-box-key-*")
	if err != nil {
		return "", fmt.Errorf("failed to stage the box key: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			path, err = "", fmt.Errorf("failed to stage the box key: %w", closeErr)
		}
	}()
	if err := file.Chmod(0o600); err != nil {
		return "", fmt.Errorf("failed to protect the box key: %w", err)
	}
	if _, err := file.WriteString(strings.TrimSpace(text) + "\n"); err != nil {
		return "", fmt.Errorf("failed to stage the box key: %w", err)
	}
	return file.Name(), nil
}
