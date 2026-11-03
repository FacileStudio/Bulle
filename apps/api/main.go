package main

import (
	"context"
	_ "embed"
	"errors"
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
	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/live"
	"github.com/FacileStudio/Bulle/apps/api/internal/runner"
	"github.com/FacileStudio/Bulle/apps/api/internal/scheduler"
	"github.com/FacileStudio/Bulle/apps/api/internal/store"
	"github.com/FacileStudio/Bulle/apps/api/internal/suite"
	"github.com/FacileStudio/Bulle/core/sandbox"

	"github.com/FacileStudio/Journal/sdk/journal"
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
	computerProvider := computers.NewProvider(computerService)
	suiteTokens := suite.NewTokens(appEnv.Suite, portepg.New(sqlDB).Identities(), nil)
	suiteGateway := suite.NewGateway(suiteTokens, appEnv.Suite, nil)
	runnerService := runner.New(runner.Config{
		Push:            pushSender,
		Store:           st,
		Credentials:     credentialStore,
		Computers:       computerProvider,
		Suite:           suiteGateway,
		Outward:         suite.WriteTools(),
		WebPromptSuffix: openuiPrompt,
		Backend:         appEnv.LLM.Backend,
		Model:           appEnv.LLM.Model,
		Preflight: func(ctx context.Context, target *sandbox.Target) error {
			if target == nil {
				return nil
			}
			_, err := sandbox.PreflightCheck(ctx, target, sandbox.GuardOptions{
				ExpectedUser:    target.User,
				ExpectedWorkdir: target.Workdir,
				Runner:          runner.CleanRunner{},
			})
			return err
		},
		Logger: appLogger,
	})
	if err := runnerService.Recover(context.Background()); err != nil {
		appLogger.Warn("failed to recover live runs", slog.Any("error", err))
	}

	sched := scheduler.New(scheduler.Config{
		DB:     db,
		Store:  st,
		Runner: runnerService,
		Push:   pushSender,
		Logger: appLogger,
	})
	sched.Start()
	defer sched.Stop()

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
	})
	if err != nil {
		appLogger.Error("failed to build the router", slog.Any("error", err))
		return 1
	}

	addr := ":" + appEnv.Port
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
