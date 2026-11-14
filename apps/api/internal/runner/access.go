package runner

import "context"

// Access decides whether a user may start AI work, in a Registre shared space
// when spaceID is not empty. A nil Access allows everyone. Its refusal is
// returned to the caller unchanged, so it can carry its own HTTP status.
type Access interface {
	CheckAIIn(ctx context.Context, userID int64, spaceID string) error
}

// CheckAccess answers the Access refusal for the user, nil when they may start
// AI work.
func (r *Runner) CheckAccess(ctx context.Context, userID int64) error {
	return r.checkAccessIn(ctx, userID, "")
}

// CheckSpaceAccess is CheckAccess for work in the Bulle space spaceID: when
// that space is linked to a Registre shared space, the shared space's plan
// counts too. A space that cannot be read falls back to the user's own plan.
func (r *Runner) CheckSpaceAccess(ctx context.Context, userID int64, spaceID string) error {
	if r.cfg.Access == nil || spaceID == "" {
		return r.CheckAccess(ctx, userID)
	}
	space, err := r.cfg.Store.GetSpace(ctx, userID, spaceID)
	if err != nil {
		return r.CheckAccess(ctx, userID)
	}
	return r.checkAccessIn(ctx, userID, space.RegistreSpaceID)
}

// sessionAccess is CheckSpaceAccess for a run in sessionID's space. A session
// that cannot be read falls back to the user's own plan.
func (r *Runner) sessionAccess(ctx context.Context, userID int64, sessionID string) error {
	if r.cfg.Access == nil {
		return nil
	}
	session, err := r.cfg.Store.GetSession(ctx, userID, sessionID)
	if err != nil {
		return r.CheckAccess(ctx, userID)
	}
	return r.CheckSpaceAccess(ctx, userID, session.SpaceID)
}

func (r *Runner) checkAccessIn(ctx context.Context, userID int64, spaceID string) error {
	if r.cfg.Access == nil {
		return nil
	}
	return r.cfg.Access.CheckAIIn(ctx, userID, spaceID)
}
