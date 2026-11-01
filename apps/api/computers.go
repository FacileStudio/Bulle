package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Bulle/apps/api/internal/computers"
	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/worker"

	"gorm.io/gorm"
)

func startComputers(db *gorm.DB, appEnv env.Config, logger *slog.Logger) (*computers.Service, func(), error) {
	config, err := computers.NewConfig(appEnv.Boite, logger)
	if err != nil {
		return nil, nil, err
	}
	service := computers.NewService(
		computers.NewStore(db),
		worker.New(appEnv.Boite.WorkerSock, appEnv.Boite.WorkerToken),
		config,
		logger,
	)
	ctx, cancel := context.WithCancel(context.Background())
	go service.RunSweeper(ctx)
	return service, cancel, nil
}
