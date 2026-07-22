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
	servers         []AppServer
	shutdownHooks   []func(ctx context.Context) error
	shutdownOnce    sync.Once
	shutdownTimeout time.Duration
	logger          *mlog.Logger
	ctx             context.Context
	cancel          context.CancelFunc
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
		if logger != nil {
			a.logger = logger
		}
	}
}

// WithShutdownTimeout sets the maximum duration for each server stop and for shutdown hooks.
func WithShutdownTimeout(timeout time.Duration) Option {
	return func(a *App) {
		if timeout > 0 {
			a.shutdownTimeout = timeout
		}
	}
}

// AppServer defines the interface for a server that can be managed by the App.
type AppServer interface {
	// Start starts the server and blocks.
	// It is expected to return an error if the server fails to start.
	// A call to Stop should cause Start to unblock and return nil.
	Start(ctx context.Context) error
	// Stop gracefully shuts down the server.
	// It should be idempotent and is expected to be called by the app framework.
	Stop(ctx context.Context) error
}

// NewApp creates a new App instance with the given options.
func NewApp(opts ...Option) *App {
	ctx, cancel := context.WithCancel(context.Background())
	app := &App{
		servers:         make([]AppServer, 0),
		shutdownHooks:   make([]func(ctx context.Context) error, 0),
		logger:          mlog.New(),
		shutdownTimeout: 10 * time.Second,
		ctx:             ctx,
		cancel:          cancel,
	}
	for _, opt := range opts {
		opt(app)
	}
	return app
}

// Run starts the application and waits for a signal to gracefully shutdown.
func (a *App) Run() error {
	defer a.cancel()
	eg, ctx := errgroup.WithContext(a.ctx)

	// Start all servers and their corresponding stop listeners.
	for _, s := range a.servers {
		srv := s
		if srv == nil {
			continue
		}
		// Start the server in a goroutine.
		eg.Go(func() error {
			err := srv.Start(ctx)
			// A server returning, with or without an error, ends the application lifecycle.
			a.cancel()
			return err
		})
		// Start a corresponding stop listener for the server.
		eg.Go(func() error {
			// Wait for the context to be canceled.
			<-ctx.Done()
			return stopServer(srv, a.shutdownTimeout)
		})
	}

	// Start a goroutine to listen for OS signals.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)
	eg.Go(func() error {
		select {
		case <-ctx.Done():
			// This can happen if another part of the group fails first.
			return nil
		case sig := <-quit:
			a.logger.Infof(context.Background(), "Received signal %v, initiating shutdown.", sig)
			// Trigger the graceful shutdown by canceling the main context.
			// This will cause <-ctx.Done() to unblock in the server stop listeners.
			a.cancel()
			return nil
		}
	})

	startErr := eg.Wait()

	// Log the startup error as soon as it's available. This is crucial for making
	// the root cause of a shutdown clear, as the "stopping server..." logs from
	// individual services may appear first due to the concurrent nature of shutdown.
	if startErr != nil && !errors.Is(startErr, context.Canceled) {
		a.logger.Warnf(context.Background(), "Shutdown initiated due to a service startup failure. You can see the error from return value.")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.shutdownTimeout)
	defer shutdownCancel()

	var shutdownErr error
	a.shutdownOnce.Do(func() {
		// Execute shutdown hooks in reverse order.
		for i := len(a.shutdownHooks) - 1; i >= 0; i-- {
			if err := a.shutdownHooks[i](shutdownCtx); err != nil {
				a.logger.Errorf(shutdownCtx, err, "Shutdown hook failed")
				shutdownErr = errors.Join(shutdownErr, err)
			}
		}
	})

	// Determine the final error to return.
	// A `context.Canceled` error from `startErr` is expected on a clean shutdown,
	// so we don't treat it as a true error.
	if startErr != nil && !errors.Is(startErr, context.Canceled) {
		return startErr
	}

	return shutdownErr
}

func stopServer(server AppServer, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- server.Stop(ctx)
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
