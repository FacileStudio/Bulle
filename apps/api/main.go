package main

import (
	"context"
	_ "embed"
	"errors"
	spacesmodule "github.com/FacileStudio/Bulle/apps/api/modules/spaces"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"github.com/FacileStudio/Bulle/apps/api/internal/computers"
	"github.com/FacileStudio/Bulle/apps/api/internal/credentials"
	"github.com/FacileStudio/Bulle/apps/api/internal/database"
	"github.com/FacileStudio/Bulle/apps/api/internal/disclosure"
	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/erasure"
	"github.com/FacileStudio/Bulle/apps/api/internal/hardware"
	"github.com/FacileStudio/Bulle/apps/api/internal/live"
	"github.com/FacileStudio/Bulle/apps/api/internal/matrixbot"
	"github.com/FacileStudio/Bulle/apps/api/internal/plans"
	"github.com/FacileStudio/Bulle/apps/api/internal/prisme"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/scheduler"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/Bulle/apps/api/modules/auth"
	"github.com/FacileStudio/Bulle/core/jev"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/logger"
	"github.com/FacileStudio/tronc/migrate"
)

//go:embed prompts/openui.txt
var openuiPrompt string

func main() { os.Exit(run()) }

func run() int {
	if healthcheck.Handle(os.Args) {
		return 0
	}

	appEnv, err := env.Load()
	appLogger := logger.New(logger.Config{})
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return 1
	}
	var journalClient *journal.Client
	appLogger = logger.New(logger.Config{
		Level: appEnv.LogLevel,
		Wrap: func(handler slog.Handler) slog.Handler {
			if appEnv.JournalURL == "" || appEnv.JournalToken == "" {
				return handler
			}
			journalClient = journal.New(journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken})
			return journal.NewHandler(journalClient, handler)
		},
	})
	if journalClient != nil {
		defer journalClient.Close()
	}

	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to open database", slog.Any("error", err))
		return 1
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to access database handle", slog.Any("error", err))
		return 1
	}

	migrations, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		appLogger.Error("failed to locate embedded migrations", slog.Any("error", err))
		return 1
	}
	migrateConfig := migrate.Config{DB: sqlDB, FS: migrations, Logger: appLogger}
	if handled, err := migrate.Command(context.Background(), os.Args, migrateConfig); handled {
		if err != nil {
			appLogger.Error("migrate", slog.Any("error", err))
			return 1
		}
		return 0
	}
	if err := migrate.Run(context.Background(), migrateConfig); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return 1
	}
	if filled, err := auth.BackfillPictures(context.Background(), db, appEnv.Porte.Issuer); err != nil {
		appLogger.Warn("failed to backfill profile pictures", slog.Any("error", err))
	} else if filled > 0 {
		appLogger.Info("backfilled profile pictures", slog.Int64("users", filled))
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("failed to close database", slog.Any("error", err))
		}
	}()

	computerService, stopComputers, err := startComputers(db, appEnv, appLogger)
	if err != nil {
		appLogger.Error("failed to configure computers", slog.Any("error", err))
		return 1
	}
	defer stopComputers()

	pushSender := newPushHub(db, appEnv.Push, appLogger)
	st := live.Wrap(store.NewPostgres(db), pushSender)
	credentialsKey, err := credentials.ParseKey(appEnv.CredentialsKey)
	if err != nil {
		appLogger.Error("failed to read the credentials key", slog.Any("error", err))
		return 1
	}
	credentialStore, err := credentials.NewPostgres(db, credentialsKey)
	if err != nil {
		appLogger.Error("failed to configure credentials", slog.Any("error", err))
		return 1
	}
	if credentialsKey == nil {
		appLogger.Warn("BULLE_CREDENTIALS_KEY is not set: per-user provider keys are disabled")
	}
	browser, err := loadBrowserBundle(appEnv.Boite.BrowserBundle, appLogger)
	if err != nil {
		appLogger.Error("failed to configure the browser", slog.Any("error", err))
		return 1
	}
	computerProvider := computers.NewProvider(computerService)
	porteStore, err := newPorteStore(sqlDB)
	if err != nil {
		appLogger.Error("failed to open the porte store", slog.Any("error", err))
		return 1
	}
	resealInBackground(context.Background(), porteStore, appLogger)
	suiteTokens := suite.NewTokens(appEnv.Suite, porteStore.Identities(), nil)
	suiteGateway := suite.NewGateway(suiteTokens, appEnv.Suite, nil)
	gate := plans.New(db, appEnv.Porte.Issuer, appLogger)
	if err := gate.Load(context.Background()); err != nil {
		appLogger.Error("failed to load the saved plan entitlements", slog.Any("error", err))
		return 1
	}
	hardwareRegistry := hardware.NewRegistry()
	runnerService := runner.New(runner.Config{
		Access:          gate,
		Devices:         hardwareRegistry.AgentTools,
		HostTools:       runner.WebTools(runner.WebOptions{Credentials: credentialStore, Quota: st, DailyLimit: appEnv.SearchDailyLimit}),
		Push:            pushSender,
		Store:           st,
		Credentials:     credentialStore,
		Computers:       computerProvider,
		Suite:           suiteGateway,
		Outward:         append(suite.WriteTools(), runner.OutwardTools()...),
		VMTools:         browser.tools(),
		TimeTracking:    appEnv.TimeTracking,
		WebPromptSuffix: openuiPrompt,
		Backend:         appEnv.LLM.Backend,
		Model:           appEnv.LLM.Model,
		Models:          appEnv.LLM.Models,
		Preflight:       browser.preflight,
		MaxIterations:   50,
		Logger:          appLogger,
	})
	if err := runnerService.Recover(context.Background()); err != nil {
		appLogger.Warn("failed to recover live runs", slog.Any("error", err))
	}

	sched := scheduler.New(scheduler.Config{
		DB:     db,
		Store:  st,
		Runner: runnerService,
		Ideas:  runnerService,
		Memory: runnerService,
		Push:   pushSender,
		Logger: appLogger,
	})
	sched.Start()
	defer sched.Stop()

	matrixBots, err := matrixbot.New(matrixbot.Config{
		SQL: sqlDB, DatabaseURL: appEnv.DatabaseURL, Store: st, Runner: runnerService, Key: credentialsKey,
		Prisme: prisme.New(appEnv.Prisme.URL, appEnv.Prisme.ServiceKey), Screen: disclosure.New(jev.Config{}), Logger: appLogger,
	})
	if err != nil {
		appLogger.Error("failed to configure matrix bots", slog.Any("error", err))
		return 1
	}
	matrixBots.Start()
	defer matrixBots.Stop()

	deletions := erasure.New(
		db,
		directory.Client{Issuer: appEnv.Porte.Issuer, ClientID: appEnv.Porte.ClientID, ClientSecret: appEnv.Porte.ClientSecret},
		portepg.New(sqlDB).EraseIdentity,
		erasure.Hooks{EraseComputer: computerService.Erase, HaltBot: matrixBots.Halt, ForgetTokens: suiteTokens.Forget},
		appLogger,
	)

	syncer := spacesmodule.NewDirectory(db, appEnv.Porte.Issuer, appEnv.Porte.ClientID, appEnv.Porte.ClientSecret, appLogger, gate.Apply)

	router, err := buildRouter(routerDeps{
		DB:          db,
		SQL:         sqlDB,
		Env:         appEnv,
		Logger:      appLogger,
		Computers:   computerService,
		Store:       st,
		Runner:      runnerService,
		Suite:       suiteGateway,
		Push:        pushSender,
		Credentials: credentialStore,
		Matrix:      matrixBots,
		Directory:   syncer,
		Deletions:   deletions,
		Gate:        gate,
		Hardware:    hardwareRegistry,
	})
	if err != nil {
		appLogger.Error("failed to build the router", slog.Any("error", err))
		return 1
	}

	addr := ":" + appEnv.Port
	// WriteTimeout stays zero on purpose. net/http's HTTP/2 server answers an
	// expired write deadline with RST_STREAM(INTERNAL_ERROR), which a Go client
	// reports as "stream error: stream ID N; INTERNAL_ERROR; received from
	// peer". A model answer — reasoning included — streams for minutes, so a
	// whole-response deadline here cut every long turn mid-stream. Handlers that
	// need a deadline set their own with http.ResponseController: the SSE stream
	// refreshes a 60s one per write, so an idle stream still dies and an active
	// one never does.
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if syncer != nil {
		go syncer.Run(shutdownSignal)
	}
	go deletions.Run(shutdownSignal)

	appLogger.Info("server starting", slog.String("addr", addr))
	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server stopped", slog.Any("error", err))
			return 1
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		sched.Stop()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		runnerService.Shutdown(shutdownContext)
		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return 1
		}
		appLogger.Info("server stopped")
	}
	return 0
}
