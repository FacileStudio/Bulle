package main

import (
	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/shares"
	artifactsmodule "github.com/FacileStudio/Bulle/apps/api/modules/artifacts"
	"gorm.io/gorm"
)

// artifactSharing wires the public links of artifacts: their table, and the
// origin of OIDC_SUCCESS_URL for the public URL. Without a database there are none.
func artifactSharing(db *gorm.DB, appEnv env.Config) artifactsmodule.Sharing {
	if db == nil {
		return artifactsmodule.Sharing{}
	}
	return artifactsmodule.Sharing{Store: shares.NewStore(db), Origin: artifactsmodule.PublicOrigin(appEnv.Porte.SuccessURL)}
}
