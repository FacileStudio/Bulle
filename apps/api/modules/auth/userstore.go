package auth

import (
	"context"
	stderrors "errors"
	"net/mail"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

const registrationLockKey = 0x62756c6c65

type UserStore struct {
	orm *gorm.DB
}

func NewUserStore(orm *gorm.DB) *UserStore {
	return &UserStore{orm: orm}
}

var _ porte.UserStore = (*UserStore)(nil)

func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	if !claims.EmailVerified {
		return 0, errors.Forbidden("the identity provider did not verify this email address")
	}
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return 0, errors.Invalid("the identity provider returned no usable email")
	}
	name := claims.DisplayName()

	var userID int64
	txErr := s.orm.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", int64(registrationLockKey)).Error; err != nil {
			return errors.Internal("failed to acquire registration lock", err)
		}

		var linked int64
		err := tx.Raw(
			`SELECT user_id FROM porte_identities WHERE provider = ? AND subject = ?`,
			claims.Provider, claims.Subject,
		).Scan(&linked).Error
		if err != nil {
			return errors.Internal("failed to resolve the identity", err)
		}
		if linked != 0 {
			userID = linked
			return refreshProfile(tx, linked, email, name)
		}

		var existing schemas.User
		err = tx.Where("email = ?", email).First(&existing).Error
		switch {
		case err == nil:
			userID = existing.ID
			return refreshProfile(tx, existing.ID, email, name)
		case !stderrors.Is(err, gorm.ErrRecordNotFound):
			return errors.Internal("failed to look up the account", err)
		}

		user := schemas.User{Email: email, Name: name}
		if err := tx.Create(&user).Error; err != nil {
			return errors.Internal("failed to create the account", err)
		}
		userID = user.ID
		return nil
	})
	if txErr != nil {
		return 0, txErr
	}
	return userID, nil
}

func refreshProfile(tx *gorm.DB, userID int64, email, name string) error {
	updates := map[string]any{"email": email}
	if name != "" {
		updates["name"] = name
	}
	if err := tx.Model(&schemas.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return errors.Internal("failed to update the account", err)
	}
	return nil
}
