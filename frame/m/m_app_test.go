package m_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/graingo/maltose/frame/m"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type blockingServer struct {
	stopContext chan context.Context
}

func (s *blockingServer) Start(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (s *blockingServer) Stop(ctx context.Context) error {
	s.stopContext <- ctx
	return nil
}

type failingServer struct{ err error }

func (s *failingServer) Start(context.Context) error { return s.err }
func (s *failingServer) Stop(context.Context) error  { return nil }

func TestAppStopsServersWithDeadlineAfterStartupFailure(t *testing.T) {
	startupErr := errors.New("startup failed")
	server := &blockingServer{stopContext: make(chan context.Context, 1)}
	app := m.NewApp(
		m.WithLogger(nil),
		m.WithShutdownTimeout(time.Second),
		m.WithServer(server, &failingServer{err: startupErr}),
	)

	err := app.Run()
	require.ErrorIs(t, err, startupErr)
	select {
	case stopCtx := <-server.stopContext:
		_, hasDeadline := stopCtx.Deadline()
		assert.True(t, hasDeadline)
	case <-time.After(time.Second):
		t.Fatal("server Stop was not called")
	}
}
