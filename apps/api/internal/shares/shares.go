// Package shares keeps the public links of Mycelium artifacts: one live
// token per artifact, revoked for good once it is revoked.
package shares

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"gorm.io/gorm"
)

// ErrNotFound is the answer for a token or an artifact with no live link.
var ErrNotFound = errors.New("share not found")

// Share is one public link. Space is the Mycelium tree that holds the artifact.
type Share struct {
	Token      string     `gorm:"column:token;primaryKey"`
	UserID     int64      `gorm:"column:user_id"`
	ArtifactID string     `gorm:"column:artifact_id"`
	Space      string     `gorm:"column:mycelium_space"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
	RevokedAt  *time.Time `gorm:"column:revoked_at"`
}

// TableName names the table the migrations create.
func (Share) TableName() string { return "artifact_shares" }

// Store reads and writes the links in Postgres.
type Store struct{ db *gorm.DB }

// NewStore answers the store over db.
func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

// newToken answers 32 random bytes as 43 base64url characters.
func newToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// Create answers the artifact's live link, making one when it has none. Two
// calls at once still end with a single live link.
func (s *Store) Create(ctx context.Context, userID int64, artifactID, space string) (Share, error) {
	token, err := newToken()
	if err != nil {
		return Share{}, err
	}
	err = s.db.WithContext(ctx).Exec(
		`INSERT INTO artifact_shares (token, user_id, artifact_id, mycelium_space) VALUES (?, ?, ?, ?)
		 ON CONFLICT (user_id, artifact_id, mycelium_space) WHERE revoked_at IS NULL DO NOTHING`,
		token, userID, artifactID, space).Error
	if err != nil {
		return Share{}, err
	}
	return s.Live(ctx, userID, artifactID, space)
}

// Live answers the artifact's live link, in any tree when space is empty, or
// ErrNotFound.
func (s *Store) Live(ctx context.Context, userID int64, artifactID, space string) (Share, error) {
	query := s.db.WithContext(ctx).Where("user_id = ? AND artifact_id = ? AND revoked_at IS NULL", userID, artifactID)
	if space != "" {
		query = query.Where("mycelium_space = ?", space)
	}
	var found Share
	err := query.Order("created_at DESC").Take(&found).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Share{}, ErrNotFound
	}
	return found, err
}

// Revoke ends the artifact's live link, if any. A revoked token never lives again.
func (s *Store) Revoke(ctx context.Context, userID int64, artifactID, space string) error {
	query := s.db.WithContext(ctx).Model(&Share{}).Where("user_id = ? AND artifact_id = ? AND revoked_at IS NULL", userID, artifactID)
	if space != "" {
		query = query.Where("mycelium_space = ?", space)
	}
	return query.Update("revoked_at", time.Now()).Error
}

// Lookup answers the live link a token names, or ErrNotFound.
func (s *Store) Lookup(ctx context.Context, token string) (Share, error) {
	var found Share
	err := s.db.WithContext(ctx).Where("token = ? AND revoked_at IS NULL", token).Take(&found).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Share{}, ErrNotFound
	}
	return found, err
}
