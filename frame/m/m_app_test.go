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

type recordingCloser struct {
	name  string
	order *[]string
}

func (c *recordingCloser) Close() error {
	*c.order = append(*c.order, c.name)
	return nil
}

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

func TestAppRunsHooksAndClosersInReverseOrder(t *testing.T) {
	startupErr := errors.New("startup failed")
	order := make([]string, 0, 3)
	app := m.NewApp(
		m.WithShutdownTimeout(time.Second),
		m.WithShutdownHook(func(context.Context) error {
			order = append(order, "hook")
			return nil
		}),
		m.WithCloser(
			&recordingCloser{name: "first closer", order: &order},
			&recordingCloser{name: "second closer", order: &order},
		),
		m.WithServer(&failingServer{err: startupErr}),
	)

	err := app.Run()
	require.ErrorIs(t, err, startupErr)
	assert.Equal(t, []string{"second closer", "first closer", "hook"}, order)
}

func TestAppGivesEachShutdownHookItsOwnTimeout(t *testing.T) {
	startupErr := errors.New("startup failed")
	const timeout = 40 * time.Millisecond
	secondHookCalled := false
	app := m.NewApp(
		m.WithShutdownTimeout(timeout),
		m.WithShutdownHook(
			func(context.Context) error {
				secondHookCalled = true
				return nil
			},
			func(ctx context.Context) error {
				<-ctx.Done()
				return ctx.Err()
			},
		),
		m.WithServer(&failingServer{err: startupErr}),
	)

	started := time.Now()
	err := app.Run()
	require.ErrorIs(t, err, startupErr)
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.True(t, secondHookCalled)
	assert.Less(t, time.Since(started), 5*timeout)
}
