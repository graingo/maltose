package mhttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// SetStaticPath enhances the static file service.
func (s *Server) SetStaticPath(prefix string, directory string) {
	// implement static file service
	s.engine.StaticFS(prefix, http.Dir(directory))
}

// Run starts the HTTP server.
func (s *Server) Run() {
	ctx := context.Background()

	// register OpenAPI and Swagger
	s.registerDoc(ctx)

	// register all routes before starting
	s.bindRoutes(ctx)

	// print route information
	s.printRoute(ctx)

	srv := &http.Server{
		Addr:           s.config.Address,
		Handler:        s.engine,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}

	// create error channel
	errChan := make(chan error, 1)
	go func() {
		var err error
		if s.config.TLSEnable {
			if s.config.TLSCertFile == "" || s.config.TLSKeyFile == "" {
				errChan <- fmt.Errorf("TLS certificate and key files are required")
				return
			}
			err = srv.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// listen system signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		s.Logger().Errorf(ctx, "HTTP server %s start failed: %v", s.config.ServerName, err)
	case <-quit:
		s.Logger().Infof(ctx, "Shutting down server...")

		timeout := 5 * time.Second
		if s.config.GracefulEnable {
			timeout = s.config.GracefulTimeout
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		if s.config.GracefulEnable {
			// wait for active connections to complete
			time.Sleep(s.config.GracefulWaitTime)
		}

		if err := srv.Shutdown(ctx); err != nil {
			s.Logger().Errorf(ctx, "Server forced to shutdown: %v", err)
		}
	}
}
