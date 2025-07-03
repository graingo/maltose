package m

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/graingo/maltose/os/mlog"
	"golang.org/x/sync/errgroup"
)

// App is the core application structure that manages the lifecycle of services and hooks.
type App struct {
	servers       []AppServer
	shutdownHooks []func(ctx context.Context) error
	shutdownOnce  sync.Once
	logger        *mlog.Logger
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

// WithLogger sets the logger for the application.
func WithLogger(logger *mlog.Logger) Option {
	return func(a *App) {
		a.logger = logger
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
		logger:        mlog.New(),
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

	// This function triggers the shutdown process.
	// Using sync.Once ensures that shutdown is only initiated once,
	// even if multiple shutdown signals are received.
	triggerShutdown := sync.OnceFunc(func() {
		// We use a background context for the shutdown signal log, as it's a global event.
		a.logger.Infof(context.Background(), "Shutdown process initiated...")
		// Close the quit channel to signal the main goroutine to proceed with shutdown.
		close(quit)
	})

	// This goroutine waits for a startup failure and triggers shutdown if one occurs.
	go func() {
		if err := eg.Wait(); err != nil {
			triggerShutdown()
		}
	}()

	select {
	case sig := <-quit:
		a.logger.Infof(context.Background(), "Received signal %v, initiating shutdown", sig)
		triggerShutdown()
	case <-ctx.Done():
		// This case is hit if errgroup context is cancelled, usually due to a startup error.
		// We trigger the shutdown process to ensure cleanup happens.
		triggerShutdown()
	}

	// Wait until the shutdown is triggered. This handles the case where eg.Wait() in the
	// goroutine finishes, but the main select hasn't processed the closed quit channel yet.
	<-quit

	// Graceful shutdown with a timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Perform shutdown hooks and stop servers.
	var shutdownErr error
	a.shutdownOnce.Do(func() {
		// Execute shutdown hooks in reverse order
		for i := len(a.shutdownHooks) - 1; i >= 0; i-- {
			if err := a.shutdownHooks[i](shutdownCtx); err != nil {
				a.logger.Errorf(shutdownCtx, err, "Shutdown hook failed")
				shutdownErr = errors.Join(shutdownErr, err)
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
			a.logger.Errorf(shutdownCtx, err, "Application stop error")
			shutdownErr = errors.Join(shutdownErr, err)
		}
	})

	// Finally, wait for the main errgroup and prioritize its error.
	startErr := eg.Wait()
	if startErr != nil && startErr != context.Canceled {
		return startErr // Return the specific startup error.
	}

	// If no startup error, return any error from the shutdown process.
	return shutdownErr
}
