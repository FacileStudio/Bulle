package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"github.com/FacileStudio/porte/directory"
	"log/slog"
	"net/http"

	"github.com/FacileStudio/Bulle/apps/api/internal/catalog"
	"github.com/FacileStudio/Bulle/apps/api/internal/credentials"
	documentation "github.com/FacileStudio/Bulle/apps/api/internal/documentation"
	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/erasure"
	"github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/FacileStudio/Bulle/apps/api/internal/middleware"
	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
	pushsender "github.com/FacileStudio/Bulle/apps/api/internal/push"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	artifactsmodule "github.com/FacileStudio/Bulle/apps/api/modules/artifacts"
	assistantmodule "github.com/FacileStudio/Bulle/apps/api/modules/assistant"
	attachmentsmodule "github.com/FacileStudio/Bulle/apps/api/modules/attachments"
	"github.com/FacileStudio/Bulle/apps/api/modules/auth"
	computersmodule "github.com/FacileStudio/Bulle/apps/api/modules/computers"
	credentialsmodule "github.com/FacileStudio/Bulle/apps/api/modules/credentials"
	hardwaremodule "github.com/FacileStudio/Bulle/apps/api/modules/hardware"
	ideasmodule "github.com/FacileStudio/Bulle/apps/api/modules/ideas"
	jobsmodule "github.com/FacileStudio/Bulle/apps/api/modules/jobs"
	llmmodule "github.com/FacileStudio/Bulle/apps/api/modules/llm"
	matrixmodule "github.com/FacileStudio/Bulle/apps/api/modules/matrix"
	preferencesmodule "github.com/FacileStudio/Bulle/apps/api/modules/preferences"
	pushmodule "github.com/FacileStudio/Bulle/apps/api/modules/push"
	sessionsmodule "github.com/FacileStudio/Bulle/apps/api/modules/sessions"
	spacesmodule "github.com/FacileStudio/Bulle/apps/api/modules/spaces"
	suitemodule "github.com/FacileStudio/Bulle/apps/api/modules/suite"
	usagemodule "github.com/FacileStudio/Bulle/apps/api/modules/usage"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/oidc"
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
			ideasmodule.Documentation,
			artifactsmodule.Documentation,
			credentialsmodule.Documentation,
			preferencesmodule.Documentation,
			llmmodule.Documentation,
			matrixmodule.Documentation,
			pushmodule.Documentation,
			suitemodule.Documentation,
			usagemodule.Documentation,
			attachmentsmodule.Documentation,
			hardwaremodule.Documentation,
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
	Deletions   *erasure.Deletions
	Gate        *plans.Gate
	Hardware    *hardware.Registry
}

func buildRouter(deps routerDeps) (chi.Router, error) {
	db, sqlDB, appEnv, appLogger := deps.DB, deps.SQL, deps.Env, deps.Logger
	identityStore, err := newPorteStore(sqlDB)
	if err != nil {
		return nil, fmt.Errorf("failed to open the porte store: %w", err)
	}

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
		appLogger.Info("single sign-on enabled", slog.String("issuer", appEnv.Porte.Issuer))
	}

	authService := auth.NewService(db)

	clientDir := spa.DirFromEnv()
	router := httpx.NewRouter(httpx.Config{
		Logger: appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: appEnv.CORSAllowedOrigins,
		},
		TrustedProxies: appEnv.TrustedProxies,
		CDNProxies:     appEnv.CDNProxies,
		CDNHeader:      appEnv.CDNHeader,
		Security:       securityConfig(clientDir, appLogger),
	})

	health.Mount(router, health.DB(sqlDB))
	router.Get("/version", versionHandler)
	apiref.Mount(router, refConfig())

	requireAuth := func(next http.Handler) http.Handler {
		return middleware.RequireAuth(sessions, authService)(next)
	}

	if deps.Directory != nil {
		router.Method(http.MethodPost, directory.SyncPath, deps.Directory.Handler())
	}
	if deps.Deletions != nil {
		router.Method(http.MethodPost, directory.DeletionsPath, deps.Deletions.Handler())
	}

	// The one catalogue both surfaces read: the web picker searches it, and
	// GET /llm hands it to the CLI as the models /model may switch to.
	var listModels func(ctx context.Context, userID int64) []catalog.Model
	if deps.Credentials != nil {
		catalogue := catalog.New(deps.Credentials, appEnv.LLM.Models)
		listModels = func(ctx context.Context, userID int64) []catalog.Model {
			return catalogue.List(ctx, userID, appEnv.LLM.Backend)
		}
	}

	router.Route("/api", func(r chi.Router) {
		sessions.Mount(r)
		kit.Mount(r)
		auth.RegisterRoutes(r, authService, deps.Store, deps.Gate.Plan, requireAuth)
		if deps.Computers != nil {
			computersmodule.RegisterRoutes(r, deps.Computers, requireAuth)
		}
		if deps.Store != nil {
			assistantmodule.RegisterRoutes(r, deps.Store, requireAuth)
			preferencesmodule.RegisterRoutes(r, deps.Store, listModels, appEnv.LLM.Model, requireAuth)
			usagemodule.RegisterRoutes(r, deps.Store, appEnv.LLM.Model, requireAuth)
			spacesmodule.RegisterRoutes(r, deps.Store, requireAuth)
			artifactsmodule.RegisterRoutes(r, deps.Store, deps.Suite, requireAuth)
		}
		pushmodule.RegisterRoutes(r, deps.Push, requireAuth)
		mountHardware(r, deps, requireAuth)
		if deps.Matrix != nil {
			matrixmodule.RegisterRoutes(r, deps.Matrix, requireAuth)
		}
		if deps.Credentials != nil {
			credentialsmodule.RegisterRoutes(r, deps.Credentials, requireAuth)
			llmmodule.RegisterRoutes(r, deps.Credentials, credentials.ErrNoKey, llmOffer(appEnv.LLM, deps.Store, listModels), listModels, requireAuth, deps.Gate.Require)
		}
		if deps.Suite != nil && len(deps.Suite.Config().Enabled()) > 0 {
			suitemodule.RegisterRoutes(r, deps.Suite, requireAuth)
		}
		if deps.Store != nil && deps.Runner != nil {
			sessionsmodule.RegisterRoutes(r, deps.Store, deps.Runner, appLogger, requireAuth)
			attachmentsmodule.RegisterRoutes(r, deps.Store, requireAuth)
			jobsmodule.RegisterRoutes(r, deps.Store, deps.Runner, appLogger, requireAuth)
			ideasmodule.RegisterRoutes(r, deps.Store, deps.Runner, appLogger, requireAuth)
		}
	})

	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router, nil
}

func securityConfig(clientDir string, appLogger *slog.Logger) troncmiddleware.SecurityConfig {
	sources := map[string][]string{}
	if spa.Available(clientDir) {
		hashes, err := spa.InlineScriptHashes(clientDir)
		if err != nil {
			appLogger.Warn("failed to hash the client's inline scripts", slog.Any("error", err))
		} else {
			sources["script-src"] = hashes
		}
	}
	return troncmiddleware.SecurityConfig{CSPSources: sources}
}
