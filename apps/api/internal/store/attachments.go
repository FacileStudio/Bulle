package store

import (
	"context"
	"errors"
	"time"
)

// AttachmentQuota is how many bytes of attachments one user may keep.
const AttachmentQuota int64 = 1 << 30

// ErrAttachmentQuota reports that storing a file would take the user past
// AttachmentQuota.
var ErrAttachmentQuota = errors.New("store: attachment quota reached")

// Attachment is a file a user uploaded to send with a message. Data is empty
// when it was read through AttachmentInfo.
type Attachment struct {
	ID        string
	UserID    int64
	Name      string
	MediaType string
	Size      int64
	Data      []byte
	CreatedAt time.Time
}

// Attachments keeps uploaded files and the personal Mycelium space each user's
// memory lives in.
type Attachments interface {
	CreateAttachment(ctx context.Context, attachment Attachment) (Attachment, error)
	GetAttachment(ctx context.Context, userID int64, id string) (Attachment, error)
	AttachmentInfo(ctx context.Context, userID int64, ids []string) ([]Attachment, error)
}

// MyceliumSpaces remembers the Mycelium space created for a user's memory.
type MyceliumSpaces interface {
	MyceliumSpace(ctx context.Context, userID int64) (string, error)
	SaveMyceliumSpace(ctx context.Context, userID int64, spaceID string) error
}

type attachmentRow struct {
	ID        string
	UserID    int64
	Name      string
	MediaType string
	Size      int64
	Data      []byte
	CreatedAt time.Time
}

func (r attachmentRow) attachment() Attachment { return Attachment(r) }

// CreateAttachment stores the bytes and answers the row without them.
func (p *Postgres) CreateAttachment(ctx context.Context, a Attachment) (Attachment, error) {
	var rows []attachmentRow
	err := p.db.WithContext(ctx).Raw(
		`INSERT INTO attachments (id, user_id, name, media_type, size, data)
		SELECT ?, ?, ?, ?, ?, ?
		WHERE (SELECT COALESCE(SUM(size), 0) FROM attachments WHERE user_id = ?) + ? <= ?
		RETURNING id, user_id, name, media_type, size, created_at`,
		a.ID, a.UserID, a.Name, a.MediaType, a.Size, a.Data, a.UserID, a.Size, AttachmentQuota,
	).Scan(&rows).Error
	if err != nil {
		return Attachment{}, err
	}
	if len(rows) == 0 {
		return Attachment{}, ErrAttachmentQuota
	}
	return rows[0].attachment(), nil
}

// GetAttachment answers one of the user's attachments with its bytes.
func (p *Postgres) GetAttachment(ctx context.Context, userID int64, id string) (Attachment, error) {
	var rows []attachmentRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT id, user_id, name, media_type, size, data, created_at FROM attachments WHERE id = ? AND user_id = ?",
		id, userID,
	).Scan(&rows).Error
	if err != nil {
		return Attachment{}, err
	}
	if len(rows) == 0 {
		return Attachment{}, ErrNotFound
	}
	return rows[0].attachment(), nil
}

// AttachmentInfo answers the metadata of the ids the user owns, in the order
// asked. Ids the user does not own are left out.
func (p *Postgres) AttachmentInfo(ctx context.Context, userID int64, ids []string) ([]Attachment, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var rows []attachmentRow
	err := p.db.WithContext(ctx).Raw(
		"SELECT id, user_id, name, media_type, size, created_at FROM attachments WHERE user_id = ? AND id IN ?",
		userID, ids,
	).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	byID := make(map[string]Attachment, len(rows))
	for _, row := range rows {
		byID[row.ID] = row.attachment()
	}
	return inOrder(ids, byID), nil
}

func inOrder(ids []string, byID map[string]Attachment) []Attachment {
	out := make([]Attachment, 0, len(ids))
	for _, id := range ids {
		if found, ok := byID[id]; ok {
			out = append(out, found)
		}
	}
	return out
}

// MyceliumSpace answers the stored space id, or ErrNotFound.
func (p *Postgres) MyceliumSpace(ctx context.Context, userID int64) (string, error) {
	var ids []string
	err := p.db.WithContext(ctx).Raw("SELECT space_id FROM mycelium_spaces WHERE user_id = ?", userID).Scan(&ids).Error
	if err != nil {
		return "", err
	}
	if len(ids) == 0 {
		return "", ErrNotFound
	}
	return ids[0], nil
}

// SaveMyceliumSpace records the user's space, replacing any earlier one.
func (p *Postgres) SaveMyceliumSpace(ctx context.Context, userID int64, spaceID string) error {
	if spaceID == "" {
		return errors.New("store: empty mycelium space id")
	}
	return p.db.WithContext(ctx).Exec(
		`INSERT INTO mycelium_spaces (user_id, space_id) VALUES (?, ?)
		ON CONFLICT (user_id) DO UPDATE SET space_id = EXCLUDED.space_id`,
		userID, spaceID,
	).Error
}
