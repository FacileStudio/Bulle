package spaces

import (
	"context"
	"errors"
	"log/slog"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/porte/directory"
	"gorm.io/gorm"
)

// Registre mirrors Registre's shared spaces into Bulle. Bulle spaces belong to
// one user each, so a shared space becomes one Bulle space per member, linked
// by registre_space_id. The user's own unlinked spaces, Personal first, are
// never touched.
type Registre struct {
	db     *gorm.DB
	issuer string
	logger *slog.Logger
}

// NewRegistre builds the sync for the identities Bulle records under issuer.
func NewRegistre(db *gorm.DB, issuer string, logger *slog.Logger) *Registre {
	return &Registre{db: db, issuer: issuer, logger: logger}
}

type membership struct {
	userID     int64
	registreID string
}

// NewDirectory builds the Registre space sync with Bulle's own client
// credentials, or answers nil when OIDC is not configured.
func NewDirectory(db *gorm.DB, issuer, clientID, clientSecret string, logger *slog.Logger, entitlements ...func(context.Context, directory.Snapshot) error) *directory.Syncer {
	if issuer == "" || clientID == "" || clientSecret == "" {
		return nil
	}
	apply := NewRegistre(db, issuer, logger).Apply
	return &directory.Syncer{
		Client: directory.Client{Issuer: issuer, ClientID: clientID, ClientSecret: clientSecret},
		Apply: func(ctx context.Context, snapshot directory.Snapshot) error {
			errs := []error{apply(ctx, snapshot)}
			for _, entitle := range entitlements {
				errs = append(errs, entitle(ctx, snapshot))
			}
			return errors.Join(errs...)
		},
		Logger: logger,
	}
}

// Apply makes every member's linked spaces match the snapshot, in one
// transaction and idempotently. A member is found by Registre subject, then by
// email, else pre-provisioned so their first sign-in already shows the space.
// A linked space the user no longer belongs to hands its chats and jobs to the
// user's default space before it is removed, so leaving loses no history.
func (r *Registre) Apply(ctx context.Context, snapshot directory.Snapshot) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		keep := map[membership]bool{}
		for _, space := range snapshot.Spaces {
			if err := r.linkMembers(tx, space, keep); err != nil {
				return err
			}
		}
		return prune(tx, keep)
	})
}

// linkMembers gives every member of one shared space its linked Bulle space
// and records the pair in keep.
func (r *Registre) linkMembers(tx *gorm.DB, space directory.Space, keep map[membership]bool) error {
	for _, member := range space.Members {
		userID, err := r.resolveMember(tx, member)
		if err != nil {
			return err
		}
		if userID == 0 {
			continue
		}
		if err := linkSpace(tx, userID, space); err != nil {
			return err
		}
		keep[membership{userID, space.ID}] = true
	}
	return nil
}

// linkSpace gives the user a space linked to Registre's, renaming it when
// Registre did, adopting an unlinked space of the same name before creating
// one, and making sure the user's default space exists first so a linked space
// never becomes the default.
func linkSpace(tx *gorm.DB, userID int64, space directory.Space) error {
	err := tx.Exec(`UPDATE spaces SET name = ?, updated_at = now()
		WHERE user_id = ? AND registre_space_id = ? AND name <> ?`, space.Name, userID, space.ID, space.Name).Error
	if err != nil {
		return err
	}
	var linked int64
	if err := tx.Raw(`SELECT count(*) FROM spaces WHERE user_id = ? AND registre_space_id = ?`,
		userID, space.ID).Scan(&linked).Error; err != nil || linked > 0 {
		return err
	}
	if _, err := defaultSpaceID(tx, userID); err != nil {
		return err
	}
	adopted := tx.Exec(`UPDATE spaces SET registre_space_id = ?, name = ?, updated_at = now()
		WHERE id = (SELECT id FROM spaces WHERE user_id = ? AND registre_space_id = '' AND lower(name) = lower(?)
			AND id <> (SELECT id FROM spaces WHERE user_id = ? ORDER BY created_at, id LIMIT 1)
			ORDER BY created_at, id LIMIT 1)`, space.ID, space.Name, userID, space.Name, userID)
	if adopted.Error != nil || adopted.RowsAffected > 0 {
		return adopted.Error
	}
	return tx.Exec(`INSERT INTO spaces (id, user_id, name, registre_space_id) VALUES (?, ?, ?, ?)`,
		store.NewID(), userID, space.Name, space.ID).Error
}

// defaultSpaceID answers the user's oldest unlinked space, creating Personal
// when there is none, dated before every other space so it stays the default.
func defaultSpaceID(tx *gorm.DB, userID int64) (string, error) {
	var ids []string
	err := tx.Raw(`SELECT id FROM spaces WHERE user_id = ? AND registre_space_id = '' ORDER BY created_at, id LIMIT 1`,
		userID).Scan(&ids).Error
	if err != nil || len(ids) > 0 {
		return first(ids), err
	}
	id := store.NewID()
	return id, tx.Exec(`INSERT INTO spaces (id, user_id, name, created_at, updated_at)
		SELECT ?, ?, ?, coalesce(min(created_at), now()) - interval '1 second', now() FROM spaces WHERE user_id = ?`,
		id, userID, store.DefaultSpaceName, userID).Error
}

// prune removes linked spaces whose user left the Registre space, moving their
// chats and jobs to the user's default space first.
func prune(tx *gorm.DB, keep map[membership]bool) error {
	var rows []struct {
		ID              string
		UserID          int64
		RegistreSpaceID string
	}
	if err := tx.Raw(`SELECT id, user_id, registre_space_id FROM spaces WHERE registre_space_id <> ''`).
		Scan(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if keep[membership{row.UserID, row.RegistreSpaceID}] {
			continue
		}
		if err := retire(tx, row.UserID, row.ID); err != nil {
			return err
		}
	}
	return nil
}

func retire(tx *gorm.DB, userID int64, spaceID string) error {
	target, err := defaultSpaceID(tx, userID)
	if err != nil {
		return err
	}
	for _, table := range []string{"sessions", "jobs"} {
		if err := tx.Exec("UPDATE "+table+" SET space_id = ? WHERE space_id = ?", target, spaceID).Error; err != nil {
			return err
		}
	}
	return tx.Exec(`DELETE FROM spaces WHERE id = ?`, spaceID).Error
}
