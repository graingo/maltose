package mhttp

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/graingo/maltose/errors/merror"
)

// SetStaticPath serves files from directory under the supplied URL prefix.
func (s *Server) SetStaticPath(prefix string, directory string) {
	s.engine.StaticFS(prefix, http.Dir(directory))
}

// Handler returns the prepared HTTP handler.
// Route registration must be complete before the first call to Handler, ServeHTTP, Start, or Run.
func (s *Server) Handler() http.Handler {
	s.prepare(context.Background())
	return s
}

// ServeHTTP implements http.Handler and allows Server to be used with httptest.
func (s *Server) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	s.prepare(request.Context())
	s.engine.ServeHTTP(writer, request)
}

// Run starts the HTTP server and waits for either shutdown or a process signal.
func (s *Server) Run() {
	ctx := context.Background()
	errChan := make(chan error, 1)
	go func() {
		errChan <- s.Start(ctx)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err := <-errChan:
		if err != nil {
			s.logger().Errorf(ctx, err, "HTTP server %s start failed", s.config.ServerName)
		}
	case <-quit:
		s.logger().Infof(ctx, "Shutting down server...")
		if err := s.Stop(ctx); err != nil {
			s.logger().Errorf(ctx, err, "HTTP server %s forced to shutdown", s.config.ServerName)
		}
	}
}

// Start starts the server on its configured address and blocks until it stops.
func (s *Server) Start(ctx context.Context) error {
	s.prepare(ctx)
	server := &http.Server{
		Addr:           s.normalizeAddress(),
		Handler:        s,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}
	s.setHTTPServer(server)
	defer s.clearHTTPServer(server)

	var err error
	if s.config.TLSEnable {
		if s.config.TLSCertFile == "" || s.config.TLSKeyFile == "" {
			return merror.New("tls certificate and key files are required")
		}
		err = server.ListenAndServeTLS(s.config.TLSCertFile, s.config.TLSKeyFile)
	} else {
		err = server.ListenAndServe()
	}
	return s.handleServeError(ctx, err)
}

// StartListener serves HTTP on listener and blocks until the server stops.
// It is useful when callers need control over port allocation, including in tests.
func (s *Server) StartListener(ctx context.Context, listener net.Listener) error {
	if listener == nil {
		return merror.New("HTTP listener is required")
	}
	s.prepare(ctx)
	server := &http.Server{
		Handler:        s,
		ReadTimeout:    s.config.ReadTimeout,
		WriteTimeout:   s.config.WriteTimeout,
		IdleTimeout:    s.config.IdleTimeout,
		MaxHeaderBytes: s.config.MaxHeaderBytes,
	}
	s.setHTTPServer(server)
	defer s.clearHTTPServer(server)

	var err error
	if s.config.TLSEnable {
		if s.config.TLSCertFile == "" || s.config.TLSKeyFile == "" {
			return merror.New("tls certificate and key files are required")
		}
		err = server.ServeTLS(listener, s.config.TLSCertFile, s.config.TLSKeyFile)
	} else {
		err = server.Serve(listener)
	}
	return s.handleServeError(ctx, err)
}

// Stop gracefully stops the active HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.logger().Infof(ctx, "HTTP server %s is stopping", s.config.ServerName)
	server := s.currentHTTPServer()
	if server == nil {
		return nil
	}
	if !s.config.GracefulEnable {
		return server.Close()
	}

	shutdownCtx, cancel := gracefulShutdownContext(ctx, s.config.GracefulTimeout)
	defer cancel()
	if err := waitForGracefulShutdown(shutdownCtx, s.config.GracefulWaitTime); err != nil {
		return err
	}
	return server.Shutdown(shutdownCtx)
}

func (s *Server) prepare(ctx context.Context) {
	s.prepareOnce.Do(func() {
		s.registerHealthCheck(ctx)
		s.registerDoc(ctx)
		s.bindRoutes(ctx)
		s.printRoute(ctx)
	})
}

func (s *Server) handleServeError(ctx context.Context, err error) error {
	if err == nil || err == http.ErrServerClosed {
		return nil
	}
	s.logger().Errorf(ctx, err, "HTTP server %s start failed", s.config.ServerName)
	return err
}

func (s *Server) setHTTPServer(server *http.Server) {
	s.serverMu.Lock()
	s.srv = server
	s.serverMu.Unlock()
}

func (s *Server) currentHTTPServer() *http.Server {
	s.serverMu.RLock()
	defer s.serverMu.RUnlock()
	return s.srv
}

func (s *Server) clearHTTPServer(server *http.Server) {
	s.serverMu.Lock()
	if s.srv == server {
		s.srv = nil
	}
	s.serverMu.Unlock()
}

func gracefulShutdownContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(ctx)
	}
	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) <= timeout {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func waitForGracefulShutdown(ctx context.Context, wait time.Duration) error {
	if wait <= 0 {
		return nil
	}
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// normalizeAddress checks and formats the server address.
// If the address only contains a port, it prepends a colon to make it a valid listening address.
func (s *Server) normalizeAddress() string {
	address := s.config.Address
	if address != "" && !strings.Contains(address, ":") {
		return ":" + address
	}
	return address
}
