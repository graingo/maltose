package m

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/graingo/maltose/internal/intlog"
	"golang.org/x/sync/errgroup"
)

// App is a server interface that defines the lifecycle of a server.
type App interface {
	// Start starts the server and blocks until the server stops.
	Start(ctx context.Context) error
	// Stop gracefully stops the server.
	Stop(ctx context.Context) error
}

// Run starts the application and waits for a signal to gracefully shutdown the servers.
func AppRun(apps ...App) error {
	eg, ctx := errgroup.WithContext(context.Background())

	for _, s := range apps {
		srv := s
		eg.Go(func() error {
			return srv.Start(ctx)
		})
	}

	// Wait for exit signal.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
		intlog.Printf(ctx, "context cancelled, shutting down")
	case sig := <-quit:
		intlog.Printf(ctx, "received signal %v, shutting down", sig)
	}

	// Create a new context for shutdown, in case the original context was cancelled.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for _, s := range apps {
		srv := s
		eg.Go(func() error {
			return srv.Stop(shutdownCtx)
		})
	}

	intlog.Printf(ctx, "waiting for servers to shutdown")
	return eg.Wait()
}
