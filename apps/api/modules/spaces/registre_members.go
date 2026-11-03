package spaces

import (
	"log/slog"
	"strings"

	"github.com/FacileStudio/porte/directory"
	"gorm.io/gorm"
)

// resolveMember answers the Bulle user for a snapshot member, or 0 to skip
// one whose email belongs to a user already bound to another subject.
func (r *Registre) resolveMember(tx *gorm.DB, member directory.Member) (int64, error) {
	if member.Subject == "" {
		return 0, nil
	}
	var ids []int64
	if err := tx.Raw(`SELECT user_id FROM porte_identities WHERE provider = ? AND subject = ?`,
		r.issuer, member.Subject).Scan(&ids).Error; err != nil || len(ids) > 0 {
		return firstID(ids), err
	}
	email := strings.ToLower(strings.TrimSpace(member.Email))
	if email == "" {
		return 0, nil
	}
	if err := tx.Raw(`SELECT id FROM users WHERE lower(email) = ?`, email).Scan(&ids).Error; err != nil {
		return 0, err
	}
	userID := firstID(ids)
	if userID != 0 {
		var other int64
		if err := tx.Raw(`SELECT count(*) FROM porte_identities WHERE user_id = ? AND provider = ?`,
			userID, r.issuer).Scan(&other).Error; err != nil || other > 0 {
			r.logger.Warn("directory: skipped a member bound to another identity", slog.String("email", email))
			return 0, err
		}
	} else if err := tx.Raw(`INSERT INTO users (email, name) VALUES (?, ?) RETURNING id`,
		email, strings.TrimSpace(member.Name)).Scan(&userID).Error; err != nil {
		return 0, err
	}
	return userID, tx.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)
		ON CONFLICT (provider, subject) DO NOTHING`, userID, r.issuer, member.Subject).Error
}

func firstID(ids []int64) int64 {
	if len(ids) == 0 {
		return 0
	}
	return ids[0]
}

func first(ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	return ids[0]
}
