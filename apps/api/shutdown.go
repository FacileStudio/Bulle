package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	runDrainBudget  = 10 * time.Second
	httpDrainBudget = 10 * time.Second
)

type drainer interface {
	Shutdown(ctx context.Context)
}

// drain stops the runner and the HTTP server at the same time, each with its
// own budget, so a slow run never eats the time open requests had to finish.
// The live streams close as soon as shutdown starts. Running out of budget is
// a clean exit: whatever was still open was cut on purpose.
func drain(runs drainer, server *http.Server, closeStreams func(), logger *slog.Logger) error {
	server.RegisterOnShutdown(closeStreams)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), runDrainBudget)
		defer cancel()
		runs.Shutdown(ctx)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), httpDrainBudget)
	defer cancel()
	err := server.Shutdown(ctx)
	wg.Wait()
	if errors.Is(err, context.DeadlineExceeded) {
		logger.Warn("open connections cut at the shutdown deadline")
		return nil
	}
	return err
}
