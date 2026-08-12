package mlog

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerReconfigurationClosesPreviousFile(t *testing.T) {
	firstPath := filepath.Join(t.TempDir(), "first.log")
	logger := New(&Config{Filepath: firstPath, Stdout: false})
	logger.Infof(context.Background(), "open first file")
	previousWriter, ok := logger.closer.(*fileWriter)
	require.True(t, ok)
	require.NotNil(t, previousWriter.file)

	require.NoError(t, logger.SetConfig(&Config{
		Filepath: filepath.Join(t.TempDir(), "second.log"),
		Stdout:   false,
	}))
	assert.Nil(t, previousWriter.file)
	require.NoError(t, logger.Close())
}

func TestLoggerConcurrentHookMutationAndLogging(_ *testing.T) {
	logger := New(&Config{Writer: discardWriter{}, Stdout: false})
	defer logger.Close()

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			logger.Infof(context.Background(), "concurrent")
		}()
		go func(index int) {
			defer wg.Done()
			hook := &namedNoopHook{name: string(rune('a' + index))}
			_ = logger.AddHook(hook)
			logger.RemoveHook(hook.name)
		}(i)
	}
	wg.Wait()
}

type discardWriter struct{}

func (discardWriter) Write(data []byte) (int, error) { return len(data), nil }

type namedNoopHook struct{ name string }

func (h *namedNoopHook) Name() string    { return h.name }
func (h *namedNoopHook) Levels() []Level { return AllLevels() }
func (h *namedNoopHook) Fire(_ *Entry)   {}
