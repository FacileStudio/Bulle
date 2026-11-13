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

// sessionAccess is CheckAccess for a run in sessionID: when the session's
// space is linked to a Registre shared space, that space's plan counts too.
// A session or space that cannot be read falls back to the user's own plan.
func (r *Runner) sessionAccess(ctx context.Context, userID int64, sessionID string) error {
	if r.cfg.Access == nil {
		return nil
	}
	session, err := r.cfg.Store.GetSession(ctx, userID, sessionID)
	if err != nil || session.SpaceID == "" {
		return r.CheckAccess(ctx, userID)
	}
	space, err := r.cfg.Store.GetSpace(ctx, userID, session.SpaceID)
	if err != nil {
		return r.CheckAccess(ctx, userID)
	}
	return r.checkAccessIn(ctx, userID, space.RegistreSpaceID)
}

func (r *Runner) checkAccessIn(ctx context.Context, userID int64, spaceID string) error {
	if r.cfg.Access == nil {
		return nil
	}
	return r.cfg.Access.CheckAIIn(ctx, userID, spaceID)
}
