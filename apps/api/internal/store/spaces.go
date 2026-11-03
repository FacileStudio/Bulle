package store

import (
	"context"
	"time"
)

// DefaultSpaceName names the space every user gets before they create one.
const DefaultSpaceName = "Personal"

// Space is one work context of a user: its own chats, jobs and the suite
// accounts the agent uses when the model does not name one.
type Space struct {
	ID              string
	UserID          int64
	Name            string
	Icon            string
	MyceliumSpaceID string
	OpusWorkspace   string
	AgendaCalendar  int64
	CourrierAccount int64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Spaces stores a user's spaces. The oldest space is the user's default, and
// deleting a space deletes its sessions and jobs.
type Spaces interface {
	SpaceReader
	SpaceWriter
}

// SpaceReader reads spaces, creating the default one when a user has none.
type SpaceReader interface {
	ListSpaces(ctx context.Context, userID int64) ([]Space, error)
	GetSpace(ctx context.Context, userID int64, id string) (Space, error)
	EnsureDefaultSpace(ctx context.Context, userID int64) (Space, error)
}

// SpaceWriter changes spaces. DeleteSpace answers ErrLastSpace rather than
// leave a user without one.
type SpaceWriter interface {
	CreateSpace(ctx context.Context, space Space) (Space, error)
	UpdateSpace(ctx context.Context, space Space) (Space, error)
	DeleteSpace(ctx context.Context, userID int64, id string) error
}
