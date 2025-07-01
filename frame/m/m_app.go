package m

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

// App is the core application structure that manages the lifecycle of services and hooks.
type App struct {
	servers       []AppServer
	shutdownHooks []func(ctx context.Context) error
	shutdownOnce  sync.Once
}

// An Option configures an App.
type Option func(*App)

// WithServer adds servers to the application.
func WithServer(servers ...AppServer) Option {
	return func(a *App) {
		a.servers = append(a.servers, servers...)
	}
}

// WithShutdownHook adds functions to be called during graceful shutdown.
func WithShutdownHook(hooks ...func(ctx context.Context) error) Option {
	return func(a *App) {
		a.shutdownHooks = append(a.shutdownHooks, hooks...)
	}
}

// AppServer defines the interface for a server that can be managed by the App.
type AppServer interface {
	Start(ctx context.Context) error
	Shutdown(ctx context.Context) error
}

// NewApp creates a new App instance with the given options.
func NewApp(opts ...Option) *App {
	app := &App{
		servers:       make([]AppServer, 0),
		shutdownHooks: make([]func(ctx context.Context) error, 0),
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// Run starts the application and waits for a signal to gracefully shutdown.
func (a *App) Run() error {
	eg, ctx := errgroup.WithContext(context.Background())

	// Start servers
	for _, s := range a.servers {
		srv := s
		eg.Go(func() error {
			return srv.Start(ctx)
		})
	}

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var anError error
	select {
	case <-ctx.Done():
		anError = ctx.Err()
		Log().Infof(ctx, "Application context cancelled, initiating shutdown: %v", anError)
	case sig := <-quit:
		Log().Infof(ctx, "Received signal %v, initiating shutdown", sig)
	}

	// Graceful shutdown with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	a.shutdownOnce.Do(func() {
		// Execute shutdown hooks in reverse order
		for i := len(a.shutdownHooks) - 1; i >= 0; i-- {
			if err := a.shutdownHooks[i](shutdownCtx); err != nil {
				Log().Errorf(shutdownCtx, err, "Shutdown hook failed")
				if anError == nil {
					anError = err
				}
			}
		}

		// Stop the main applications in a separate error group
		stopGroup, _ := errgroup.WithContext(shutdownCtx)
		for _, srv := range a.servers {
			s := srv
			stopGroup.Go(func() error {
				return s.Shutdown(shutdownCtx)
			})
		}
		if err := stopGroup.Wait(); err != nil {
			Log().Errorf(shutdownCtx, err, "Application stop error")
			if anError == nil {
				anError = err
			}
		}
	})

	// Wait for the main services to finish and return any error
	if err := eg.Wait(); err != nil && anError == nil {
		return err
	}
	return anError
}
