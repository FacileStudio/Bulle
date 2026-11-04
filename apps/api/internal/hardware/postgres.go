package hardware

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const columns = `id, user_id, name, device_class, hardware_id, capabilities::text AS capabilities,
	last_seen_at, COALESCE(last_transport, '') AS last_transport, created_at`

// Postgres stores devices in the hardware_devices table.
type Postgres struct{ db *gorm.DB }

// NewPostgres answers a device store backed by Postgres.
func NewPostgres(db *gorm.DB) *Postgres { return &Postgres{db: db} }

// Claim records a device for its owner under tokenHash. Claiming a hardware id
// the owner already paired renames it and replaces its token, so a device
// reset and paired again keeps its id while the old token stops working.
func (p *Postgres) Claim(ctx context.Context, d Device, tokenHash string) (Device, error) {
	var rows []Device
	err := p.db.WithContext(ctx).Raw(
		`INSERT INTO hardware_devices (id, user_id, name, device_class, hardware_id, token_hash)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id, hardware_id) DO UPDATE SET name = EXCLUDED.name,
		device_class = EXCLUDED.device_class, token_hash = EXCLUDED.token_hash
		RETURNING `+columns,
		d.ID, d.UserID, d.Name, d.DeviceClass, d.HardwareID, tokenHash,
	).Scan(&rows).Error
	return first(rows, err)
}

// Devices lists the user's devices, oldest first.
func (p *Postgres) Devices(ctx context.Context, userID int64) ([]Device, error) {
	var rows []Device
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+columns+" FROM hardware_devices WHERE user_id = ? ORDER BY created_at, id", userID,
	).Scan(&rows).Error
	return rows, err
}

// Owned answers one of the user's devices, or ErrNotFound when the id is
// unknown or paired to another account.
func (p *Postgres) Owned(ctx context.Context, userID int64, id string) (Device, error) {
	var rows []Device
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+columns+" FROM hardware_devices WHERE id = ? AND user_id = ?", id, userID,
	).Scan(&rows).Error
	return first(rows, err)
}

// ByTokenHash answers the device a token belongs to, or ErrNotFound.
func (p *Postgres) ByTokenHash(ctx context.Context, hash string) (Device, error) {
	var rows []Device
	err := p.db.WithContext(ctx).Raw(
		"SELECT "+columns+" FROM hardware_devices WHERE token_hash = ?", hash,
	).Scan(&rows).Error
	return first(rows, err)
}

// Revoke deletes one of the user's devices, or answers ErrNotFound.
func (p *Postgres) Revoke(ctx context.Context, userID int64, id string) error {
	result := p.db.WithContext(ctx).Exec("DELETE FROM hardware_devices WHERE id = ? AND user_id = ?", id, userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Touch records when and over which transport the device last connected.
func (p *Postgres) Touch(ctx context.Context, id, transport string, at time.Time) error {
	return p.db.WithContext(ctx).Exec(
		"UPDATE hardware_devices SET last_seen_at = ?, last_transport = ? WHERE id = ?", at, transport, id,
	).Error
}

func first(rows []Device, err error) (Device, error) {
	if err != nil {
		return Device{}, err
	}
	if len(rows) == 0 {
		return Device{}, ErrNotFound
	}
	return rows[0], nil
}
