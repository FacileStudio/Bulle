package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/http"

	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	assistantmodule "github.com/FacileStudio/Bulle/apps/api/modules/assistant"
	"github.com/FacileStudio/Bulle/apps/api/modules/auth"
	computersmodule "github.com/FacileStudio/Bulle/apps/api/modules/computers"
	sessionsmodule "github.com/FacileStudio/Bulle/apps/api/modules/sessions"

	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/httpx"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

const apiTitle = "Bulle API"

const apiDescription = "Authentication for the Bulle web app."

//go:embed migrations/*.sql
var migrationsFS embed.FS

func registry() documentation.Response {
	return documentation.Response{
		Modules: []documentation.Module{
			auth.Documentation,
			computersmodule.Documentation,
			sessionsmodule.Documentation,
			assistantmodule.Documentation,
		},
	}
}

func refConfig() apiref.Config {
	return apiref.Config{
		Title:       apiTitle,
		Description: apiDescription,
		Servers:     []string{"/api"},
		Registry:    registry(),
	}
}

type routerDeps struct {
	DB     *gorm.DB
	SQL    *sql.DB
	Env    env.Config
	Logger *slog.Logger

	Computers computersmodule.Manager
	Store     store.Store
	Runner    *runner.Runner
}

func buildRouter(deps routerDeps) (chi.Router, error) {
	db, sqlDB, appEnv, appLogger := deps.DB, deps.SQL, deps.Env, deps.Logger
	identityStore := portepg.New(sqlDB)

	sessions, err := session.New(appEnv.Porte, session.Deps{
		Sessions: identityStore.Sessions(),
		Logger:   appLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to configure sessions: %w", err)
	}

	kit, err := oidc.New(context.Background(), appEnv.Porte, oidc.Deps{
		Users:      auth.NewUserStore(db),
		Identities: identityStore.Identities(),
		Codes:      identityStore.LoginCodes(),
		Sessions:   sessions,
		Logger:     appLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to configure authentication: %w", err)
	}
	if kit.Enabled() {
		appLogger.Info("single sign-on enabled",
			slog.String("issuer", appEnv.Porte.Issuer),
			slog.Bool("sso_only", appEnv.Porte.SSOOnly))
	}

	authService := auth.NewService(db)

	router := httpx.NewRouter(httpx.Config{
		Logger: appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: appEnv.CORSAllowedOrigins,
		},
	})

	health.Mount(router, health.DB(sqlDB))
	apiref.Mount(router, refConfig())

	requireAuth := func(next http.Handler) http.Handler {
		return middleware.RequireAuth(sessions, authService)(next)
	}

	router.Route("/api", func(r chi.Router) {
		sessions.Mount(r)
		kit.Mount(r)
		auth.RegisterRoutes(r, authService, deps.Store, requireAuth)
		if deps.Computers != nil {
			computersmodule.RegisterRoutes(r, deps.Computers, requireAuth)
		}
		if deps.Store != nil {
			assistantmodule.RegisterRoutes(r, deps.Store, requireAuth)
		}
		if deps.Store != nil && deps.Runner != nil {
			sessionsmodule.RegisterRoutes(r, deps.Store, deps.Runner, appLogger, requireAuth)
		}
	})

	clientDir := spa.DirFromEnv()
	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router, nil
}
