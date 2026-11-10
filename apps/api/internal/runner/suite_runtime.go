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
		Attachment: func(ctx context.Context, id string) (suite.File, error) {
			found, err := r.cfg.Store.GetAttachment(ctx, userID, id)
			if err != nil {
				return suite.File{}, err
			}
			return suite.File{Name: found.Name, MediaType: found.MediaType, Data: found.Data}, nil
		},
		Spaces: func(ctx context.Context) ([]store.Space, error) {
			return r.cfg.Store.ListSpaces(ctx, userID)
		},
	}
}
