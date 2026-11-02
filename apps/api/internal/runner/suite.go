package runner

import (
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/nacelle"
)

func (r *Runner) suiteTools(userID int64) ([]nacelle.Tool, string) {
	if r.cfg.Suite == nil {
		return nil, ""
	}
	tools, note, err := suite.Tools(r.cfg.Suite, userID)
	if err != nil {
		r.cfg.Logger.Error("building suite tools", "error", err)
		return nil, ""
	}
	return tools, note
}
