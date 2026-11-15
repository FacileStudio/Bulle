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
		return r.CheckAccess(ctx, userID)
	}
	session, err := r.cfg.Store.GetSession(ctx, userID, sessionID)
	if err != nil {
		return r.CheckAccess(ctx, userID)
	}
	return r.CheckSpaceAccess(ctx, userID, session.SpaceID)
}

func (r *Runner) checkAccessIn(ctx context.Context, userID int64, spaceID string) error {
	if r.cfg.Access != nil {
		if err := r.cfg.Access.CheckAIIn(ctx, userID, spaceID); err != nil {
			return err
		}
	}
	return r.CheckBudget(ctx, userID)
}

// Budget answers the refusal for a user whose monthly AI budget on the
// server's keys is used up.
type Budget interface {
	Check(ctx context.Context, userID int64) error
}

// CheckBudget answers the Budget refusal for the user, nil when they may spend
// more. A user whose default backend runs on their own key is never refused:
// the server pays nothing for them.
func (r *Runner) CheckBudget(ctx context.Context, userID int64) error {
	if r.cfg.Budget == nil {
		return nil
	}
	if r.cfg.Credentials != nil {
		if _, own, err := r.cfg.Credentials.Resolve(ctx, userID, r.cfg.Backend); err == nil && own {
			return nil
		}
	}
	return r.cfg.Budget.Check(ctx, userID)
}
