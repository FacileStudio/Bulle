package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/FacileStudio/porte/directory"
	"log/slog"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/credentials"
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	pushsender "github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	assistantmodule "github.com/FacileStudio/Bulle/apps/api/modules/assistant"
	attachmentsmodule "github.com/FacileStudio/Bulle/apps/api/modules/attachments"
	"github.com/FacileStudio/Bulle/apps/api/modules/auth"
	computersmodule "github.com/FacileStudio/Bulle/apps/api/modules/computers"
	credentialsmodule "github.com/FacileStudio/Bulle/apps/api/modules/credentials"
	jobsmodule "github.com/FacileStudio/Bulle/apps/api/modules/jobs"
	llmmodule "github.com/FacileStudio/Bulle/apps/api/modules/llm"
	matrixmodule "github.com/FacileStudio/Bulle/apps/api/modules/matrix"
	pushmodule "github.com/FacileStudio/Bulle/apps/api/modules/push"
	sessionsmodule "github.com/FacileStudio/Bulle/apps/api/modules/sessions"
	spacesmodule "github.com/FacileStudio/Bulle/apps/api/modules/spaces"
	suitemodule "github.com/FacileStudio/Bulle/apps/api/modules/suite"
	usagemodule "github.com/FacileStudio/Bulle/apps/api/modules/usage"

	"github.com/FacileStudio/porte"
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
			spacesmodule.Documentation,
			assistantmodule.Documentation,
			jobsmodule.Documentation,
			credentialsmodule.Documentation,
			llmmodule.Documentation,
			matrixmodule.Documentation,
			pushmodule.Documentation,
			suitemodule.Documentation,
			usagemodule.Documentation,
			attachmentsmodule.Documentation,
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

	Computers   computersmodule.Manager
	Store       store.Store
	Runner      *runner.Runner
	Suite       *suite.Gateway
	Push        *pushsender.Hub
	Credentials credentials.Store
	Matrix      matrixmodule.Bots
	Directory   *directory.Syncer
}

func buildRouter(deps routerDeps) (chi.Router, error) {
	db, sqlDB, appEnv, appLogger := deps.DB, deps.SQL, deps.Env, deps.Logger
	identityStore := portepg.New(sqlDB)

	var sessionStore porte.SessionStore = identityStore.Sessions()
	if deps.Suite != nil {
		sessionStore = suite.ForgetOnRevoke(sessionStore, deps.Suite.Tokens())
	}
	sessions, err := session.New(appEnv.Porte, session.Deps{
		Sessions: sessionStore,
		Logger:   appLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to configure sessions: %w", err)
	}

	kit, err := oidc.New(context.Background(), appEnv.Porte, oidc.Deps{
		Users:      auth.NewUserStore(db, appEnv.Suite.Enabled()),
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

	if deps.Directory != nil {
		router.Method(http.MethodPost, directory.SyncPath, deps.Directory.Handler())
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
			usagemodule.RegisterRoutes(r, deps.Store, appEnv.LLM.Model, requireAuth)
			spacesmodule.RegisterRoutes(r, deps.Store, requireAuth)
		}
		pushmodule.RegisterRoutes(r, deps.Push, requireAuth)
		if deps.Matrix != nil {
			matrixmodule.RegisterRoutes(r, deps.Matrix, requireAuth)
		}
		if deps.Credentials != nil {
			credentialsmodule.RegisterRoutes(r, deps.Credentials, requireAuth)
			llmmodule.RegisterRoutes(r, deps.Credentials, credentials.ErrNoKey, llmOffer(appEnv.LLM, deps.Store), requireAuth)
		}
		if deps.Suite != nil && len(deps.Suite.Config().Enabled()) > 0 {
			suitemodule.RegisterRoutes(r, deps.Suite, requireAuth)
		}
		if deps.Store != nil && deps.Runner != nil {
			sessionsmodule.RegisterRoutes(r, deps.Store, deps.Runner, appLogger, requireAuth)
			attachmentsmodule.RegisterRoutes(r, deps.Store, requireAuth)
			jobsmodule.RegisterRoutes(r, deps.Store, deps.Runner, appLogger, requireAuth)
		}
	})

	clientDir := spa.DirFromEnv()
	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router, nil
}
