package runner

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
)

func (r *Runner) suiteRuntime(userID int64) suite.Runtime {
	if r.cfg.Store == nil {
		return suite.Runtime{}
	}
	mycelium := suite.NewMycelium(r.cfg.Suite)
	return suite.Runtime{
		PersonalSpace: func(ctx context.Context) (string, error) {
			return mycelium.PersonalSpace(ctx, userID, r.cfg.Store)
		},
		Spaces: func(ctx context.Context) ([]store.Space, error) {
			return r.cfg.Store.ListSpaces(ctx, userID)
		},
	}
}
