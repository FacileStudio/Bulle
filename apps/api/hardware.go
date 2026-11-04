package main

import (
	"log/slog"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	hardwaremodule "github.com/FacileStudio/Bulle/apps/api/modules/hardware"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func mountHardware(r chi.Router, db *gorm.DB, appEnv env.Config, logger *slog.Logger, requireAuth func(http.Handler) http.Handler) {
	if db == nil {
		return
	}
	devices := hardware.NewPostgres(db)
	hardwaremodule.RegisterRoutes(r, hardwaremodule.Config{
		Devices:  devices,
		Sockets:  devices,
		Registry: hardware.NewRegistry(),
		Origins:  appEnv.CORSAllowedOrigins,
		Logger:   logger,
	}, requireAuth)
}
