package matrixbot

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"
)

// ErrDisabled is returned when the server has no BULLE_CREDENTIALS_KEY, so
// it cannot keep a bot password at rest.
var ErrDisabled = errors.New("matrix bots need BULLE_CREDENTIALS_KEY on the server")

type sealer struct {
	aead   cipher.AEAD
	master []byte
}

func newSealer(master []byte) (*sealer, error) {
	if len(master) == 0 {
		return nil, nil
	}
	block, err := aes.NewCipher(master)
	if err != nil {
		return nil, fmt.Errorf("matrix key: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("matrix key: %w", err)
	}
	return &sealer{aead: aead, master: master}, nil
}

func aad(userID int64) []byte {
	return []byte("matrix:" + strconv.FormatInt(userID, 10))
}

func (s *sealer) seal(userID int64, plain string) ([]byte, []byte, error) {
	if s == nil {
		return nil, nil, ErrDisabled
	}
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	return s.aead.Seal(nil, nonce, []byte(plain), aad(userID)), nonce, nil
}

func (s *sealer) open(userID int64, sealed, nonce []byte) (string, error) {
	if s == nil {
		return "", ErrDisabled
	}
	plain, err := s.aead.Open(nil, nonce, sealed, aad(userID))
	if err != nil {
		return "", fmt.Errorf("matrix password: %w", err)
	}
	return string(plain), nil
}

// pickleKey derives the key that encrypts one bot's olm account at rest, so
// the crypto tables are useless without the server's credentials key.
func (s *sealer) pickleKey(userID int64) []byte {
	mac := hmac.New(sha256.New, s.master)
	mac.Write(aad(userID))
	return mac.Sum(nil)
}
