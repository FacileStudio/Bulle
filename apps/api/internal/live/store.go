package live

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

// Store publishes session.created and session.deleted for every session the
// wrapped store creates or deletes, whichever handler or job asked for it.
type Store struct {
	store.Store
	pub Publisher
}

// Wrap returns st publishing session changes to pub. A nil pub returns st.
func Wrap(st store.Store, pub Publisher) store.Store {
	if pub == nil {
		return st
	}
	return &Store{Store: st, pub: pub}
}

// CreateSession creates the session and announces it to its owner, except an
// ideas batch session, which never shows in the session list.
func (s *Store) CreateSession(ctx context.Context, session store.Session) (store.Session, error) {
	created, err := s.Store.CreateSession(ctx, session)
	if err == nil && created.Channel != store.ChannelIdeas {
		s.pub.Publish(created.UserID, SessionCreated, SessionOf(created, StatusIdle))
	}
	return created, err
}

// DeleteSession deletes the user's session and announces it is gone.
func (s *Store) DeleteSession(ctx context.Context, userID int64, id string) error {
	err := s.Store.DeleteSession(ctx, userID, id)
	if err == nil {
		s.pub.Publish(userID, SessionDeleted, Deleted{SessionID: id})
	}
	return err
}
