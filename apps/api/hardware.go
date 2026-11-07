package main

import (
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	hardwaremodule "github.com/FacileStudio/Bulle/apps/api/modules/hardware"
	"github.com/go-chi/chi/v5"
)

func mountHardware(r chi.Router, deps routerDeps, requireAuth func(http.Handler) http.Handler) {
	if deps.DB == nil {
		return
	}
	registry := deps.Hardware
	if registry == nil {
		registry = hardware.NewRegistry()
	}
	devices := hardware.NewPostgres(deps.DB)
	hardwaremodule.RegisterRoutes(r, hardwaremodule.Config{
		Devices:    devices,
		Sockets:    devices,
		Registry:   registry,
		Origins:    deps.Env.CORSAllowedOrigins,
		GatewayURL: deps.Env.HardwareGatewayURL,
		Logger:     deps.Logger,
	}, requireAuth)
}
