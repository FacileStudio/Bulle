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
	orm  *gorm.DB
	apps []string
}

// NewUserStore seeds each new account with the starter jobs, including the
// ones tied to a suite app only when that app is in apps.
func NewUserStore(orm *gorm.DB, apps []string) *UserStore {
	return &UserStore{orm: orm, apps: apps}
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
	picture := pictureURL(claims.Picture)

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
			return refreshProfile(tx, linked, email, name, picture)
		}

		var existing schemas.User
		err = tx.Where("email = ?", email).First(&existing).Error
		switch {
		case err == nil:
			userID = existing.ID
			return refreshProfile(tx, existing.ID, email, name, picture)
		case !stderrors.Is(err, gorm.ErrRecordNotFound):
			return errors.Internal("failed to look up the account", err)
		}

		user := schemas.User{Email: email, Name: name, PictureURL: picture}
		if err := tx.Create(&user).Error; err != nil {
			return errors.Internal("failed to create the account", err)
		}
		userID = user.ID
		return seedStarterJobs(tx, user.ID, s.apps)
	})
	if txErr != nil {
		return 0, txErr
	}
	return userID, nil
}

func refreshProfile(tx *gorm.DB, userID int64, email, name, picture string) error {
	updates := map[string]any{"email": email, "picture_url": picture}
	if name != "" {
		updates["name"] = name
	}
	if err := tx.Model(&schemas.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		return errors.Internal("failed to update the account", err)
	}
	return nil
}

// pictureURL keeps the picture claim only when it is an https URL. Registre emits a
// stable https://<issuer>/avatars/<sub> for every human, so anything else is not a
// picture this app should render.
func pictureURL(claim string) string {
	if strings.HasPrefix(strings.ToLower(claim), "https://") {
		return claim
	}
	return ""
}

// BackfillPictures gives every user who signed in before pictures were stored the
// stable Registre avatar URL derived from their identity row. It is idempotent: it only
// touches users whose picture_url is still empty, and does nothing when the issuer is
// not https.
func BackfillPictures(ctx context.Context, orm *gorm.DB, issuer string) (int64, error) {
	base := strings.TrimRight(issuer, "/")
	if pictureURL(base) == "" {
		return 0, nil
	}
	result := orm.WithContext(ctx).Exec(
		`UPDATE users SET picture_url = ? || '/avatars/' || porte_identities.subject
		FROM porte_identities
		WHERE porte_identities.user_id = users.id AND porte_identities.provider = ? AND users.picture_url = ''`,
		base, issuer,
	)
	if result.Error != nil {
		return 0, errors.Internal("failed to backfill profile pictures", result.Error)
	}
	return result.RowsAffected, nil
}
